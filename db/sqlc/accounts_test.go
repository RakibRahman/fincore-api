package sqlc

import (
	"context"
	"testing"

	"github.com/RakibRahman/fincore-api/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	// First create a user to own the account
	user := createRandomUser(t)

	arg := CreateAccountParams{
		OwnerID:      user.ID,
		BalanceCents: utils.RandomInt(0, 10000),
		Currency:     CurrencyUSD,
	}

	ctx := context.Background()
	account, err := testQueries.CreateAccount(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	return account
}

func TestCreateAccount(t *testing.T) {
	account := createRandomAccount(t)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.OwnerID)
	require.NotZero(t, account.CreatedAt)
	require.NotEmpty(t, account.Currency)
	require.Equal(t, AccountStatusActive, account.Status)
}

func TestGetAccount(t *testing.T) {
	account1 := createRandomAccount(t)
	ctx := context.Background()
	account2, err := testQueries.GetAccount(ctx, account1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.OwnerID, account2.OwnerID)
	require.Equal(t, account1.BalanceCents, account2.BalanceCents)
	require.Equal(t, account1.Currency, account2.Currency)
	require.Equal(t, account1.Status, account2.Status)
}

func TestUpdateAccount(t *testing.T) {
	account1 := createRandomAccount(t)
	ctx := context.Background()

	arg := UpdateAccountParams{
		ID:           account1.ID,
		BalanceCents: utils.RandomInt(0, 10000),
	}

	account2, err := testQueries.UpdateAccount(ctx, arg)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, arg.BalanceCents, account2.BalanceCents)
	require.Equal(t, account1.OwnerID, account2.OwnerID)
	require.Equal(t, account1.Currency, account2.Currency)
}

func TestDeleteAccount(t *testing.T) {
	account1 := createRandomAccount(t)
	ctx := context.Background()

	err := testQueries.DeleteAccount(ctx, account1.ID)
	require.NoError(t, err)

	// Verify account is deleted
	account2, err := testQueries.GetAccount(ctx, account1.ID)
	require.Error(t, err)
	require.Empty(t, account2)
}

func TestListAccounts(t *testing.T) {
	ctx := context.Background()
	const accountLimit = 5

	for i := 0; i < accountLimit; i++ {
		createRandomAccount(t)
	}

	params := ListAccountsParams{
		Limit:  accountLimit,
		Offset: 0,
	}

	accounts, err := testQueries.ListAccounts(ctx, params)

	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	require.GreaterOrEqual(t, len(accounts), accountLimit)

	for _, account := range accounts {
		require.NotEmpty(t, account.ID)
		require.NotEmpty(t, account.OwnerID)
		require.NotEmpty(t, account.Currency)
		require.NotEmpty(t, account.Status)
	}
}

// Negative test cases

func TestGetAccountNotFound(t *testing.T) {
	ctx := context.Background()
	// Use a non-existent account ID
	fakeID := int64(99999999)

	account, err := testQueries.GetAccount(ctx, fakeID)

	require.Error(t, err)
	require.Empty(t, account.ID)
}

func TestCreateAccountInvalidOwner(t *testing.T) {
	ctx := context.Background()
	// Try to create account with non-existent owner
	var fakeOwnerID pgtype.UUID
	fakeOwnerID.Scan("00000000-0000-0000-0000-000000000000")

	arg := CreateAccountParams{
		OwnerID:      fakeOwnerID,
		BalanceCents: utils.RandomInt(0, 10000),
		Currency:     CurrencyUSD,
	}

	account, err := testQueries.CreateAccount(ctx, arg)

	require.Error(t, err) // Should fail due to foreign key constraint
	require.Empty(t, account.ID)
}

func TestUpdateAccountNotFound(t *testing.T) {
	ctx := context.Background()
	// Try to update a non-existent account
	fakeID := int64(99999999)

	arg := UpdateAccountParams{
		ID:           fakeID,
		BalanceCents: utils.RandomInt(0, 10000),
	}

	account, err := testQueries.UpdateAccount(ctx, arg)

	require.Error(t, err)
	require.Empty(t, account.ID)
}

func TestDeleteAccountNotFound(t *testing.T) {
	ctx := context.Background()
	// Try to delete a non-existent account
	fakeID := int64(99999999)

	err := testQueries.DeleteAccount(ctx, fakeID)

	// Delete might not error on non-existent ID (depends on implementation)
	// But we can verify the account doesn't exist
	require.NoError(t, err) // DELETE typically doesn't error if ID not found
}
