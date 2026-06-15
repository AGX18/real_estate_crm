package auth

import (
	"fmt"
	"strings"
	"time"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 24 * time.Hour

type TokenManager struct {
	secret []byte
	now    func() time.Time
}

type Claims struct {
	BrokerID int64   `json:"broker_id"`
	TenantID string  `json:"tenant_id"`
	Email    string  `json:"email"`
	Role     db.Role `json:"role"`
	jwt.RegisteredClaims
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		now:    time.Now,
	}
}

func (m *TokenManager) Generate(claims Claims) (string, error) {
	if claims.ExpiresAt == nil {
		claims.ExpiresAt = jwt.NewNumericDate(m.now().Add(tokenTTL))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func tenantIDString(tenantID fmt.Stringer) string {
	value := tenantID.String()
	return strings.Trim(value, "{}")
}
