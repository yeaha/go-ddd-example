package service

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
)

func TestSessionTokenService(t *testing.T) {
	user := &domain.User{
		ID:          uuid.NewV4(),
		SessionSalt: "54381095jfepwoqrp2",
	}
	token := domain.NewSessionToken(user)

	service := SessionTokenService{}
	payload := service.encode(token, user.SessionSalt)
	require.NotEmpty(t, payload)

	decoded, err := service.decode(payload)
	require.NoError(t, err)

	require.Equal(t, token, decoded)
	require.Equal(t, service.encode(decoded, user.SessionSalt), payload)
	require.NotEqual(t, service.encode(decoded, "abcfaof"), payload)
}
