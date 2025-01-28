package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valkyraycho/bank_project/utils"
)

func TestCreateEntry(t *testing.T) {
	randomEntry(t)
}

func TestGetEntry(t *testing.T) {
	entry := randomEntry(t)

	gotEntry, err := testStore.GetEntry(context.Background(), entry.ID)
	require.NoError(t, err)
	require.NotEmpty(t, gotEntry)
	require.Equal(t, entry.AccountID, gotEntry.AccountID)
	require.Equal(t, entry.Amount, gotEntry.Amount)
	require.Equal(t, entry.CreatedAt, gotEntry.CreatedAt)
}

func TestListEntries(t *testing.T) {

	var entry Entry

	for i := 0; i < 10; i++ {
		entry = randomEntry(t)
	}

	args := ListEntriesParams{
		AccountID: entry.AccountID,
		Limit:     10,
		Offset:    0,
	}

	entries, err := testStore.ListEntries(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
		require.Equal(t, args.AccountID, entry.AccountID)
	}
}

func randomEntry(t *testing.T) Entry {
	account := randomAccount(t)

	args := CreateEntryParams{
		AccountID: account.ID,
		Amount:    utils.RandomMoney(),
	}

	entry, err := testStore.CreateEntry(context.Background(), args)

	require.NoError(t, err)
	require.NotEmpty(t, entry)
	require.Equal(t, args.AccountID, entry.AccountID)
	require.Equal(t, args.Amount, entry.Amount)
	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}
