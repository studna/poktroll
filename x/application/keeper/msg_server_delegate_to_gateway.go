package keeper

import (
	"context"

	"pocket/x/application/types"

	sdkerrors "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) DelegateToGateway(goCtx context.Context, msg *types.MsgDelegateToGateway) (*types.MsgDelegateToGatewayResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	logger := k.Logger(ctx).With("method", "DelegateToGateway")
	logger.Info("About to delegate application to gateway with msg: %v", msg)

	// DISCUSS_IN_THIS_PR: Should an application have to stake prior to delegation?

	// Retrieve the application from the store
	app, found := k.GetApplication(ctx, msg.AppAddress)
	if !found {
		logger.Info("Application not found with address [%s]", msg.AppAddress)
		return nil, types.ErrAppNotFound
	}
	logger.Info("Application found with address [%s]", msg.AppAddress)

	// Check if the application is already delegated to the gateway
	for _, delegateePubKey := range app.DelegateePubKeys {
		// Convert the any type to a public key
		delegateePubKey, err := types.AnyToPubKey(delegateePubKey)
		if err != nil {
			logger.Error("unable to convert any type to public key: %v", err)
			return nil, sdkerrors.Wrapf(types.ErrAppAnyConversion, "unable to convert any type to public key: %v", err)
		}
		// Convert the public key to an address
		delegateeAddress := types.PublicKeyToAddress(delegateePubKey)
		if delegateeAddress == msg.GatewayAddress {
			logger.Info("Application already delegated to gateway with address: %s", msg.GatewayAddress)
			return nil, sdkerrors.Wrapf(types.ErrAppAlreadyDelegated, "application already delegated to gateway with address: %s", msg.GatewayAddress)
		}
	}

	// Retrieve the public key of the gateway
	pubKey, err := k.addressToPubKey(ctx, msg.GatewayAddress)
	if err != nil {
		logger.Error("unable to get public key from address [%s]: %v", msg.GatewayAddress, err)
		return nil, sdkerrors.Wrapf(types.ErrAppInvalidGatewayAddress, "unable to get public key from address %s", msg.GatewayAddress)
	}
	anyPubKey, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		logger.Error("unable to create any type from public key: %v", err)
		return nil, sdkerrors.Wrapf(types.ErrAppAnyConversion, "unable to create any type from public key: %v", err)
	}

	// Update the application with the new delegatee public key
	app.DelegateePubKeys = append(app.DelegateePubKeys, *anyPubKey)
	logger.Info("Successfully added delegatee public key to application")

	// Update the application store with the new delegation
	k.SetApplication(ctx, app)
	logger.Info("Successfully delegated application to gateway for app: %+v", app)

	return &types.MsgDelegateToGatewayResponse{}, nil
}

func (k msgServer) addressToPubKey(ctx sdk.Context, address string) (cryptotypes.PubKey, error) {
	// Retrieve the address of the address
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}
	// Return the public key of the address
	return k.accountKeeper.GetPubKey(ctx, accAddr)
}
