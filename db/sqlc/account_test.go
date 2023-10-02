package db

import (
	"bank/util"
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestAccount(t *testing.T) (Account) {
	user := createTestUser(t)
	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  util.RandomAmount(),
		Currency: util.RandomCurrency(),
	}
	acc, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotZero(t, acc)

	return acc
}

func TestCreateAccount(t *testing.T) {
	createTestAccount(t)
}

func TestGetAccount(t *testing.T) {
	acc := createTestAccount(t)
	result, err := testQueries.GetAccount(context.Background(), acc.ID)
	require.NoError(t, err)
	require.NotZero(t, acc.ID)
	require.NotEmpty(t, acc)

	require.Equal(t, acc.ID, result.ID)
	require.Equal(t, acc.Owner, result.Owner)
	require.Equal(t, acc.Balance, result.Balance)
	require.Equal(t, acc.Currency, result.Currency)
}

func TestUpdateAccount(t *testing.T) {
	acc := createTestAccount(t)
	params := UpdateAccountParams{
		ID: acc.ID,
		Balance: util.RandomAmount(),
	}
	err := testQueries.UpdateAccount(context.Background(),params)
	require.NoError(t, err)

	result, err :=  testQueries.GetAccount(context.Background(), acc.ID)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	require.Equal(t, acc.ID, result.ID)
	require.Equal(t, acc.Owner, result.Owner)
	require.Equal(t, params.Balance, result.Balance)
	require.Equal(t, acc.Currency, result.Currency)
}

func TestDeleteAccount(t *testing.T) {
	acc := createTestAccount(t)
	err := testQueries.DeleteAccount(context.Background(), acc.ID)
	require.NoError(t, err)

	result, err := testQueries.GetAccount(context.Background(), acc.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, result)
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
