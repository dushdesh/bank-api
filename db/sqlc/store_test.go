package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	acc1_id, args1 := createTestAccount(t)
	acc2_id, args2 := createTestAccount(t)

	fmt.Println(">> before:", args1.Balance, args2.Balance)

	// run concurrent routines to test the transaction
	n := 5
	amount := int64(10)
	errs := make(chan error, 10)
	results := make(chan TransferTxResult, 10)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParms{
				FromAccountID: acc1_id,
				ToAccountID:   acc2_id,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}

	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)
		require.NotEmpty(t, result.TransferID)
		require.NotEmpty(t, result.FromEntryID)
		require.NotEmpty(t, result.ToEntryID)

		transfer, err := store.GetTransfer(context.Background(), result.TransferID)
		require.NoError(t, err)
		require.Equal(t, acc1_id, transfer.FromAccountID)
		require.Equal(t, acc2_id, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)

		entryFrom, err := store.GetEntry(context.Background(), GetEntryParams{
			AccountID: acc1_id,
			ID:        result.FromEntryID,
		})
		require.NoError(t, err)
		require.NotZero(t, entryFrom.ID)
		require.NotZero(t, entryFrom.CreatedAt)
		require.Equal(t, -amount, entryFrom.Amount)

		entryTo, err := store.GetEntry(context.Background(), GetEntryParams{
			AccountID: acc2_id,
			ID:        result.ToEntryID,
		})
		require.NoError(t, err)
		require.NotZero(t, entryFrom.ID)
		require.NotZero(t, entryFrom.CreatedAt)
		require.Equal(t, amount, entryTo.Amount)

		// TODO: Check account balnace
		fmt.Println(">>tx:", result.FromBalance, result.ToBalance)

		diff1 := args1.Balance - result.FromBalance
		diff2 := result.ToBalance - args2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) // for every transaction the amount will be reduced by amount, 2*amount, 3*amount ..

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}
	close(errs)
	close(results)

	// Check final account balance
	acc1, err := testQueries.GetAccount(context.Background(), acc1_id)
	require.NoError(t, err)
	acc2, err := testQueries.GetAccount(context.Background(), acc2_id)
	require.NoError(t, err)

	fmt.Println(">>after:", acc1.Balance, acc2.Balance)
	require.Equal(t, args1.Balance-int64(n)*amount, acc1.Balance)
	require.Equal(t, args2.Balance+int64(n)*amount, acc2.Balance)
}

func TestDeadlockTransferTx(t *testing.T) {
	store := NewStore(testDB)

	acc1_id, args1 := createTestAccount(t)
	acc2_id, args2 := createTestAccount(t)

	fmt.Println(">> before:", args1.Balance, args2.Balance)

	// run concurrent routines to test the transaction
	n := 10
	amount := int64(10)

	errs := make(chan (error), 10)

	for i := 0; i < n; i++ {

		fromAccountID := acc1_id
		toAccountID := acc2_id

		if i%2 == 1 {
			fromAccountID = acc2_id
			toAccountID = acc1_id
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParms{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})
			errs <- err
		}()
	}

	// Check final account balance
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}
	close(errs)

	updatedAcc1, err := testQueries.GetAccount(context.Background(), acc1_id)
	require.NoError(t, err)
	updatedAcc2, err := testQueries.GetAccount(context.Background(), acc2_id)
	require.NoError(t, err)

	fmt.Println(">>after:", updatedAcc1.Balance, updatedAcc2.Balance)
	require.Equal(t, args1.Balance, updatedAcc1.Balance)
	require.Equal(t, args2.Balance, updatedAcc2.Balance)
}
