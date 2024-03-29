package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Provides functions to execute all Quesris and Transactions
type Store interface {
	Querier // Inherit all quering functions generated by SQLC
	TransferTx(ctx context.Context, arg TransferTxParms) (TransferTxResult, error)
}

// Provides functions to execute all Queries and Transations on a SQL database
type SQLStore struct {
	*Queries
	db *sql.DB
}

// Creates a new store instance
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

//  Executes a function within a databasr transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
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
	FromBalance   int64 `json:"from_balance"`
	ToBalance     int64 `json:"to_balance"`
}

func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParms) (TransferTxResult, error) {

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
		// To avoid deadlock between concurrent transactions always
		// add or remove from the account with the smallest ID first

		if (arg.FromAccountID < arg.ToAccountID){
			result.FromAccountID, result.ToAccountID, result.FromBalance, result.ToBalance, err = moveMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
			if err != nil {
				return err
			}
		} else {
			result.ToAccountID, result.FromAccountID, result.ToBalance, result.FromBalance, err = moveMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return result, err
}

func moveMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amt1 int64,
	accountID2 int64,
	amt2 int64,
) (accID1, accID2, bal1, bal2 int64, err error) {
	account1, err := q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID: accountID1,
			Amount: amt1,
	})
	if err != nil {
		return
	}
	account2, err := q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID: accountID2,
		Amount: amt2,
	})
	if err != nil {
		return
	}
	accID1 = account1.ID
	bal1 = account1.Balance
	accID2 = account2.ID
	bal2 = account2.Balance
	return
}
