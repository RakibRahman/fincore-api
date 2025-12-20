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

	// run n concurrent transfer transactions to make sure transaction works smoothly
	n := 5
	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferMoneyResult)

	for range n {
		go func() {
			result, err := store.TransferMoneyTx(context.Background(), CreateTransferParams{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				AmountCents:   amount,
			})

			errs <- err       //send err to errs channel
			results <- result // send result to results channel
		}()
	}

	// verify results

	for range n {
		err := <-errs       // store err data
		result := <-results // store transaction result

		require.NoError(t, err)
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, fromAccount.ID, transfer.FromAccountID)
		require.Equal(t, toAccount.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.AmountCents)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check transaction (fromAccount)
		fromTranx := result.FromTx
		require.NotEmpty(t, fromTranx)
		require.Equal(t, fromAccount.ID, fromTranx.AccountID)
		require.Equal(t, -amount, fromTranx.AmountCents)
		require.NotZero(t, fromTranx.ID)
		require.NotZero(t, fromTranx.CreatedAt)

		_, err = store.GetTransaction(context.Background(), fromTranx.ID)
		require.NoError(t, err)

		// check transaction (toAccount)
		toTranx := result.ToTx
		require.NotEmpty(t, toTranx)
		require.Equal(t, toAccount.ID, toTranx.AccountID)
		require.Equal(t, amount, toTranx.AmountCents)
		require.NotZero(t, toTranx.ID)
		require.NotZero(t, toTranx.CreatedAt)

		_, err = store.GetTransaction(context.Background(), toTranx.ID)
		require.NoError(t, err)
	}

	// Check final account balances after all transfers
	updatedFromAccount, err := store.GetAccountForUpdate(context.Background(), fromAccount.ID)
	require.NoError(t, err)

	updatedToAccount, err := store.GetAccountForUpdate(context.Background(), toAccount.ID)
	require.NoError(t, err)

	// Verify the total amount transferred
	require.Equal(t, fromAccount.BalanceCents-int64(n)*amount, updatedFromAccount.BalanceCents)
	require.Equal(t, toAccount.BalanceCents+int64(n)*amount, updatedToAccount.BalanceCents)
}

func TestTransferMoneyTx_Bidirectional(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccountWithQueries(t, store.Queries)
	account2 := createRandomAccountWithQueries(t, store.Queries)

	n := 5
	amount := int64(10)

	//channels
	errs := make(chan error)
	results := make(chan TransferMoneyResult)

	for range n {
		go func() {
			result, err := store.TransferMoneyTx(context.Background(), CreateTransferParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				AmountCents:   amount,
			})

			errs <- err
			results <- result
		}()
	}
	for range n {
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

	for range 2 * n {
		err := <-errs
		result := <-results

		require.NoError(t, err)
		require.NotEmpty(t, result)
	}

	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount1)
	require.Equal(t, updatedAccount1.BalanceCents, account1.BalanceCents)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount2)
	require.Equal(t, updatedAccount2.BalanceCents, account2.BalanceCents)
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
