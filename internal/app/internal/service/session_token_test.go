package service

import (
	"testing"

	"ddd-example/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSessionTokenService(t *testing.T) {
	account := &domain.Account{
		ID:          uuid.New(),
		SessionSalt: "54381095jfepwoqrp2",
	}
	token := newSessionToken(account)

	service := SessionTokenService{}
	payload := service.encode(token, account.SessionSalt)
	require.NotEmpty(t, payload)

	decoded, err := service.decode(payload)
	require.NoError(t, err)

	require.Equal(t, token, decoded)
	require.Equal(t, service.encode(decoded, account.SessionSalt), payload)
	require.NotEqual(t, service.encode(decoded, "abcfaof"), payload)
}
