package sqlc

import (
	"context"
	"testing"

	"github.com/RakibRahman/fincore-api/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomTransaction(t *testing.T) Transaction {
	// First create an account for the transaction
	account := createRandomAccount(t)

	amountCents := utils.RandomInt(100, 5000)
	balanceAfter := account.BalanceCents + amountCents

	arg := CreateTransactionParams{
		AccountID:         account.ID,
		Type:              TransactionTypeDeposit,
		AmountCents:       amountCents,
		BalanceAfterCents: balanceAfter,
	}

	ctx := context.Background()
	transaction, err := testQueries.CreateTransaction(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, transaction)

	return transaction
}

func TestCreateTransaction(t *testing.T) {
	transaction := createRandomTransaction(t)

	require.NotZero(t, transaction.ID)
	require.NotZero(t, transaction.AccountID)
	require.NotZero(t, transaction.AmountCents)
	require.NotZero(t, transaction.BalanceAfterCents)
	require.NotEmpty(t, transaction.Type)
	require.NotZero(t, transaction.CreatedAt)
}

func TestGetTransaction(t *testing.T) {
	transaction1 := createRandomTransaction(t)
	ctx := context.Background()
	transaction2, err := testQueries.GetTransaction(ctx, transaction1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, transaction2)

	require.Equal(t, transaction1.ID, transaction2.ID)
	require.Equal(t, transaction1.AccountID, transaction2.AccountID)
	require.Equal(t, transaction1.Type, transaction2.Type)
	require.Equal(t, transaction1.AmountCents, transaction2.AmountCents)
	require.Equal(t, transaction1.BalanceAfterCents, transaction2.BalanceAfterCents)
}

func TestListTransactions(t *testing.T) {
	ctx := context.Background()
	const transactionLimit = 5

	// Create an account and add multiple transactions to it
	account := createRandomAccount(t)

	for i := 0; i < transactionLimit; i++ {
		amountCents := utils.RandomInt(100, 5000)
		balanceAfter := account.BalanceCents + amountCents

		arg := CreateTransactionParams{
			AccountID:         account.ID,
			Type:              TransactionTypeDeposit,
			AmountCents:       amountCents,
			BalanceAfterCents: balanceAfter,
		}

		_, err := testQueries.CreateTransaction(ctx, arg)
		require.NoError(t, err)
	}

	params := ListTransactionsParams{
		AccountID: account.ID,
		Limit:     transactionLimit,
		Offset:    0,
	}

	transactions, err := testQueries.ListTransactions(ctx, params)

	require.NoError(t, err)
	require.NotEmpty(t, transactions)

	require.GreaterOrEqual(t, len(transactions), transactionLimit)

	for _, transaction := range transactions {
		require.NotEmpty(t, transaction.ID)
		require.Equal(t, account.ID, transaction.AccountID)
		require.NotEmpty(t, transaction.Type)
		require.NotZero(t, transaction.AmountCents)
		require.NotZero(t, transaction.BalanceAfterCents)
	}
}

// Negative test cases

func TestGetTransactionNotFound(t *testing.T) {
	ctx := context.Background()
	// Use a non-existent transaction ID
	var fakeID pgtype.UUID
	fakeID.Scan("00000000-0000-0000-0000-000000000000")

	transaction, err := testQueries.GetTransaction(ctx, fakeID)

	require.Error(t, err)
	require.Empty(t, transaction.ID)
}

func TestCreateTransactionInvalidAccount(t *testing.T) {
	ctx := context.Background()
	// Try to create transaction with non-existent account
	fakeAccountID := int64(99999999)

	arg := CreateTransactionParams{
		AccountID:         fakeAccountID,
		Type:              TransactionTypeDeposit,
		AmountCents:       utils.RandomInt(100, 5000),
		BalanceAfterCents: utils.RandomInt(1000, 10000),
	}

	transaction, err := testQueries.CreateTransaction(ctx, arg)

	require.Error(t, err) // Should fail due to foreign key constraint
	require.Empty(t, transaction.ID)
}

func TestListTransactionsEmptyResult(t *testing.T) {
	ctx := context.Background()
	// Create an account with no transactions
	account := createRandomAccount(t)

	params := ListTransactionsParams{
		AccountID: account.ID,
		Limit:     10,
		Offset:    0,
	}

	transactions, err := testQueries.ListTransactions(ctx, params)

	require.NoError(t, err)
	require.Empty(t, transactions) // Should return empty list, not error
}
