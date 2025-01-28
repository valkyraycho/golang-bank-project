package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valkyraycho/bank_project/utils"
)

func TestCreateTransfer(t *testing.T) {
	randomTransfer(t, randomAccount(t), randomAccount(t))
}

func TestGetTransfer(t *testing.T) {
	transfer := randomTransfer(t, randomAccount(t), randomAccount(t))

	gotTransfer, err := testStore.GetTransfer(context.Background(), transfer.ID)

	require.NoError(t, err)
	require.NotEmpty(t, gotTransfer)
	require.Equal(t, transfer.ID, gotTransfer.ID)
	require.Equal(t, transfer.FromAccountID, gotTransfer.FromAccountID)
	require.Equal(t, transfer.ToAccountID, gotTransfer.ToAccountID)
	require.Equal(t, transfer.Amount, gotTransfer.Amount)
	require.Equal(t, transfer.CreatedAt, gotTransfer.CreatedAt)
}

func TestListTransfers(t *testing.T) {
	account1 := randomAccount(t)
	account2 := randomAccount(t)

	for i := 0; i < 5; i++ {
		randomTransfer(t, account1, account2)
		randomTransfer(t, account2, account1)
	}

	args := ListTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account1.ID,
		Limit:         5,
		Offset:        0,
	}

	transfers, err := testStore.ListTransfer(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, transfers)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
		require.True(t, args.FromAccountID == account1.ID || args.ToAccountID == account1.ID)
	}
}

func randomTransfer(t *testing.T, account1, account2 Account) Transfer {

	args := CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        utils.RandomMoney(),
	}

	transfer, err := testStore.CreateTransfer(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, transfer)
	require.Equal(t, args.FromAccountID, transfer.FromAccountID)
	require.Equal(t, args.ToAccountID, transfer.ToAccountID)
	require.Equal(t, args.Amount, transfer.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}
