package jwt

import (
	"fmt"
	"time"

	"github.com/crxfoz/teaserad/crmadm/internal/domain/entity"
	"github.com/dgrijalva/jwt-go"
)

type UserClaims struct {
	jwt.StandardClaims
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	Role     string `json:"role"`
}

type JWTManager struct {
	tokenDuration time.Duration
	secretKey     string
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

func (manager *JWTManager) Generate(user entity.User) (entity.UserContext, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		},
		Username: user.Username,
		UserID:   user.ID,
		Role:     user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenKey, err := token.SignedString([]byte(manager.secretKey))
	if err != nil {
		return entity.UserContext{}, err
	}

	return entity.UserContext{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		Token:    tokenKey,
	}, nil
}

func (manager *JWTManager) Verify(accessToken string) (entity.UserContext, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return entity.UserContext{}, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return entity.UserContext{}, fmt.Errorf("invalid token claims")
	}

	return entity.UserContext{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		Token:    accessToken,
	}, nil
}
