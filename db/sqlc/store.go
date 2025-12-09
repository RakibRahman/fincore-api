package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// we'll expand Quesries funcality by embedding to store. store will provides necessay functions to execute db queries and transactions.
type Store struct {
	*Queries // all indivdual Queries functions will be available to Store
	pool     *pgxpool.Pool
}

type TransferMoneyResult struct {
	Transfer    Transfer
	FromAccount Account
	ToAccount   Account
	FromTx      Transaction
	ToTx        Transaction
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		Queries: New(pool),
	}
}

// TxFunc is a function that executes database operations within a transaction
type TxFunc func(q *Queries) error

func (store *Store) executeTransaction(ctx context.Context, fn TxFunc) error {
	tx, err := store.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := New(tx)
	err = fn(q)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (store *Store) TransferMoneyTx(ctx context.Context, arg CreateTransferParams) (TransferMoneyResult, error) {
	var transferMoneyResult TransferMoneyResult

	err := store.executeTransaction(ctx, func(q *Queries) error {
		var err error

		// 1. Create the transfer record
		transferMoneyResult.Transfer, err = q.CreateTransfer(ctx, arg)
		if err != nil {
			return err
		}

		// 2. Get sender account
		transferMoneyResult.FromAccount, err = q.GetAccount(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		// 3. Create transaction entry for sender
		transferMoneyResult.FromTx, err = q.CreateTransaction(ctx, CreateTransactionParams{
			AccountID:         arg.FromAccountID,
			Type:              TransactionTypeTransferOut,
			AmountCents:       -arg.AmountCents,
			BalanceAfterCents: transferMoneyResult.FromAccount.BalanceCents - arg.AmountCents,
		})
		if err != nil {
			return err
		}

		// 4. Update sender balance
		transferMoneyResult.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:           arg.FromAccountID,
			BalanceCents: transferMoneyResult.FromAccount.BalanceCents - arg.AmountCents,
		})
		if err != nil {
			return err
		}

		// 5. Get receiver account
		transferMoneyResult.ToAccount, err = q.GetAccount(ctx, arg.ToAccountID)
		if err != nil {
			return err
		}

		// 6. Create transaction entry for receiver
		transferMoneyResult.ToTx, err = q.CreateTransaction(ctx, CreateTransactionParams{
			AccountID:         arg.ToAccountID,
			Type:              TransactionTypeTransferIn,
			AmountCents:       arg.AmountCents,
			BalanceAfterCents: transferMoneyResult.ToAccount.BalanceCents + arg.AmountCents,
		})
		if err != nil {
			return err
		}

		// 7. Update receiver balance
		transferMoneyResult.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:           arg.ToAccountID,
			BalanceCents: transferMoneyResult.ToAccount.BalanceCents + arg.AmountCents,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return transferMoneyResult, err
}
