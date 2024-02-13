package token

import (
	"bank/util"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	
	maker := NewPasetoMaker(paseto.NewV4SymmetricKey())
	
	username := util.RandomOwner()

	duration := time.Minute
	issuedAt := time.Now()
	expireAt := issuedAt.Add(duration)

	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expireAt, payload.ExpireAt, time.Second)

}

func TestExpiredPaseto(t *testing.T) {
	maker := NewPasetoMaker(paseto.NewV4SymmetricKey())
	
	username := util.RandomOwner()

	duration := -time.Minute

	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}


