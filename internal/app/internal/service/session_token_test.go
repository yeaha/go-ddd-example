package service

import (
	"testing"

	"ddd-example/internal/domain"

	"github.com/google/uuid"
)

func TestSessionTokenService(t *testing.T) {
	account := &domain.Account{
		ID:          uuid.New(),
		SessionSalt: "54381095jfepwoqrp2",
	}
	token := newSessionToken(account)

	service := SessionTokenService{}
	payload := service.encode(token, account.SessionSalt)
	if payload == "" {
		t.Fatal("payload should not be empty")
	}

	decoded, err := service.decode(payload)
	if err != nil {
		t.Fatal(err)
	} else if decoded != token {
		t.Fatalf("decoded token should be equal to token")
	} else if v := service.encode(decoded, account.SessionSalt); v != payload {
		t.Fatalf("encoded token should be equal to payload")
	} else if v := service.encode(decoded, "abcfaof"); v == payload {
		t.Fatalf("encoded token should not be equal to payload")
	}
}
