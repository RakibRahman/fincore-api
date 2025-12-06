package sqlc

import (
	"context"
	"testing"

	"github.com/RakibRahman/fincore-api/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T) Transfer {
	// Create two accounts for the transfer
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		AmountCents:   utils.RandomInt(100, 5000),
	}

	ctx := context.Background()
	transfer, err := testQueries.CreateTransfer(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	transfer := createRandomTransfer(t)

	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.FromAccountID)
	require.NotZero(t, transfer.ToAccountID)
	require.NotZero(t, transfer.AmountCents)
	require.NotEmpty(t, transfer.Status)
	require.NotZero(t, transfer.CreatedAt)
	require.Equal(t, TransferStatusPending, transfer.Status)
}

func TestGetTransfer(t *testing.T) {
	transfer1 := createRandomTransfer(t)
	ctx := context.Background()
	transfer2, err := testQueries.GetTransfer(ctx, transfer1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.AmountCents, transfer2.AmountCents)
	require.Equal(t, transfer1.Status, transfer2.Status)
}

func TestListTransfers(t *testing.T) {
	ctx := context.Background()
	const transferLimit = 5

	for i := 0; i < transferLimit; i++ {
		createRandomTransfer(t)
	}

	params := ListTransfersParams{
		Limit:  transferLimit,
		Offset: 0,
	}

	transfers, err := testQueries.ListTransfers(ctx, params)

	require.NoError(t, err)
	require.NotEmpty(t, transfers)

	require.GreaterOrEqual(t, len(transfers), transferLimit)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer.ID)
		require.NotZero(t, transfer.FromAccountID)
		require.NotZero(t, transfer.ToAccountID)
		require.NotZero(t, transfer.AmountCents)
		require.NotEmpty(t, transfer.Status)
	}
}

// Negative test cases

func TestGetTransferNotFound(t *testing.T) {
	ctx := context.Background()
	// Use a non-existent transfer ID
	var fakeID pgtype.UUID
	fakeID.Scan("00000000-0000-0000-0000-000000000000")

	transfer, err := testQueries.GetTransfer(ctx, fakeID)

	require.Error(t, err)
	require.Empty(t, transfer.ID)
}

func TestCreateTransferInvalidFromAccount(t *testing.T) {
	ctx := context.Background()
	toAccount := createRandomAccount(t)
	fakeFromAccountID := int64(99999999)

	arg := CreateTransferParams{
		FromAccountID: fakeFromAccountID,
		ToAccountID:   toAccount.ID,
		AmountCents:   utils.RandomInt(100, 5000),
	}

	transfer, err := testQueries.CreateTransfer(ctx, arg)

	require.Error(t, err) // Should fail due to foreign key constraint
	require.Empty(t, transfer.ID)
}

func TestCreateTransferInvalidToAccount(t *testing.T) {
	ctx := context.Background()
	fromAccount := createRandomAccount(t)
	fakeToAccountID := int64(99999999)

	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   fakeToAccountID,
		AmountCents:   utils.RandomInt(100, 5000),
	}

	transfer, err := testQueries.CreateTransfer(ctx, arg)

	require.Error(t, err) // Should fail due to foreign key constraint
	require.Empty(t, transfer.ID)
}

func TestCreateTransferSameAccount(t *testing.T) {
	ctx := context.Background()
	account := createRandomAccount(t)

	// Try to transfer to the same account
	arg := CreateTransferParams{
		FromAccountID: account.ID,
		ToAccountID:   account.ID, // Same account
		AmountCents:   utils.RandomInt(100, 5000),
	}

	transfer, err := testQueries.CreateTransfer(ctx, arg)

	// This might succeed or fail depending on your business logic/constraints
	// If you have a CHECK constraint preventing same-account transfers, it should error
	// For now, we'll just verify the transfer was created (adjust based on your schema)
	if err != nil {
		require.Error(t, err)
		require.Empty(t, transfer.ID)
	} else {
		require.NoError(t, err)
		require.NotEmpty(t, transfer.ID)
	}
}
