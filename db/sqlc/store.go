package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("TxErr: %v and RbErr: %v", err, rbErr)
		}
	}
	return tx.Commit()
}

type TransferTxParms struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	TransferID    int64 `json:"transfer_id"`
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	FromEntryID   int64 `json:"from_entry_id"`
	ToEntryID     int64 `json:"to_entry_id"`
	FromBalance   int64 `json:"from_balance`
	ToBalance     int64 `json:"to_balance`
}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParms) (TransferTxResult, error) {

	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.TransferID, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		result.FromEntryID, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntryID, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// update account balances
		fmt.Println("Remove from account", arg.Amount)
		addFromAccResult, err := q.AddAccountBalance(ctx, AddAccountBalanceParams{
			Amount: -arg.Amount,
			ID:     arg.FromAccountID,
		})
		if err != nil {
			return err
		}

		fmt.Println("Add to account", arg.Amount)
		addToAccResult, err := q.AddAccountBalance(ctx, AddAccountBalanceParams{
			Amount: arg.Amount,
			ID:     arg.ToAccountID,
		})
		if err != nil {
			return err
		}
		result.FromAccountID = addFromAccResult.ID
		result.FromBalance = addFromAccResult.Balance
		result.ToAccountID = addToAccResult.ID
		result.ToBalance = addToAccResult.Balance
		
		return nil
	})

	return result, err
}
