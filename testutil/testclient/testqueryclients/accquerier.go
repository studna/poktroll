package testqueryclients

import (
	"context"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	accounttypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/mock/gomock"

	"github.com/pokt-network/poktroll/testutil/mockclient"
	"github.com/pokt-network/poktroll/testutil/sample"
)

// NewTestAccountQueryClient creates a mock of the AccountQueryClient
// which allows the caller to call GetApplication any times and will return
// an application with the given address.
// The public key in the account it returns is a randomly generated secp256k1
// public key, not related to the address provided.
func NewTestAccountQueryClient(
	t *testing.T,
	ctx context.Context,
) *mockclient.MockAccountQueryClient {
	ctrl := gomock.NewController(t)

	accoutQuerier := mockclient.NewMockAccountQueryClient(ctrl)
	accoutQuerier.EXPECT().GetAccount(gomock.Eq(ctx), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			address string,
		) (account accounttypes.AccountI, err error) {
			// Generate a random public key
			_, pk := sample.AccAddressAndPubKey()
			anyPk, err := codectypes.NewAnyWithValue(pk)
			if err != nil {
				return nil, err
			}
			return &accounttypes.BaseAccount{
				Address: address,
				PubKey:  anyPk,
			}, nil
		}).
		AnyTimes()

	return accoutQuerier
}
