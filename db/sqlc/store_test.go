package sqlc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferMoneyTx(t *testing.T) {
	store := NewStore(testDB)

	// Create accounts directly using store (not in a test transaction)
	// These will be committed to the database and available for TransferMoneyTx
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
