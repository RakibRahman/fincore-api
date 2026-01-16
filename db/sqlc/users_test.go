package sqlc

import (
	"context"
	"testing"

	"github.com/RakibRahman/fincore-api/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomUserWithQueries(t *testing.T, q *Queries) User {
	arg := CreateUserParams{
		FirstName:    utils.RandomString(6),
		LastName:     utils.RandomString(4),
		Email:        utils.RandomEmail(),
		PasswordHash: utils.RandomString(12),
	}
	ctx := context.Background()
	user, err := q.CreateUser(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	return user
}

func TestCreateUser(t *testing.T) {
	_, q := createTestTx(t)
	user := createRandomUserWithQueries(t, q)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)
	require.NotZero(t, user.UpdatedAt)
	require.NotEmpty(t, user.Email)
}

func TestGetUser(t *testing.T) {
	_, q := createTestTx(t)
	user1 := createRandomUserWithQueries(t, q)
	ctx := context.Background()
	user2, err := q.GetUserByEmail(ctx, user1.Email)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.FirstName, user2.FirstName)
}

func TestUpdateUser(t *testing.T) {
	_, q := createTestTx(t)
	user1 := createRandomUserWithQueries(t, q)
	ctx := context.Background()

	newFirstName := pgtype.Text{String: utils.RandomString(6), Valid: true}
	newLastName := pgtype.Text{String: utils.RandomString(4), Valid: true}

	arg := UpdateUserParams{
		ID:        user1.ID,
		FirstName: newFirstName,
		LastName:  newLastName,
	}

	user2, err := q.UpdateUser(ctx, arg)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email) // Email should not change
	require.Equal(t, newFirstName.String, user2.FirstName)
	require.Equal(t, newLastName.String, user2.LastName)
	require.NotZero(t, user2.UpdatedAt) // UpdatedAt should be set
}

func TestListUsers(t *testing.T) {
	_, q := createTestTx(t)
	ctx := context.Background()
	const userLimit = 5

	for i := 0; i < userLimit; i++ {
		createRandomUserWithQueries(t, q)
	}

	params := ListUsersParams{
		Limit:  userLimit,
		Offset: 0,
	}

	users, err := q.ListUsers(ctx, params)

	require.NoError(t, err)
	require.Len(t, users, userLimit)
	for _, user := range users {
		require.NotEmpty(t, user.ID)
		require.NotEmpty(t, user.Email)
		require.NotEmpty(t, user.FirstName)
		require.NotEmpty(t, user.LastName)
	}
}

// Negative test cases

func TestGetUserNotFound(t *testing.T) {
	_, q := createTestTx(t)
	ctx := context.Background()

	user, err := q.GetUserByEmail(ctx, "dummy111@fake.com")

	require.Error(t, err)
	require.Empty(t, user.Email)
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	_, q := createTestTx(t)
	user1 := createRandomUserWithQueries(t, q)
	ctx := context.Background()

	arg := CreateUserParams{
		FirstName:    utils.RandomString(6),
		LastName:     utils.RandomString(4),
		Email:        user1.Email, // Duplicate email
		PasswordHash: utils.RandomString(12),
	}

	user2, err := q.CreateUser(ctx, arg)

	require.Error(t, err)
	require.Empty(t, user2.ID)
}

func TestUpdateUserNotFound(t *testing.T) {
	_, q := createTestTx(t)
	ctx := context.Background()
	// Try to update a non-existent user
	var fakeID pgtype.UUID
	fakeID.Scan("00000000-0000-0000-0000-000000000000")

	arg := UpdateUserParams{
		ID:        fakeID,
		FirstName: pgtype.Text{String: utils.RandomString(6), Valid: true},
		LastName:  pgtype.Text{String: utils.RandomString(4), Valid: true},
	}

	user, err := q.UpdateUser(ctx, arg)

	require.Error(t, err)
	require.Empty(t, user.ID)
}

func TestUpdateUserPartialUpdate(t *testing.T) {
	_, q := createTestTx(t)
	user1 := createRandomUserWithQueries(t, q)
	ctx := context.Background()

	// Only update first name (last name should remain unchanged)
	newFirstName := pgtype.Text{String: utils.RandomString(6), Valid: true}

	arg := UpdateUserParams{
		ID:        user1.ID,
		FirstName: newFirstName,
		LastName:  pgtype.Text{Valid: false}, // NULL value, should keep existing
	}

	user2, err := q.UpdateUser(ctx, arg)

	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, newFirstName.String, user2.FirstName)
	require.Equal(t, user1.LastName, user2.LastName) // Should remain unchanged
}
