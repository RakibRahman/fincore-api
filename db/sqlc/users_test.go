package sqlc

import (
	"context"
	"testing"

	"github.com/RakibRahman/fincore-api/utils"
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
