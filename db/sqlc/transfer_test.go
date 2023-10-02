package db

import (
	"bank/util"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTransferBetweenAcc(t *testing.T, fromAcc, toAcc Account) (int64, CreateTransferParams){
	arg := CreateTransferParams{
		FromAccountID: fromAcc.ID,
		ToAccountID:   toAcc.ID,
		Amount:        util.RandomAmount(),
	}
	traId, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotZero(t, traId)

	return traId, arg
}

func createTestTransfer(t *testing.T) (int64, CreateTransferParams) {
	fromAcc := createTestAccount(t)
	toAcc := createTestAccount(t)
	return createTransferBetweenAcc(t, fromAcc, toAcc)
}

func TestCreateTransfer(t *testing.T) {
	createTestTransfer(t)
}

func TestGetTransfer(t *testing.T) {
	traId, arg := createTestTransfer(t)

	transfer, err := testQueries.GetTransfer(context.Background(), traId)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)
	require.Equal(t, transfer.ID, traId)
	require.Equal(t, transfer.Amount, arg.Amount)
	require.Equal(t, transfer.FromAccountID, arg.FromAccountID)
	require.Equal(t, transfer.ToAccountID, arg.ToAccountID)
}


func TestListTransfersBetAccounts(t *testing.T) {
	fromAcc := createTestAccount(t)
	toAcc := createTestAccount(t)

	for i:=0; i<10; i++ {
		createTransferBetweenAcc(t, fromAcc, toAcc)
	}

	params := ListTransfersBetAccountsParams{
		ToAccountID: toAcc.ID,
		FromAccountID: fromAcc.ID,
		Limit: 5,
		Offset: 5,
	}

	transfers, err := testQueries.ListTransfersBetAccounts(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, 5, len(transfers))
}

func TestListTransfersFromAccount(t *testing.T) {
	fromAcc := createTestAccount(t)

	for i:=0; i<10; i++ {
		toAcc := createTestAccount(t)
		createTransferBetweenAcc(t, fromAcc, toAcc)
	}

	params := ListTransfersFromAccountParams{
		FromAccountID: fromAcc.ID,
		Limit: 5,
		Offset: 5,
	}

	transfers, err := testQueries.ListTransfersFromAccount(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, 5, len(transfers))
}


func TestListTransfersToAccount(t *testing.T) {
	toAcc := createTestAccount(t)

	for i:=0; i<10; i++ {
		fromAcc := createTestAccount(t)
		createTransferBetweenAcc(t, fromAcc, toAcc)
	}

	params := ListTransfersToAccountParams{
		ToAccountID: toAcc.ID,
		Limit: 5,
		Offset: 5,
	}

	transfers, err := testQueries.ListTransfersToAccount(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, 5, len(transfers))
}
