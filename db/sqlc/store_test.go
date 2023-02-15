package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	acc1_id, _ := createTestAccount(t)
	acc2_id, _ := createTestAccount(t)

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
			ID: result.FromEntryID,
		})
		require.NoError(t, err)
		require.NotZero(t, entryFrom.ID)
		require.NotZero(t, entryFrom.CreatedAt)
		require.Equal(t, -amount, entryFrom.Amount)

		entryTo, err := store.GetEntry(context.Background(), GetEntryParams{
			AccountID: acc2_id,
			ID: result.ToEntryID,
		})
		require.NoError(t, err)
		require.NotZero(t, entryFrom.ID)
		require.NotZero(t, entryFrom.CreatedAt)
		require.Equal(t, amount, entryTo.Amount)

		// TODO: Check account balnace
	}
}
