package auth

import (
	"context"
	"errors"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrRegistrationUnavailable = errors.New("registration unavailable")

type Service struct {
	queries      db.Querier
	store        txStore
	tokenManager *TokenManager
}

type txStore interface {
	WithTx(ctx context.Context, fn func(q db.Querier) error) error
}

type LoginParams struct {
	TenantName string
	Email      string
	Password   string
}

type RegisterParams struct {
	TenantName    string
	AdminUsername string
	AdminEmail    string
	AdminPassword string
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

func NewServiceWithStore(store txStore, queries db.Querier, tokenManager *TokenManager) *Service {
	return &Service{store: store, queries: queries, tokenManager: tokenManager}
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

func (s *Service) Register(ctx context.Context, params RegisterParams) (LoginResult, error) {
	if s.store == nil {
		return LoginResult{}, ErrRegistrationUnavailable
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return LoginResult{}, err
	}

	var broker db.Broker
	if err := s.store.WithTx(ctx, func(q db.Querier) error {
		tenant, err := q.CreateTenant(ctx, db.CreateTenantParams{
			Name: params.TenantName,
		})
		if err != nil {
			return err
		}

		broker, err = q.CreateBroker(ctx, db.CreateBrokerParams{
			TenantID:     tenant.ID,
			Username:     params.AdminUsername,
			Email:        params.AdminEmail,
			PasswordHash: string(passwordHash),
			Role:         db.RoleAdmin,
		})
		return err
	}); err != nil {
		return LoginResult{}, err
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
