package db

import (
	"bank/util"
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestAccount(t *testing.T) (int64, CreateAccountParams) {
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomAmount(),
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

func TestUpdateAccount(t *testing.T) {
	accID, arg := createTestAccount(t)
	params := UpdateAccountParams{
		ID: accID,
		Balance: util.RandomAmount(),
	}
	err := testQueries.UpdateAccount(context.Background(),params)
	require.NoError(t, err)

	acc, err :=  testQueries.GetAccount(context.Background(), accID)
	require.NoError(t, err)
	require.NotEmpty(t, acc)

	require.Equal(t, accID, acc.ID)
	require.Equal(t, arg.Owner, acc.Owner)
	require.Equal(t, params.Balance, acc.Balance)
	require.Equal(t, arg.Currency, acc.Currency)
}

func TestDeleteAccount(t *testing.T) {
	accID, _ := createTestAccount(t)
	err := testQueries.DeleteAccount(context.Background(), accID)
	require.NoError(t, err)

	acc, err := testQueries.GetAccount(context.Background(), accID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, acc)
}

func TestListAccounts(t *testing.T) {
	for i:=0; i<10; i++ {
		createTestAccount(t)
	}

	arg := ListAccountsParams{
		Limit: 5,
		Offset: 5,
	}

	acc, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, 5, len(acc))
}
