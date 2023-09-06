package db

import (
	"bank/util"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T) User {
	arg := CreateUserParams{
		Username: util.RandomOwner(),
		HashedPassword: "secretPassword",
		FullName: util.RandomOwner(),
		Email: util.RandomEmail(),
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.NotZero(t, user.CreatedAt)
	require.True(t, user.PaswordChangedAt.IsZero())

	return user
}

func TestCreateUser(t *testing.T) {
	createTestUser(t)
}

func TestGetUser(t *testing.T) {
	user := createTestUser(t)
	testUser, err := testQueries.GetUser(context.Background(), user.Username)
	
	require.NoError(t, err)
	require.NotEmpty(t, testUser)

	require.Equal(t, user.Username, testUser.Username)
	require.Equal(t, user.HashedPassword, testUser.HashedPassword)
	require.Equal(t, user.FullName, testUser.FullName)
	require.Equal(t, user.Email, testUser.Email)
	require.WithinDuration(t, user.CreatedAt, testUser.CreatedAt, time.Second)
	require.WithinDuration(t, user.PaswordChangedAt, testUser.PaswordChangedAt, time.Second)
}