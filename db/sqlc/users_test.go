package sqlc

import (
	"context"
	"testing"

	"github.com/RakibRahman/fincore-api/utils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		FirstName:    utils.RandomString(6),
		LastName:     utils.RandomString(4),
		Email:        utils.RandomEmail(),
		PasswordHash: utils.RandomString(12),
	}
	ctx := context.Background()
	user, err := testQueries.CreateUser(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	return user
}

func TestCreateUser(t *testing.T) {
	user := createRandomUser(t)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)
	require.NotEmpty(t, user.Email)

}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	ctx := context.Background()
	user2, err := testQueries.GetUser(ctx, user1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.FirstName, user2.FirstName)
}

func TestUpdateUser(t *testing.T) {
	user1 := createRandomUser(t)
	ctx := context.Background()

	arg := UpdateUserParams{
		ID:        user1.ID,
		FirstName: utils.RandomString(6),
		LastName:  utils.RandomString(4),
		Email:     utils.RandomEmail(),
	}

	user2, err := testQueries.UpdateUser(ctx, arg)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, arg.FirstName, user2.FirstName)
	require.Equal(t, arg.LastName, user2.LastName)
	require.Equal(t, arg.Email, user2.Email)
}

func TestListUsers(t *testing.T) {
	ctx := context.Background()
	const userLimit = 5
	var userList1 []User
	for i := 0; i < userLimit; i++ {
		userList1 = append(userList1, createRandomUser(t))
	}

	params := ListUsersParams{
		Limit:  userLimit,
		Offset: 0,
	}

	users, err := testQueries.ListUsers(ctx, params)

	require.NoError(t, err)
	require.NotEmpty(t, users)

	require.GreaterOrEqual(t, len(users), userLimit)

	for _, user := range users {
		require.NotEmpty(t, user.ID)
		require.NotEmpty(t, user.Email)
		require.NotEmpty(t, user.FirstName)
		require.NotEmpty(t, user.LastName)
	}
}

// Negative test cases

func TestGetUserNotFound(t *testing.T) {
	ctx := context.Background()
	// Use a non-existent UUID
	var fakeID pgtype.UUID
	fakeID.Scan("00000000-0000-0000-0000-000000000000")

	user, err := testQueries.GetUser(ctx, fakeID)

	require.Error(t, err)
	require.Empty(t, user.ID)
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	user1 := createRandomUser(t)
	ctx := context.Background()

	// Try to create another user with the same email (should violate unique constraint)
	arg := CreateUserParams{
		FirstName:    utils.RandomString(6),
		LastName:     utils.RandomString(4),
		Email:        user1.Email, // Duplicate email
		PasswordHash: utils.RandomString(12),
	}

	user2, err := testQueries.CreateUser(ctx, arg)

	require.Error(t, err)
	require.Empty(t, user2.ID)
}

func TestUpdateUserNotFound(t *testing.T) {
	ctx := context.Background()
	// Try to update a non-existent user
	var fakeID pgtype.UUID
	fakeID.Scan("00000000-0000-0000-0000-000000000000")

	arg := UpdateUserParams{
		ID:        fakeID,
		FirstName: utils.RandomString(6),
		LastName:  utils.RandomString(4),
		Email:     utils.RandomEmail(),
	}

	user, err := testQueries.UpdateUser(ctx, arg)

	require.Error(t, err)
	require.Empty(t, user.ID)
}

func TestUpdateUserDuplicateEmail(t *testing.T) {
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)
	ctx := context.Background()

	// Try to update user2 with user1's email
	arg := UpdateUserParams{
		ID:        user2.ID,
		FirstName: user2.FirstName,
		LastName:  user2.LastName,
		Email:     user1.Email, // Duplicate email
	}

	updatedUser, err := testQueries.UpdateUser(ctx, arg)

	require.Error(t, err)
	require.Empty(t, updatedUser.ID)
}
