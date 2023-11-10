package miner

import (
	"context"
	"crypto/sha256"
	"hash"

	"cosmossdk.io/depinject"

	"github.com/pokt-network/poktroll/pkg/client"
	"github.com/pokt-network/poktroll/pkg/either"
	"github.com/pokt-network/poktroll/pkg/observable"
	"github.com/pokt-network/poktroll/pkg/observable/channel"
	"github.com/pokt-network/poktroll/pkg/observable/filter"
	"github.com/pokt-network/poktroll/pkg/observable/logging"
	"github.com/pokt-network/poktroll/pkg/relayer"
	"github.com/pokt-network/poktroll/pkg/relayer/protocol"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
)

var (
	_                        relayer.Miner = (*miner)(nil)
	defaultHasherConstructor               = sha256.New
	// TODO_BLOCKER: query on-chain governance params once available.
	// Setting this to 0 to effectively disables mining for now.
	// I.e., all relays are added to the tree.
	defaultRelayDifficulty = 0
)

// Miner is responsible for observing servedRelayObs, hashing and checking the
// difficulty of each, finally publishing those with sufficient difficulty to
// minedRelayObs as they are applicable for relay volume.
type miner struct {
	relayHasher func() hash.Hash
	relayDifficulty   int

	// Injected dependencies
	sessionManager relayer.RelayerSessionsManager
	blockClient    client.BlockClient
}

// NewMiner creates a new miner from the given dependencies and options. It
// returns an error if it has not been sufficiently configured or supplied.
func NewMiner(
	deps depinject.Config,
	opts ...relayer.MinerOption,
) (*miner, error) {
	mnr := &miner{}

	if err := depinject.Inject(
		deps,
		&mnr.sessionManager,
		&mnr.blockClient,
	); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(mnr)
	}

	mnr.setDefaults()

	return mnr, nil
}

// MinedRelays maps servedRelaysObs through a pipeline which:
// 1. Hashes the relay
// 2. Checks if it's above the mining difficulty
// 3. Adds it to the session tree if so
// It DOES NOT BLOCK as map operations run in their own goroutines.
func (mnr *miner) MinedRelays(
	ctx context.Context,
	servedRelaysObs observable.Observable[*servicetypes.Relay],
) observable.Observable[*relayer.MinedRelay] {
	// Map servedRelaysObs to a new observable of an either type, populated with
	// the minedRelay or an error. It is notified after the relay has been mined
	// or an error has been encountered, respectively.
	eitherMinedRelaysObs := channel.Map(ctx, servedRelaysObs, mnr.mapMineRelay)
	logging.LogErrors(ctx, filter.EitherError(ctx, eitherMinedRelaysObs))

	return filter.EitherSuccess(ctx, eitherMinedRelaysObs)
}

// setDefaults ensures that the miner has been configured with a hasherConstructor and uses
// the default hasherConstructor if not.
func (mnr *miner) setDefaults() {
	if mnr.hasherConstructor == nil {
		mnr.hasherConstructor = defaultHasherConstructor
	}
}

// mapMineRelay is intended to be used as a MapFn.
// 1. It hashes the relay and compares its difficult to the minimum threshold.
// 2. If the relay difficulty is sufficient -> return an Either[MineRelay Value]
// 3. If an error is encountered -> return an Either[error]
// 4. Otherwise, skip the relay.
func (mnr *miner) mapMineRelay(
	_ context.Context,
	relay *servicetypes.Relay,
) (_ either.Either[*relayer.MinedRelay], skip bool) {
	relayBz, err := relay.Marshal()
	if err != nil {
		return either.Error[*relayer.MinedRelay](err), false
	}

	// TODO_BLOCKER: Centralize the logic of hashing a relay. It should live
	// alongside signing & verification.
	//
	// TODO_IMPROVE: We need to hash the key; it would be nice if smst.Update() could do it
	// since smst has a reference to the hasherConstructor
	relayHash := mnr.hash(relayBz)

	// The relay IS NOT volume / reward applicable
	if !protocol.BytesDifficultyGreaterThan(relayHash, defaultRelayDifficulty) {
		return either.Success[*relayer.MinedRelay](nil), true
	}

	// The relay IS volume / reward applicable
	return either.Success(&relayer.MinedRelay{
		Relay: *relay,
		Bytes: relayBz,
		Hash:  relayHash,
	}), false
}

// hash constructs a new hasher and hashes the given input bytes.
func (mnr *miner) hash(inputBz []byte) []byte {
	hasher := mnr.hasherConstructor()
	hasher.Write(inputBz)
	return hasher.Sum(nil)
}
