package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	Subject string `json:"sub"`
	jwt.RegisteredClaims
}

type TokenIssuer struct {
	secret   []byte
	tokenTTL time.Duration
}

func NewTokenIssuer(secret string, tokenTTL time.Duration) *TokenIssuer {
	return &TokenIssuer{
		secret:   []byte(secret),
		tokenTTL: tokenTTL,
	}
}

func (issuer *TokenIssuer) IssueToken(subject string) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(issuer.tokenTTL)

	claims := Claims{
		Subject: subject,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "parkiroid-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(issuer.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

func (issuer *TokenIssuer) ValidateToken(tokenString string) error {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return issuer.secret, nil
	})
	if err != nil {
		return ErrInvalidToken
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	return nil
}
