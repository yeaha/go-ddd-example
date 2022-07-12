package service

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	uuid "github.com/satori/go.uuid"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/app/adapter"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
)

// SessionTokenService 会话凭证逻辑
type SessionTokenService struct {
	Users adapter.UserRepository
}

// Generate 构造会话凭证
func (s *SessionTokenService) Generate(ctx context.Context, user *domain.User) (payload string, err error) {
	if err := s.Suspend(ctx, user); err != nil {
		return "", err
	}

	return s.Renew(user)
}

// Renew 构造会话凭证，但不刷新session salt
func (s *SessionTokenService) Renew(user *domain.User) (payload string, err error) {
	token := domain.NewSessionToken(user)
	return s.encode(token, user.SessionSalt), nil
}

// Suspend 使指定账号会话失效
func (s *SessionTokenService) Suspend(ctx context.Context, user *domain.User) error {
	// 通过替换sesion salt达到会话失效的目的
	if err := user.RefreshSessionSalt(); err != nil {
		return fmt.Errorf("refresh session token, %w", err)
	} else if err := s.Users.Save(ctx, user); err != nil {
		return fmt.Errorf("save user, %w", err)
	}
	return nil
}

// Retrieve 恢复凭证内的信息
func (s *SessionTokenService) Retrieve(ctx context.Context, payload string) (*domain.User, domain.SessionToken, error) {
	token, err := s.decode(payload)
	if err != nil {
		return nil, token, fmt.Errorf("decode token, %w", err)
	}

	user, err := s.Users.Find(ctx, token.UserID)
	if err != nil {
		return nil, token, err
	}

	if s.encode(token, user.SessionSalt) != payload {
		return nil, token, errors.New("invalid token signature")
	}
	return user, token, nil
}

// 构造包含签名的token字符串
func (s *SessionTokenService) encode(token domain.SessionToken, salt string) string {
	payload := fmt.Sprintf("%s,%d", token.UserID, token.Expire)
	signature := s.sign(payload, salt)

	payload = fmt.Sprintf("%s:%s", payload, signature)
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func (s *SessionTokenService) decode(payload string) (domain.SessionToken, error) {
	data, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return domain.SessionToken{}, fmt.Errorf("base64 decode, %w", err)
	}
	payload = string(data)

	payload, _, ok := strings.Cut(payload, ":")
	if !ok {
		return domain.SessionToken{}, domain.ErrInvalidSessionToken
	}

	id, expire, ok := strings.Cut(payload, ",")
	if !ok {
		return domain.SessionToken{}, domain.ErrInvalidSessionToken
	}

	userID, err := uuid.FromString(id)
	if err != nil {
		return domain.SessionToken{}, fmt.Errorf("invalid user id, %w", err)
	}

	expireTime, err := strconv.Atoi(expire)
	if err != nil {
		return domain.SessionToken{}, fmt.Errorf("invalid expire time, %w", err)
	}

	return domain.SessionToken{UserID: userID, Expire: int64(expireTime)}, nil
}

// 计算数字签名
func (s *SessionTokenService) sign(payload string, salt string) string {
	signature := fmt.Sprintf("%s,%s", payload, salt)
	return fmt.Sprintf("%x", md5.Sum([]byte(signature)))
}
