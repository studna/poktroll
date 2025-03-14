package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"cosmossdk.io/depinject"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	cosmosflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/pokt-network/poktroll/cmd/signals"
	"github.com/pokt-network/poktroll/pkg/appgateserver"
	appgateconfig "github.com/pokt-network/poktroll/pkg/appgateserver/config"
	"github.com/pokt-network/poktroll/pkg/deps/config"
)

// We're `explicitly omitting default` so that the appgateserver crashes if these aren't specified.
const omittedDefaultFlagValue = "explicitly omitting default"

var (
	flagAppGateConfig string
)

func AppGateServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "appgate-server",
		Short: "Starts the AppGate server",
		Long: `Starts the AppGate server that listens for incoming relay requests and handles
the necessary on-chain interactions (sessions, suppliers, etc) to receive the
respective relay response.

-- App Mode --
If the server is started with a defined 'self-signing' configuration directive,
it will behave as an Application. Any incoming requests will be signed by using
the private key and ring associated with the 'signing_key' configuration directive.

-- Gateway Mode --
If the 'self_signing' configuration directive is not provided, the server will
behave as a Gateway.
It will sign relays on behalf of any Application sending it relays, provided
that the address associated with 'signing_key' has been delegated to. This is
necessary for the application<->gateway ring signature to function.

-- App Mode (HTTP) --
If an application doesn't provide the 'self_signing' configuration directive,
it can still send relays to the AppGate server and function as an Application,
provided that:
1. Each request contains the '?senderAddress=[address]' query parameter
2. The key associated with the 'signing_key' configuration directive belongs
   to the address provided in the request, otherwise the ring signature will not be valid.`,
		Args: cobra.NoArgs,
		RunE: runAppGateServer,
	}

	// Custom flags
	cmd.Flags().StringVar(&flagAppGateConfig, "config", "", "The path to the appgate config file")

	// Cosmos flags
	cmd.Flags().String(cosmosflags.FlagKeyringBackend, "", "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(cosmosflags.FlagNode, omittedDefaultFlagValue, "registering the default cosmos node flag; needed to initialize the cosmostx and query contexts correctly and uses flagQueryNodeUrl underneath")

	return cmd
}

func runAppGateServer(cmd *cobra.Command, _ []string) error {
	// Create a context that is canceled when the command is interrupted
	ctx, cancelCtx := context.WithCancel(cmd.Context())
	defer cancelCtx()

	// Handle interrupt and kill signals asynchronously.
	signals.GoOnExitSignal(cancelCtx)

	configContent, err := os.ReadFile(flagAppGateConfig)
	if err != nil {
		return err
	}

	appGateConfigs, err := appgateconfig.ParseAppGateServerConfigs(configContent)
	if err != nil {
		return err
	}

	// Setup the AppGate server dependencies.
	appGateServerDeps, err := setupAppGateServerDependencies(ctx, cmd, appGateConfigs)
	if err != nil {
		return fmt.Errorf("failed to setup AppGate server dependencies: %w", err)
	}

	log.Println("INFO: Creating AppGate server...")

	// Create the AppGate server.
	appGateServer, err := appgateserver.NewAppGateServer(
		appGateServerDeps,
		appgateserver.WithSigningInformation(&appgateserver.SigningInformation{
			// provide the name of the key to use for signing all incoming requests
			SigningKeyName: appGateConfigs.SigningKey,
			// provide whether the appgate server should sign all incoming requests
			// with its own ring (for applications) or not (for gateways)
			SelfSigning: appGateConfigs.SelfSigning,
		}),
		appgateserver.WithListeningUrl(appGateConfigs.ListeningEndpoint),
	)
	if err != nil {
		return fmt.Errorf("failed to create AppGate server: %w", err)
	}

	log.Printf("INFO: Starting AppGate server, listening on %s...", appGateConfigs.ListeningEndpoint.String())

	// Start the AppGate server.
	if err := appGateServer.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start app gate server: %w", err)
	} else if errors.Is(err, http.ErrServerClosed) {
		log.Println("INFO: AppGate server stopped")
	}

	return nil
}

func setupAppGateServerDependencies(
	ctx context.Context,
	cmd *cobra.Command,
	appGateConfig *appgateconfig.AppGateServerConfig,
) (depinject.Config, error) {
	pocketNodeWebsocketUrl := fmt.Sprintf("ws://%s/websocket", appGateConfig.QueryNodeUrl.Host)

	supplierFuncs := []config.SupplierFn{
		config.NewSupplyEventsQueryClientFn(pocketNodeWebsocketUrl),
		config.NewSupplyBlockClientFn(pocketNodeWebsocketUrl),
		newSupplyQueryClientContextFn(appGateConfig.QueryNodeUrl.String()),
	}

	return config.SupplyConfig(ctx, cmd, supplierFuncs)
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
