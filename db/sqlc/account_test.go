package db

import (
	"bank/util"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestAccount(t *testing.T) (int64, CreateAccountParams) {
	arg := CreateAccountParams{
		Owner: util.RandomOwner(),
		Balance: util.RandomAmount(),
		Currency: util.RandomCurrency(),
	}
	accID, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotZero(t, accID)

	return accID, arg
}

func TestCreateAccount(t *testing.T) {
	createTestAccount(t)
}

func TestGetAccount(t *testing.T) {
	accID, arg := createTestAccount(t)
	acc, err := testQueries.GetAccount(context.Background(), accID)
	require.NoError(t, err)
	require.NotZero(t, accID)
	require.NotEmpty(t, acc)

	require.Equal(t, accID, acc.ID)
	require.Equal(t, arg.Owner, acc.Owner)
	require.Equal(t, arg.Balance, acc.Balance)
	require.Equal(t, arg.Currency, acc.Currency)
}
