package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/valkyraycho/bank_project/utils"
)

func TestCreateAccount(t *testing.T) {
	randomAccount(t)
}

func TestGetAccount(t *testing.T) {
	account := randomAccount(t)

	gotAccount, err := testStore.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, gotAccount)
	require.Equal(t, account.ID, gotAccount.ID)
	require.Equal(t, account.OwnerID, gotAccount.OwnerID)
	require.Equal(t, account.Balance, gotAccount.Balance)
	require.Equal(t, account.CreatedAt, gotAccount.CreatedAt)
	require.Equal(t, account.Currency, gotAccount.Currency)
}

func TestDeleteAccount(t *testing.T) {
	account := randomAccount(t)

	err := testStore.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, err)

	deletedAccount, err := testStore.GetAccount(context.Background(), account.ID)
	require.Error(t, err)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
	require.Empty(t, deletedAccount)
}

func TestListAccount(t *testing.T) {
	var account Account

	for i := 0; i < 10; i++ {
		account = randomAccount(t)
	}

	args := ListAccountParams{
		OwnerID: account.OwnerID,
		Limit:   10,
		Offset:  0,
	}

	accounts, err := testStore.ListAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.Equal(t, args.OwnerID, account.OwnerID)
	}
}

func TestUpdateAccount(t *testing.T) {
	account := randomAccount(t)

	args := UpdateAccountParams{
		ID:      account.ID,
		Balance: utils.RandomMoney(),
	}

	updatedAccount, err := testStore.UpdateAccount(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount)
	require.Equal(t, args.ID, updatedAccount.ID)
	require.Equal(t, args.Balance, updatedAccount.Balance)
	require.Equal(t, account.Currency, updatedAccount.Currency)
	require.Equal(t, account.CreatedAt, updatedAccount.CreatedAt)
}

func TestAddAccountBalance(t *testing.T) {
	account := randomAccount(t)

	args := AddAccountBalanceParams{
		ID:     account.ID,
		Amount: utils.RandomMoney(),
	}

	updatedAccount, err := testStore.AddAccountBalance(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount)
	require.Equal(t, args.ID, updatedAccount.ID)
	require.Equal(t, account.Balance+args.Amount, updatedAccount.Balance)
	require.Equal(t, account.Currency, updatedAccount.Currency)
	require.Equal(t, account.CreatedAt, updatedAccount.CreatedAt)
}

func randomAccount(t *testing.T) Account {
	user := randomUser(t)

	args := CreateAccountParams{
		OwnerID:  user.ID,
		Currency: utils.RandomCurrency(),
		Balance:  utils.RandomMoney(),
	}
	account, err := testStore.CreateAccount(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.Equal(t, account.OwnerID, args.OwnerID)
	require.Equal(t, account.Currency, args.Currency)
	require.Equal(t, account.Balance, args.Balance)

	require.NotZero(t, account.CreatedAt)
	require.NotZero(t, account.ID)
	return account
}
