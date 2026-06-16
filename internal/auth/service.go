package auth

import (
	"context"
	"errors"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type Service struct {
	queries      db.Querier
	tokenManager *TokenManager
}

type LoginParams struct {
	TenantName string
	Email      string
	Password   string
}

type BrokerResponse struct {
	ID       int64       `json:"id"`
	TenantID pgtype.UUID `json:"tenant_id"`
	Username string      `json:"username"`
	Email    string      `json:"email"`
	Role     db.Role     `json:"role"`
}

type LoginResult struct {
	Token  string         `json:"token"`
	Broker BrokerResponse `json:"broker"`
}

func NewService(queries db.Querier, tokenManager *TokenManager) *Service {
	return &Service{queries: queries, tokenManager: tokenManager}
}

func (s *Service) Login(ctx context.Context, params LoginParams) (LoginResult, error) {
	tenant, err := s.queries.GetTenantByName(ctx, params.TenantName)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	broker, err := s.queries.GetBrokerByEmail(ctx, db.GetBrokerByEmailParams{
		Email:    params.Email,
		TenantID: tenant.ID,
	})
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(broker.PasswordHash), []byte(params.Password)); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	token, err := s.tokenManager.Generate(Claims{
		BrokerID: broker.ID,
		TenantID: tenantIDString(broker.TenantID),
		Email:    broker.Email,
		Role:     broker.Role,
	})
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		Token: token,
		Broker: BrokerResponse{
			ID:       broker.ID,
			TenantID: broker.TenantID,
			Username: broker.Username,
			Email:    broker.Email,
			Role:     broker.Role,
		},
	}, nil
}
