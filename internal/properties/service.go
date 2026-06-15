package properties

import (
	"context"

	db "real_estate_crm/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
)

type Service struct {
	queries db.Querier
}

type CreateParams struct {
	Property  db.CreatePropertyParams
	Content   string
	Embedding []float32
}

type UpdateParams struct {
	Property  db.UpdatePropertyParams
	Content   string
	Embedding []float32
}

type PropertyResult struct {
	Property        db.Property `json:"property"`
	EmbeddingStored bool        `json:"embedding_stored"`
}

func NewService(queries db.Querier) *Service {
	return &Service{queries: queries}
}

func (s *Service) Create(ctx context.Context, params CreateParams) (PropertyResult, error) {
	property, err := s.queries.CreateProperty(ctx, params.Property)
	if err != nil {
		return PropertyResult{}, err
	}

	embeddingStored, err := s.createEmbedding(ctx, property.TenantID, property.ID, params.Content, params.Embedding)
	if err != nil {
		return PropertyResult{}, err
	}

	return PropertyResult{Property: property, EmbeddingStored: embeddingStored}, nil
}

func (s *Service) CreateMany(ctx context.Context, properties []CreateParams) ([]PropertyResult, error) {
	results := make([]PropertyResult, 0, len(properties))
	for _, property := range properties {
		result, err := s.Create(ctx, property)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Service) Get(ctx context.Context, params db.GetPropertyByIDParams) (db.Property, error) {
	return s.queries.GetPropertyByID(ctx, params)
}

func (s *Service) List(ctx context.Context, tenantID pgtype.UUID) ([]db.Property, error) {
	return s.queries.ListProperties(ctx, tenantID)
}

func (s *Service) ListByStatus(ctx context.Context, params db.ListPropertiesByStatusParams) ([]db.Property, error) {
	return s.queries.ListPropertiesByStatus(ctx, params)
}

func (s *Service) ListByType(ctx context.Context, params db.ListPropertiesByTypeParams) ([]db.Property, error) {
	return s.queries.ListPropertiesByType(ctx, params)
}

func (s *Service) Update(ctx context.Context, params UpdateParams) (PropertyResult, error) {
	property, err := s.queries.UpdateProperty(ctx, params.Property)
	if err != nil {
		return PropertyResult{}, err
	}

	if err := s.queries.DeletePropertyEmbedding(ctx, db.DeletePropertyEmbeddingParams{
		PropertyID: property.ID,
		TenantID:   property.TenantID,
	}); err != nil {
		return PropertyResult{}, err
	}

	embeddingStored, err := s.createEmbedding(ctx, property.TenantID, property.ID, params.Content, params.Embedding)
	if err != nil {
		return PropertyResult{}, err
	}

	return PropertyResult{Property: property, EmbeddingStored: embeddingStored}, nil
}

func (s *Service) Delete(ctx context.Context, params db.DeletePropertyParams) error {
	return s.queries.DeleteProperty(ctx, params)
}

func (s *Service) Search(ctx context.Context, params db.SearchPropertyEmbeddingsParams) ([]db.SearchPropertyEmbeddingsRow, error) {
	return s.queries.SearchPropertyEmbeddings(ctx, params)
}

func (s *Service) createEmbedding(ctx context.Context, tenantID pgtype.UUID, propertyID int64, content string, embedding []float32) (bool, error) {
	if len(embedding) == 0 {
		return false, nil
	}

	if _, err := s.queries.CreatePropertyEmbedding(ctx, db.CreatePropertyEmbeddingParams{
		TenantID:   tenantID,
		PropertyID: propertyID,
		Embedding:  pgvector.NewVector(embedding),
		Content:    content,
	}); err != nil {
		return false, err
	}

	return true, nil
}
