package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"cosmossdk.io/depinject"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	cosmosflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/pokt-network/poktroll/cmd/signals"
	"github.com/pokt-network/poktroll/pkg/appgateserver"
	"github.com/pokt-network/poktroll/pkg/deps/config"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
)

// We're `explicitly omitting default` so that the appgateserver crashes if these aren't specified.
const omittedDefaultFlagValue = "explicitly omitting default"

var (
	flagSigningKey     string
	flagSelfSigning    bool
	flagListenEndpoint string
	flagQueryNodeUrl   string
)

func AppGateServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "appgate-server",
		Short: "Starts the AppGate server",
		Long: `Starts the AppGate server that listens for incoming relay requests and handles
the necessary on-chain interactions (sessions, suppliers, etc) to receive the
respective relay response.

-- App Mode (Flag)- -
If the server is started with a defined '--self-signing' flag, it will behave
as an Application. Any incoming requests will be signed by using the private
key and ring associated with the '--signing-key' flag.

-- Gateway Mode (Flag)--
If the '--self-signing' flag is not provided, the server will behave as a Gateway.
It will sign relays on behalf of any Application sending it relays, provided
that the address associated with '--signing-key' has been delegated to. This is
necessary for the application<->gateway ring signature to function.

-- App Mode (HTTP) --
If an application doesn't provide the '--self-signing' flag, it can still send
relays to the AppGate server and function as an Application, provided that:
1. Each request contains the '?senderAddress=[address]' query parameter
2. The key associated with the '--signing-key' flag belongs to the address
   provided in the request, otherwise the ring signature will not be valid.`,
		Args: cobra.NoArgs,
		RunE: runAppGateServer,
	}

	// Custom flags
	cmd.Flags().StringVar(&flagSigningKey, "signing-key", "", "The name of the key that will be used to sign relays")
	cmd.Flags().StringVar(&flagListenEndpoint, "listening-endpoint", "http://localhost:42069", "The host and port that the appgate server will listen on")
	cmd.Flags().BoolVar(&flagSelfSigning, "self-signing", false, "Whether the server should sign all incoming requests with its own ring (for applications)")
	cmd.Flags().StringVar(&flagQueryNodeUrl, "query-node", omittedDefaultFlagValue, "tcp://<host>:<port> to a full pocket node for reading data and listening for on-chain events")

	// Cosmos flags
	cmd.Flags().String(cosmosflags.FlagKeyringBackend, "", "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(cosmosflags.FlagNode, omittedDefaultFlagValue, "registering the default cosmos node flag; needed to initialize the cosmostx and query contexts correctly and uses flagQueryNodeUrl underneath")

	return cmd
}

func runAppGateServer(cmd *cobra.Command, _ []string) error {
	// Create a context that is canceled when the command is interrupted
	ctx, cancelCtx := context.WithCancel(cmd.Context())
	defer cancelCtx()

	// Construct a logger and associate it with the command context.
	logger := polyzero.NewLogger()
	ctx = polylog.WithContext(ctx, logger)
	cmd.SetContext(ctx)

	// Handle interrupt and kill signals asynchronously.
	signals.GoOnExitSignal(cancelCtx)

	// Parse the listening endpoint.
	listenUrl, err := url.Parse(flagListenEndpoint)
	if err != nil {
		return fmt.Errorf("failed to parse listening endpoint: %w", err)
	}

	// Setup the AppGate server dependencies.
	appGateServerDeps, err := setupAppGateServerDependencies(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to setup AppGate server dependencies: %w", err)
	}

	logger.Info().Msg("Creating AppGate server...")

	// Create the AppGate server.
	appGateServer, err := appgateserver.NewAppGateServer(
		appGateServerDeps,
		appgateserver.WithSigningInformation(&appgateserver.SigningInformation{
			// provide the name of the key to use for signing all incoming requests
			SigningKeyName: flagSigningKey,
			// provide whether the appgate server should sign all incoming requests
			// with its own ring (for applications) or not (for gateways)
			SelfSigning: flagSelfSigning,
		}),
		appgateserver.WithListeningUrl(listenUrl),
	)
	if err != nil {
		return fmt.Errorf("failed to create AppGate server: %w", err)
	}

	logger.Info().
		Str("listen_url", listenUrl.String()).
		Msg("Starting AppGate server...")

	// Start the AppGate server.
	if err := appGateServer.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start app gate server: %w", err)
	} else if errors.Is(err, http.ErrServerClosed) {
		logger.Info().Msg("AppGate server stopped")
	}

	return nil
}

func setupAppGateServerDependencies(
	ctx context.Context,
	cmd *cobra.Command,
) (depinject.Config, error) {
	pocketNodeWebsocketUrl, err := getPocketNodeWebsocketUrl()
	if err != nil {
		return nil, err
	}

	supplierFuncs := []config.SupplierFn{
		config.NewSupplyEventsQueryClientFn(pocketNodeWebsocketUrl),
		config.NewSupplyBlockClientFn(pocketNodeWebsocketUrl),
		newSupplyQueryClientContextFn(flagQueryNodeUrl),
	}

	return config.SupplyConfig(ctx, cmd, supplierFuncs)
}

// getPocketNodeWebsocketUrl returns the websocket URL of the Pocket Node to
// connect to for subscribing to on-chain events.
func getPocketNodeWebsocketUrl() (string, error) {
	if flagQueryNodeUrl == omittedDefaultFlagValue {
		return "", errors.New("missing required flag: --query-node")
	}

	pocketNodeURL, err := url.Parse(flagQueryNodeUrl)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ws://%s/websocket", pocketNodeURL.Host), nil
}

// newSupplyQueryClientContextFn returns a new depinject.Config which is supplied with
// the given deps and a new cosmos ClientCtx
func newSupplyQueryClientContextFn(pocketQueryClientUrl string) config.SupplierFn {
	return func(
		_ context.Context,
		deps depinject.Config,
		cmd *cobra.Command,
	) (depinject.Config, error) {
		// Set --node flag to the pocketQueryClientUrl for the client context
		// This flag is read by cosmosclient.GetClientQueryContext.
		err := cmd.Flags().Set(cosmosflags.FlagNode, pocketQueryClientUrl)
		if err != nil {
			return nil, err
		}

		// Get the client context from the command.
		queryClientCtx, err := cosmosclient.GetClientQueryContext(cmd)
		if err != nil {
			return nil, err
		}
		deps = depinject.Configs(deps, depinject.Supply(
			queryClientCtx,
		))
		return deps, nil
	}
}
