package db

import (
	"bank/util"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestEntryForAccount(t *testing.T, accId int64) (int64, CreateEntryParams) {
	arg := CreateEntryParams{
		AccountID: accId,
		Amount: util.RandomSignedAmount(),
	}
	entId, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotZero(t, entId)
	return entId, arg
}

func createTestEntry(t *testing.T) (int64, CreateEntryParams) {
	accId, _ := createTestAccount(t)
	return createTestEntryForAccount(t,accId)
}

func TestCreateEntry(t *testing.T) {
	createTestEntry(t)
}

func TestGetEntry(t *testing.T) {
	entId, arg := createTestEntry(t)
	params := GetEntryParams{
		ID: entId,
		AccountID: arg.AccountID,
	}

	entry, err := testQueries.GetEntry(context.Background(), params)
	require.NoError(t, err)
	require.NotZero(t, entry.ID)
	require.NotEmpty(t, entry)
	require.Equal(t, entId, entry.ID)
	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)
	require.NotEmpty(t, entry.CreatedAt)
}

func TestListEntries(t *testing.T) {
	accId, _ := createTestAccount(t)

	for i:=0; i<10; i++ {
		createTestEntryForAccount(t, accId)
	}

	arg := ListEntriesParams{
		AccountID: accId,
		Limit: 5,
		Offset: 5,
	}

	acc, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, 5, len(acc))
}
