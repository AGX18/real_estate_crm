package properties

import (
	"context"
	"strings"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"
	"github.com/AGX18/real_estate_crm/internal/embeddings"
	"github.com/AGX18/real_estate_crm/internal/store"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
)

type Service struct {
	queries  db.Querier
	store    txRunner
	embedder embeddings.Embedder
}

type txRunner interface {
	WithTx(ctx context.Context, fn func(q db.Querier) error) error
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
	return &Service{queries: queries, embedder: embeddings.NoopEmbedder{}}
}

func NewServiceWithEmbedder(queries db.Querier, embedder embeddings.Embedder) *Service {
	if embedder == nil {
		embedder = embeddings.NoopEmbedder{}
	}
	return &Service{queries: queries, embedder: embedder}
}

func NewServiceWithStore(store *store.Store, embedder embeddings.Embedder) *Service {
	if embedder == nil {
		embedder = embeddings.NoopEmbedder{}
	}
	return &Service{queries: store.Queries(), store: store, embedder: embedder}
}

func (s *Service) Create(ctx context.Context, params CreateParams) (PropertyResult, error) {
	params, err := s.withEmbedding(ctx, params)
	if err != nil {
		return PropertyResult{}, err
	}
	if s.store == nil {
		return s.create(ctx, s.queries, params)
	}

	var result PropertyResult
	err = s.store.WithTx(ctx, func(q db.Querier) error {
		var err error
		result, err = s.create(ctx, q, params)
		return err
	})
	return result, err
}

func (s *Service) create(ctx context.Context, q db.Querier, params CreateParams) (PropertyResult, error) {
	property, err := q.CreateProperty(ctx, params.Property)
	if err != nil {
		return PropertyResult{}, err
	}

	embeddingStored, err := s.createEmbedding(ctx, q, property.TenantID, property.ID, params.Content, params.Embedding)
	if err != nil {
		return PropertyResult{}, err
	}

	return PropertyResult{Property: property, EmbeddingStored: embeddingStored}, nil
}

func (s *Service) CreateMany(ctx context.Context, properties []CreateParams) ([]PropertyResult, error) {
	prepared := make([]CreateParams, 0, len(properties))
	for _, property := range properties {
		nextProperty, err := s.withEmbedding(ctx, property)
		if err != nil {
			return nil, err
		}
		prepared = append(prepared, nextProperty)
	}

	if s.store != nil {
		results := make([]PropertyResult, 0, len(prepared))
		err := s.store.WithTx(ctx, func(q db.Querier) error {
			for _, property := range prepared {
				result, err := s.create(ctx, q, property)
				if err != nil {
					return err
				}
				results = append(results, result)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return results, nil
	}

	results := make([]PropertyResult, 0, len(properties))
	for _, property := range prepared {
		result, err := s.create(ctx, s.queries, property)
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
	params, err := s.withUpdateEmbedding(ctx, params)
	if err != nil {
		return PropertyResult{}, err
	}
	if s.store == nil {
		return s.update(ctx, s.queries, params)
	}

	var result PropertyResult
	err = s.store.WithTx(ctx, func(q db.Querier) error {
		var err error
		result, err = s.update(ctx, q, params)
		return err
	})
	return result, err
}

func (s *Service) update(ctx context.Context, q db.Querier, params UpdateParams) (PropertyResult, error) {
	property, err := q.UpdateProperty(ctx, params.Property)
	if err != nil {
		return PropertyResult{}, err
	}

	if err := q.DeletePropertyEmbedding(ctx, db.DeletePropertyEmbeddingParams{
		PropertyID: property.ID,
		TenantID:   property.TenantID,
	}); err != nil {
		return PropertyResult{}, err
	}

	embeddingStored, err := s.createEmbedding(ctx, q, property.TenantID, property.ID, params.Content, params.Embedding)
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

func (s *Service) Embed(ctx context.Context, text string) ([]float32, error) {
	return s.embedder.Embed(ctx, text)
}

func (s *Service) createEmbedding(ctx context.Context, q db.Querier, tenantID pgtype.UUID, propertyID int64, content string, embedding []float32) (bool, error) {
	if len(embedding) == 0 {
		return false, nil
	}

	if _, err := q.CreatePropertyEmbedding(ctx, db.CreatePropertyEmbeddingParams{
		TenantID:   tenantID,
		PropertyID: propertyID,
		Embedding:  pgvector.NewVector(embedding),
		Content:    content,
	}); err != nil {
		return false, err
	}

	return true, nil
}

func (s *Service) withEmbedding(ctx context.Context, params CreateParams) (CreateParams, error) {
	if len(params.Embedding) > 0 || strings.TrimSpace(params.Content) == "" {
		return params, nil
	}
	embedding, err := s.embedder.Embed(ctx, params.Content)
	if err != nil {
		return CreateParams{}, err
	}
	params.Embedding = embedding
	return params, nil
}

func (s *Service) withUpdateEmbedding(ctx context.Context, params UpdateParams) (UpdateParams, error) {
	if len(params.Embedding) > 0 || strings.TrimSpace(params.Content) == "" {
		return params, nil
	}
	embedding, err := s.embedder.Embed(ctx, params.Content)
	if err != nil {
		return UpdateParams{}, err
	}
	params.Embedding = embedding
	return params, nil
}
