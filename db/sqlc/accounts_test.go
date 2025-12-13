package sqlc

import (
	"context"
	"testing"

	"github.com/RakibRahman/fincore-api/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomAccountWithQueries(t *testing.T, q *Queries) Account {
	user := createRandomUserWithQueries(t, q)

	arg := CreateAccountParams{
		OwnerID:      user.ID,
		BalanceCents: utils.RandomInt(0, 10000),
		Currency:     CurrencyUSD,
	}

	ctx := context.Background()
	account, err := q.CreateAccount(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	return account
}

func TestCreateAccount(t *testing.T) {
	_, q := createTestTx(t)
	account := createRandomAccountWithQueries(t, q)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.OwnerID)
	require.NotZero(t, account.CreatedAt)
	require.NotEmpty(t, account.Currency)
	require.Equal(t, AccountStatusActive, account.Status)
}

func TestGetAccount(t *testing.T) {
	_, q := createTestTx(t)
	account1 := createRandomAccountWithQueries(t, q)
	ctx := context.Background()
	account2, err := q.GetAccount(ctx, account1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, account1.OwnerID, account2.OwnerID)
	require.Equal(t, account1.BalanceCents, account2.BalanceCents)
	require.Equal(t, account1.Currency, account2.Currency)
	require.Equal(t, account1.Status, account2.Status)
}

func TestUpdateAccount(t *testing.T) {
	_, q := createTestTx(t)
	account1 := createRandomAccountWithQueries(t, q)
	ctx := context.Background()

	arg := UpdateAccountBalanceParams{
		ID:           account1.ID,
		BalanceCents: utils.RandomInt(0, 10000),
	}

	account2, err := q.UpdateAccountBalance(ctx, arg)

	require.NoError(t, err)
	require.NotEmpty(t, account2)

	require.Equal(t, account1.ID, account2.ID)
	require.Equal(t, arg.BalanceCents, account2.BalanceCents)
	require.Equal(t, account1.OwnerID, account2.OwnerID)
	require.Equal(t, account1.Currency, account2.Currency)
}

func TestDeleteAccount(t *testing.T) {
	_, q := createTestTx(t)
	account1 := createRandomAccountWithQueries(t, q)
	ctx := context.Background()

	err := q.DeleteAccount(ctx, account1.ID)
	require.NoError(t, err)

	// Verify account is deleted
	account2, err := q.GetAccount(ctx, account1.ID)
	require.Error(t, err)
	require.Empty(t, account2)
}

func TestListAccounts(t *testing.T) {
	_, q := createTestTx(t)
	ctx := context.Background()
	const accountLimit = 5

	for i := 0; i < accountLimit; i++ {
		createRandomAccountWithQueries(t, q)
	}

	params := ListAccountsParams{
		Limit:  accountLimit,
		Offset: 0,
	}

	accounts, err := q.ListAccounts(ctx, params)

	require.NoError(t, err)
	require.Len(t, accounts, accountLimit) // Exact count since we're in isolated transaction

	for _, account := range accounts {
		require.NotEmpty(t, account.ID)
		require.NotEmpty(t, account.OwnerID)
		require.NotEmpty(t, account.Currency)
		require.NotEmpty(t, account.Status)
	}
}

// Negative test cases

func TestGetAccountNotFound(t *testing.T) {
	_, q := createTestTx(t)
	ctx := context.Background()
	// Use a non-existent account ID
	fakeID := int64(99999999)

	account, err := q.GetAccount(ctx, fakeID)

	require.Error(t, err)
	require.Empty(t, account.ID)
}

func TestCreateAccountInvalidOwner(t *testing.T) {
	_, q := createTestTx(t)
	ctx := context.Background()
	// Try to create account with non-existent owner
	var fakeOwnerID pgtype.UUID
	fakeOwnerID.Scan("00000000-0000-0000-0000-000000000000")

	arg := CreateAccountParams{
		OwnerID:      fakeOwnerID,
		BalanceCents: utils.RandomInt(0, 10000),
		Currency:     CurrencyUSD,
	}

	account, err := q.CreateAccount(ctx, arg)

	require.Error(t, err) // Should fail due to foreign key constraint
	require.Empty(t, account.ID)
}

func TestUpdateAccountNotFound(t *testing.T) {
	_, q := createTestTx(t)
	ctx := context.Background()
	// Try to update a non-existent account
	fakeID := int64(99999999)

	arg := UpdateAccountBalanceParams{
		ID:           fakeID,
		BalanceCents: utils.RandomInt(0, 10000),
	}

	account, err := q.UpdateAccountBalance(ctx, arg)

	require.Error(t, err)
	require.Empty(t, account.ID)
}

func TestDeleteAccountNotFound(t *testing.T) {
	_, q := createTestTx(t)
	ctx := context.Background()
	// Try to delete a non-existent account
	fakeID := int64(99999999)

	err := q.DeleteAccount(ctx, fakeID)

	require.NoError(t, err) // DELETE typically doesn't error if ID not found
}
