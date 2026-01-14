package sqlc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferMoneyTx(t *testing.T) {
	store := NewStore(testDB)

	fromAccount := createRandomAccountWithQueries(t, store.Queries)
	toAccount := createRandomAccountWithQueries(t, store.Queries)
	amount := int64(10)

	result, err := store.TransferMoneyTx(context.Background(), CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		AmountCents:   amount,
	})

	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Check transfer
	transfer := result.Transfer
	require.NotEmpty(t, transfer)
	require.Equal(t, fromAccount.ID, transfer.FromAccountID)
	require.Equal(t, toAccount.ID, transfer.ToAccountID)
	require.Equal(t, amount, transfer.AmountCents)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	_, err = store.GetTransfer(context.Background(), transfer.ID)
	require.NoError(t, err)

	// Check transaction (fromAccount)
	fromTranx := result.FromTx
	require.NotEmpty(t, fromTranx)
	require.Equal(t, fromAccount.ID, fromTranx.AccountID)
	require.Equal(t, -amount, fromTranx.AmountCents)
	require.NotZero(t, fromTranx.ID)
	require.NotZero(t, fromTranx.CreatedAt)

	_, err = store.GetTransaction(context.Background(), fromTranx.ID)
	require.NoError(t, err)

	// Check transaction (toAccount)
	toTranx := result.ToTx
	require.NotEmpty(t, toTranx)
	require.Equal(t, toAccount.ID, toTranx.AccountID)
	require.Equal(t, amount, toTranx.AmountCents)
	require.NotZero(t, toTranx.ID)
	require.NotZero(t, toTranx.CreatedAt)

	_, err = store.GetTransaction(context.Background(), toTranx.ID)
	require.NoError(t, err)

	// Check account balances
	fromAccountUpdated := result.FromAccount
	require.NotEmpty(t, fromAccountUpdated)
	require.Equal(t, fromAccount.BalanceCents-amount, fromAccountUpdated.BalanceCents)

	toAccountUpdated := result.ToAccount
	require.NotEmpty(t, toAccountUpdated)
	require.Equal(t, toAccount.BalanceCents+amount, toAccountUpdated.BalanceCents)

	// Make sure balance persisted in DB
	fromAccountPersisted, err := store.GetAccount(context.Background(), fromAccount.ID)
	require.NoError(t, err)
	require.Equal(t, fromAccountUpdated.BalanceCents, fromAccountPersisted.BalanceCents)

	toAccountPersisted, err := store.GetAccount(context.Background(), toAccount.ID)
	require.NoError(t, err)
	require.Equal(t, toAccountUpdated.BalanceCents, toAccountPersisted.BalanceCents)
}

func TestTransferMoneyTx_Bidirectional(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccountWithQueries(t, store.Queries)
	account2 := createRandomAccountWithQueries(t, store.Queries)
	amount := int64(5)

	n := 10

	errs := make(chan error, n*2)
	results := make(chan TransferMoneyResult, n*2)

	for i := 0; i < n; i++ {
		// Account1 -> Account2
		go func() {
			result, err := store.TransferMoneyTx(context.Background(), CreateTransferParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				AmountCents:   amount,
			})
			errs <- err
			results <- result
		}()

		// Account2 -> Account1 (reverse direction)
		go func() {
			result, err := store.TransferMoneyTx(context.Background(), CreateTransferParams{
				FromAccountID: account2.ID,
				ToAccountID:   account1.ID,
				AmountCents:   amount,
			})
			errs <- err
			results <- result
		}()
	}

	// Collect all results
	for i := 0; i < n*2; i++ {
		err := <-errs
		<-results

		require.NoError(t, err)
	}

	// Check final balances (should be unchanged since equal bidirectional transfers)
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, account1.BalanceCents, updatedAccount1.BalanceCents)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.Equal(t, account2.BalanceCents, updatedAccount2.BalanceCents)
}

func TestTransferMoneyTxManyToOne(t *testing.T) {
	store := NewStore(testDB)
	var fromAccounts []Account
	toAccount := createRandomAccountWithQueries(t, store.Queries)
	n := 10
	amount := int64(5)

	for range n {
		fromAccounts = append(fromAccounts, createRandomAccountWithQueries(t, store.Queries))
	}

	errs := make(chan error)
	results := make(chan TransferMoneyResult)

	for i := range n {
		go func(index int) {
			result, err := store.TransferMoneyTx(context.Background(), CreateTransferParams{
				FromAccountID: fromAccounts[index].ID,
				ToAccountID:   toAccount.ID,
				AmountCents:   amount,
			})
			errs <- err
			results <- result
		}(i)
	}

	for range n {
		err := <-errs
		result := <-results

		require.NoError(t, err)
		require.NotEmpty(t, result)

		// Check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, toAccount.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.AmountCents)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// Check transaction (fromAccount)
		fromTranx := result.FromTx
		require.NotEmpty(t, fromTranx)
		require.Equal(t, transfer.FromAccountID, fromTranx.AccountID)
		require.Equal(t, -amount, fromTranx.AmountCents)
		require.NotZero(t, fromTranx.ID)
		require.NotZero(t, fromTranx.CreatedAt)

		_, err = store.GetTransaction(context.Background(), fromTranx.ID)
		require.NoError(t, err)

		// Check transaction (toAccount)
		toTranx := result.ToTx
		require.NotEmpty(t, toTranx)
		require.Equal(t, toAccount.ID, toTranx.AccountID)
		require.Equal(t, amount, toTranx.AmountCents)
		require.NotZero(t, toTranx.ID)
		require.NotZero(t, toTranx.CreatedAt)

		_, err = store.GetTransaction(context.Background(), toTranx.ID)
		require.NoError(t, err)
	}

	// Check final balance for toAccount
	updatedToAccount, err := store.GetAccount(context.Background(), toAccount.ID)
	require.NoError(t, err)
	require.Equal(t, toAccount.BalanceCents+int64(n)*amount, updatedToAccount.BalanceCents)

	// Check final balances for all fromAccounts
	for i := range n {
		updatedFromAccount, err := store.GetAccount(context.Background(), fromAccounts[i].ID)
		require.NoError(t, err)
		require.Equal(t, fromAccounts[i].BalanceCents-amount, updatedFromAccount.BalanceCents)
	}
}

func TestDepositMoneyTx(t *testing.T) {
	store := NewStore(testDB)
	account := createRandomAccountWithQueries(t, store.Queries)
	amount := int64(10)

	result, err := store.DepositMoneyTx(context.Background(), AccountTransactionParams{
		AccountID: account.ID,
		Amount:    amount,
	})

	require.NoError(t, err)
	require.NotEmpty(t, result.Account)

	require.Equal(t, account.ID, result.Transaction.AccountID)
	require.Equal(t, TransactionTypeDeposit, result.Transaction.Type)
	require.Equal(t, amount, result.Transaction.AmountCents)
	require.Equal(t, account.BalanceCents+amount, result.Transaction.BalanceAfterCents)

	require.Equal(t, account.BalanceCents+amount, result.Account.BalanceCents)

	updatedAccount, err := store.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.Equal(t, result.Account.BalanceCents, updatedAccount.BalanceCents)
}

func TestConcurrentDepositMoneyTx(t *testing.T) {
	store := NewStore(testDB)
	account := createRandomAccountWithQueries(t, store.Queries)
	amount := int64(40)
	n := 10

	errors := make(chan error, n) //  buffered channels -> call goroutine any order
	results := make(chan AccountTransactionResult, n)

	for range n {
		go func() {
			result, err := store.DepositMoneyTx(context.Background(), AccountTransactionParams{
				AccountID: account.ID,
				Amount:    amount,
			})
			results <- result
			errors <- err
		}()
	}

	for range n {
		err := <-errors
		result := <-results
		require.NoError(t, err)
		require.NotEmpty(t, result)

		require.NotEmpty(t, result.Transaction)
		require.NotEmpty(t, result.Account)
		require.Equal(t, account.ID, result.Transaction.AccountID)
		require.Equal(t, TransactionTypeDeposit, result.Transaction.Type)
		require.Equal(t, amount, result.Transaction.AmountCents)
	}

	updatedAccount, err := store.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	expectedBalance := account.BalanceCents + (int64(n) * amount)
	require.Equal(t, expectedBalance, updatedAccount.BalanceCents)
}

func TestWithdrawMoneyTx(t *testing.T) {
	store := NewStore(testDB)
	account := createRandomAccountWithQueries(t, store.Queries)
	amount := int64(40)
	result, err := store.WithdrawMoneyTx(context.Background(), AccountTransactionParams{
		AccountID: account.ID,
		Amount:    amount,
	})

	require.NoError(t, err)
	require.NotEmpty(t, result.Account)

	require.Equal(t, account.ID, result.Transaction.AccountID)
	require.Equal(t, TransactionTypeWithdrawal, result.Transaction.Type)
	require.Equal(t, -amount, result.Transaction.AmountCents)
	require.Equal(t, account.BalanceCents-amount, result.Transaction.BalanceAfterCents)

	// Also verify the account balance was updated
	require.Equal(t, account.BalanceCents-amount, result.Account.BalanceCents)

	// Verify persistence
	updatedAccount, err := store.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.Equal(t, result.Account.BalanceCents, updatedAccount.BalanceCents)
}

func TestConcurrentWithdrawMoneyTx(t *testing.T) {
	store := NewStore(testDB)
	account := createRandomAccountWithQueries(t, store.Queries)
	amount := int64(40)
	n := 10

	// Ensure account has enough balance for all withdrawals
	// Deposit enough money first
	initialDeposit := int64(n) * amount
	_, err := store.DepositMoneyTx(context.Background(), AccountTransactionParams{
		AccountID: account.ID,
		Amount:    initialDeposit,
	})
	require.NoError(t, err)

	// Get updated account balance
	account, err = store.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)

	errors := make(chan error, n) // buffered channels -> call goroutine any order
	results := make(chan AccountTransactionResult, n)

	// Run n concurrent withdrawals
	for range n {
		go func() {
			result, err := store.WithdrawMoneyTx(context.Background(), AccountTransactionParams{
				AccountID: account.ID,
				Amount:    amount,
			})
			results <- result
			errors <- err
		}()
	}

	// Collect and validate results
	for range n {
		err := <-errors
		result := <-results
		require.NoError(t, err)
		require.NotEmpty(t, result)

		require.NotEmpty(t, result.Transaction)
		require.NotEmpty(t, result.Account)
		require.Equal(t, account.ID, result.Transaction.AccountID)
		require.Equal(t, TransactionTypeWithdrawal, result.Transaction.Type)
		require.Equal(t, -amount, result.Transaction.AmountCents)
	}

	// Verify final balance
	updatedAccount, err := store.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	expectedBalance := account.BalanceCents - (int64(n) * amount)
	require.Equal(t, expectedBalance, updatedAccount.BalanceCents)
}

func TestWithdrawMoneyTx_InsufficientBalance(t *testing.T) {
	store := NewStore(testDB)
	account := createRandomAccountWithQueries(t, store.Queries)

	// Try to withdraw more than current balance
	amount := account.BalanceCents + 100

	_, err := store.WithdrawMoneyTx(context.Background(), AccountTransactionParams{
		AccountID: account.ID,
		Amount:    amount,
	})

	// Should fail with insufficient balance error
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInsufficientBalance)

	// Verify balance unchanged
	updatedAccount, err := store.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.Equal(t, account.BalanceCents, updatedAccount.BalanceCents)
}

func TestWithdrawMoneyTx_InvalidAmount(t *testing.T) {
	store := NewStore(testDB)
	account := createRandomAccountWithQueries(t, store.Queries)

	testCases := []struct {
		name   string
		amount int64
	}{
		{
			name:   "negative amount",
			amount: -100,
		},
		{
			name:   "zero amount",
			amount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := store.WithdrawMoneyTx(context.Background(), AccountTransactionParams{
				AccountID: account.ID,
				Amount:    tc.amount,
			})

			// Should fail with invalid amount error
			require.Error(t, err)
			require.ErrorIs(t, err, ErrInvalidAmount)

			// Verify balance unchanged
			updatedAccount, err := store.GetAccount(context.Background(), account.ID)
			require.NoError(t, err)
			require.Equal(t, account.BalanceCents, updatedAccount.BalanceCents)
		})
	}
}

func TestTransferMoneyTx_SameAccount(t *testing.T) {
	store := NewStore(testDB)
	account := createRandomAccountWithQueries(t, store.Queries)
	amount := int64(100)

	// Try to transfer to the same account
	_, err := store.TransferMoneyTx(context.Background(), CreateTransferParams{
		FromAccountID: account.ID,
		ToAccountID:   account.ID,
		AmountCents:   amount,
	})

	// Should fail
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot transfer to the same account")

	// Verify balance unchanged
	updatedAccount, err := store.GetAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.Equal(t, account.BalanceCents, updatedAccount.BalanceCents)
}

func TestTransferMoneyTx_InsufficientBalance(t *testing.T) {
	store := NewStore(testDB)
	fromAccount := createRandomAccountWithQueries(t, store.Queries)
	toAccount := createRandomAccountWithQueries(t, store.Queries)

	// Try to transfer more than available balance
	amount := fromAccount.BalanceCents + 100

	_, err := store.TransferMoneyTx(context.Background(), CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		AmountCents:   amount,
	})

	// Should fail (transaction will be rolled back)
	require.Error(t, err)

	// Verify both balances unchanged
	updatedFromAccount, err := store.GetAccount(context.Background(), fromAccount.ID)
	require.NoError(t, err)
	require.Equal(t, fromAccount.BalanceCents, updatedFromAccount.BalanceCents)

	updatedToAccount, err := store.GetAccount(context.Background(), toAccount.ID)
	require.NoError(t, err)
	require.Equal(t, toAccount.BalanceCents, updatedToAccount.BalanceCents)
}
