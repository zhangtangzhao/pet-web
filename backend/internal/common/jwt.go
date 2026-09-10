package common

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeMember = "member"
	TokenTypeAdmin  = "admin"
)

func GenToken(secret string, expire time.Duration, uid int64, typ, role string) (string, error) {
	claims := jwt.MapClaims{
		"uid": uid,
		"typ": typ,
		"exp": time.Now().Add(expire).Unix(),
	}
	if role != "" {
		claims["role"] = role
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ParseToken(secret, token string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !t.Valid {
		return nil, ErrTokenExpired
	}
	mc, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenExpired
	}
	return mc, nil
}

// BearerToken 从 Authorization 头提取 token
func BearerToken(r *http.Request) string { //nolint:revive // 保持无依赖
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if len(h) > len(prefix) && h[:len(prefix)] == prefix {
		return h[len(prefix):]
	}
	return ""
}
