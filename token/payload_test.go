package token

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPayload(t *testing.T) {
	id := uuid.New()
	username := "testuser"
	issuedAt := time.Now()
	expireAt := time.Now().Add(time.Hour)

	payload := Payload{
		ID:       id,
		Username: username,
		IssuedAt: issuedAt,
		ExpireAt: expireAt,
	}

	require.Equal(t, id, payload.ID)
	require.Equal(t, username, payload.Username)
	require.Equal(t, issuedAt, payload.IssuedAt)
	require.Equal(t, expireAt, payload.ExpireAt)
}