package sqlc

import (
	"context"
	"errors"

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
type AccountTransactionResult struct {
	Transaction Transaction
	Account     Account
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		Queries: New(pool),
	}
}

// TxFunc is a function that executes database operations within a transaction
type TxFunc func(q *Queries) error

var (
	ErrInsufficientBalance = errors.New("insufficient balance for withdrawal")
	ErrInvalidAmount       = errors.New("withdrawal amount must be positive")
)

func (store *Store) executeTransaction(ctx context.Context, fn TxFunc) error {
	tx, err := store.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
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

// TransferMoneyTx performs a money transfer between two accounts within a database transaction.
// It creates a transfer record, transaction entries for both accounts, and updates account balances.
// The function uses row-level locking (SELECT FOR UPDATE) to prevent race conditions and ensures
// consistent lock ordering by ID to avoid deadlocks.
func (store *Store) TransferMoneyTx(ctx context.Context, arg CreateTransferParams) (TransferMoneyResult, error) {
	var transferMoneyResult TransferMoneyResult

	if arg.FromAccountID == arg.ToAccountID {
		return transferMoneyResult, errors.New("cannot transfer to the same account")
	}

	err := store.executeTransaction(ctx, func(q *Queries) error {
		var err error

		// Lock both accounts in ascending ID order to prevent deadlocks.
		// When multiple concurrent transfers involve the same accounts in different directions,
		// locking in a consistent order ensures no circular wait conditions occur.
		if arg.FromAccountID < arg.ToAccountID {
			transferMoneyResult.FromAccount, err = q.GetAccountForUpdate(ctx, arg.FromAccountID)
			if err != nil {
				return err
			}
			transferMoneyResult.ToAccount, err = q.GetAccountForUpdate(ctx, arg.ToAccountID)
			if err != nil {
				return err
			}
		} else {
			transferMoneyResult.ToAccount, err = q.GetAccountForUpdate(ctx, arg.ToAccountID)
			if err != nil {
				return err
			}
			transferMoneyResult.FromAccount, err = q.GetAccountForUpdate(ctx, arg.FromAccountID)
			if err != nil {
				return err
			}
		}

		// Create transfer record
		transferMoneyResult.Transfer, err = q.CreateTransfer(ctx, arg)
		if err != nil {
			return err
		}

		// Create transaction entry for sender (debit)
		transferMoneyResult.FromTx, err = q.CreateTransaction(ctx, CreateTransactionParams{
			AccountID:         arg.FromAccountID,
			Type:              TransactionTypeTransferOut,
			AmountCents:       -arg.AmountCents,
			BalanceAfterCents: transferMoneyResult.FromAccount.BalanceCents - arg.AmountCents,
		})
		if err != nil {
			return err
		}

		// Create transaction entry for receiver (credit)
		transferMoneyResult.ToTx, err = q.CreateTransaction(ctx, CreateTransactionParams{
			AccountID:         arg.ToAccountID,
			Type:              TransactionTypeTransferIn,
			AmountCents:       arg.AmountCents,
			BalanceAfterCents: transferMoneyResult.ToAccount.BalanceCents + arg.AmountCents,
		})
		if err != nil {
			return err
		}

		// Update sender account balance
		transferMoneyResult.FromAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
			ID:           arg.FromAccountID,
			BalanceCents: transferMoneyResult.FromAccount.BalanceCents - arg.AmountCents,
		})
		if err != nil {
			return err
		}

		// Update receiver account balance
		transferMoneyResult.ToAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
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

type AccountTransactionParams struct {
	AccountID int64
	Amount    int64
}

func (store *Store) DepositMoneyTx(ctx context.Context, arg AccountTransactionParams) (AccountTransactionResult, error) {
	var depositMoneyResult AccountTransactionResult
	err := store.executeTransaction(ctx, func(q *Queries) error {
		var err error
		depositMoneyResult.Account, err = q.GetAccountForUpdate(ctx, arg.AccountID)
		if err != nil {
			return err
		}
		depositMoneyResult.Transaction, err = q.CreateTransaction(ctx, CreateTransactionParams{
			AccountID:         arg.AccountID,
			Type:              TransactionTypeDeposit,
			AmountCents:       arg.Amount,
			BalanceAfterCents: depositMoneyResult.Account.BalanceCents + arg.Amount,
		})
		if err != nil {
			return err
		}
		depositMoneyResult.Account, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
			ID:           arg.AccountID,
			BalanceCents: depositMoneyResult.Account.BalanceCents + arg.Amount,
		})
		if err != nil {
			return err
		}
		return nil
	})

	return depositMoneyResult, err
}

func (store *Store) WithdrawMoneyTx(ctx context.Context, arg AccountTransactionParams) (AccountTransactionResult, error) {
	var withdrawMoneyResult AccountTransactionResult
	if arg.Amount <= 0 {
		return AccountTransactionResult{}, ErrInvalidAmount
	}

	err := store.executeTransaction(ctx, func(q *Queries) error {
		var err error
		withdrawMoneyResult.Account, err = q.GetAccountForUpdate(ctx, arg.AccountID)
		if err != nil {
			return err
		}

		if withdrawMoneyResult.Account.BalanceCents < arg.Amount {
			return ErrInsufficientBalance
		}

		balanceAfterWithdrawal := withdrawMoneyResult.Account.BalanceCents - arg.Amount
		withdrawMoneyResult.Transaction, err = q.CreateTransaction(ctx, CreateTransactionParams{
			AccountID:         arg.AccountID,
			Type:              TransactionTypeWithdrawal,
			AmountCents:       -arg.Amount,
			BalanceAfterCents: balanceAfterWithdrawal,
		})
		if err != nil {
			return err
		}
		withdrawMoneyResult.Account, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
			ID:           arg.AccountID,
			BalanceCents: balanceAfterWithdrawal,
		})
		if err != nil {
			return err
		}
		return nil
	})

	return withdrawMoneyResult, err
}
