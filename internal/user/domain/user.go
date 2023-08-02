package domain

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
	uuid "github.com/satori/go.uuid"
)

// User 系统用户
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Password     string    `json:"-"`
	PasswordSalt string    `json:"-"`
	SessionSalt  string    `json:"-"`
}

// SetPassword 设置密码
func (u *User) SetPassword(password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return errors.New("empty password")
	}

	if err := u.refreshPasswordSalt(); err != nil {
		return fmt.Errorf("refresh password salt, %w", err)
	}

	u.Password = newPassword(password, u.PasswordSalt)
	return nil
}

// SetEmail 设置email
func (u *User) SetEmail(email string) error {
	email = NormalizeEmail(email)
	if !govalidator.IsEmail(email) {
		return errors.New("invalid email")
	}

	u.Email = email
	return nil
}

// ComparePassword 验证密码是否一致
func (u *User) ComparePassword(password string) bool {
	return password != "" &&
		u.Password == newPassword(password, u.PasswordSalt)
}

// RefreshSessionSalt 更新会话签名盐，更新后同一账号的其它会话会自动失效
func (u *User) RefreshSessionSalt() error {
	s, err := newSalt(8)
	if err != nil {
		return err
	}
	u.SessionSalt = s
	return nil
}

func (u *User) refreshPasswordSalt() error {
	s, err := newSalt(8)
	if err != nil {
		return err
	}
	u.PasswordSalt = s
	return nil
}

func newSalt(length int) (string, error) {
	data := make([]byte, length)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", data), nil
}

func newPassword(password string, salt string) string {
	data := append([]byte(password), []byte(salt)...)
	return fmt.Sprintf("%x", md5.Sum(data))
}

// NormalizeEmail 规范化email输入
func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
