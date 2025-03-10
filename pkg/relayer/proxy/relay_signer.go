package proxy

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/crypto"

	"github.com/pokt-network/poktroll/pkg/signer"
	"github.com/pokt-network/poktroll/x/service/types"
)

// SignRelayResponse is a shared method used by the RelayServers to sign the hash of the RelayResponse.
// It uses the keyring and keyName to sign the payload and returns the signature.
// TODO_TECHDEBT(@red-0ne): This method should be moved out of the RelayerProxy interface
// that should not be responsible for signing relay responses.
// See https://github.com/pokt-network/poktroll/issues/160 for a better design.
func (rp *relayerProxy) SignRelayResponse(relayResponse *types.RelayResponse) error {
	// create a simple signer for the request
	signer := signer.NewSimpleSigner(rp.keyring, rp.signingKeyName)

	// extract and hash the relay response's signable bytes
	signableBz, err := relayResponse.GetSignableBytes()
	if err != nil {
		return sdkerrors.Wrapf(ErrRelayerProxyInvalidRelayResponse, "error getting signable bytes: %v", err)
	}
	hash := crypto.Sha256(signableBz)

	// sign the relay response
	sig, err := signer.Sign(hash)
	if err != nil {
		return sdkerrors.Wrapf(ErrRelayerProxyInvalidRelayResponse, "error signing relay response: %v", err)
	}

	// set the relay response's signature
	relayResponse.Meta.SupplierSignature = sig
	return nil
}
