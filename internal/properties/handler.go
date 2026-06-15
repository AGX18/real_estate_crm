package properties

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"
	"github.com/AGX18/real_estate_crm/internal/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
)

const (
	embeddingDimensions = 1536
	maxImportFileSize   = 10 << 20
)

type Handler struct {
	service *Service
}

type propertyRequest struct {
	Description string            `json:"description"`
	Price       string            `json:"price"`
	Location    string            `json:"location"`
	AreaSqm     string            `json:"area_sqm"`
	Type        db.PropertyType   `json:"type"`
	City        string            `json:"city"`
	Governorate string            `json:"governorate"`
	Bedrooms    int32             `json:"bedrooms"`
	Bathrooms   int32             `json:"bathrooms"`
	Status      db.PropertyStatus `json:"status"`
	Embedding   []float32         `json:"embedding"`
}

type searchRequest struct {
	Embedding []float32 `json:"embedding"`
	Query     string    `json:"query"`
	Limit     int32     `json:"limit"`
}

type importRequest struct {
	Properties []propertyRequest `json:"properties"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return
	}

	var body propertyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	params, content, err := createParams(tenantID, body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateEmbedding(body.Embedding); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.Create(r.Context(), CreateParams{
		Property:  params,
		Content:   content,
		Embedding: body.Embedding,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create property")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, result)
}

func (h *Handler) CreateMany(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return
	}

	var body []propertyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(body) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "properties are required")
		return
	}

	h.createMany(w, r, tenantID, body)
}

func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxImportFileSize)
	if err := r.ParseMultipartForm(maxImportFileSize); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxImportFileSize+1))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "failed to read file")
		return
	}
	if len(data) > maxImportFileSize {
		httpx.WriteError(w, http.StatusBadRequest, "file is too large")
		return
	}

	properties, err := decodeImportProperties(data)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid properties json")
		return
	}
	if len(properties) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "properties are required")
		return
	}

	h.createMany(w, r, tenantID, properties)
}

func (h *Handler) createMany(w http.ResponseWriter, r *http.Request, tenantID pgtype.UUID, body []propertyRequest) {
	properties := make([]CreateParams, 0, len(body))
	for _, item := range body {
		params, content, err := createParams(tenantID, item)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := validateEmbedding(item.Embedding); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		properties = append(properties, CreateParams{
			Property:  params,
			Content:   content,
			Embedding: item.Embedding,
		})
	}

	results, err := h.service.CreateMany(r.Context(), properties)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create properties")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, results)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, propertyID, ok := parseTenantAndPropertyID(w, r)
	if !ok {
		return
	}

	property, err := h.service.Get(r.Context(), db.GetPropertyByIDParams{
		ID:       propertyID,
		TenantID: tenantID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "property not found")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, property)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return
	}

	status := db.PropertyStatus(r.URL.Query().Get("status"))
	if status != "" {
		properties, err := h.service.ListByStatus(r.Context(), db.ListPropertiesByStatusParams{
			TenantID: tenantID,
			Status:   propertyStatusParam(status),
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch properties")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, properties)
		return
	}

	propertyType := db.PropertyType(r.URL.Query().Get("type"))
	if propertyType != "" {
		properties, err := h.service.ListByType(r.Context(), db.ListPropertiesByTypeParams{
			TenantID: tenantID,
			Type:     propertyTypeParam(propertyType),
		})
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch properties")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, properties)
		return
	}

	properties, err := h.service.List(r.Context(), tenantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch properties")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, properties)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID, propertyID, ok := parseTenantAndPropertyID(w, r)
	if !ok {
		return
	}

	var body propertyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	params, content, err := updateParams(tenantID, propertyID, body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateEmbedding(body.Embedding); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.Update(r.Context(), UpdateParams{
		Property:  params,
		Content:   content,
		Embedding: body.Embedding,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update property")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, propertyID, ok := parseTenantAndPropertyID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), db.DeletePropertyParams{
		ID:       propertyID,
		TenantID: tenantID,
	}); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete property")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "property deleted"})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return
	}

	var body searchRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(body.Embedding) == 0 && strings.TrimSpace(body.Query) != "" {
		embedding, err := h.service.Embed(r.Context(), body.Query)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to embed query")
			return
		}
		body.Embedding = embedding
	}
	if len(body.Embedding) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "embedding or query is required")
		return
	}
	if err := validateEmbedding(body.Embedding); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if body.Limit <= 0 {
		body.Limit = 10
	}

	results, err := h.service.Search(r.Context(), db.SearchPropertyEmbeddingsParams{
		TenantID:  tenantID,
		Embedding: pgvector.NewVector(body.Embedding),
		Limit:     body.Limit,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to search properties")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, results)
}

func createParams(tenantID pgtype.UUID, body propertyRequest) (db.CreatePropertyParams, string, error) {
	if body.AreaSqm == "" {
		return db.CreatePropertyParams{}, "", fmt.Errorf("area_sqm is required")
	}
	if body.Bedrooms < 0 || body.Bathrooms < 0 {
		return db.CreatePropertyParams{}, "", fmt.Errorf("bedrooms and bathrooms cannot be negative")
	}
	if body.Type == "" {
		body.Type = db.PropertyTypeValue1
	}
	if body.Status == "" {
		body.Status = db.PropertyStatusAvailable
	}

	price, err := numericParam(body.Price)
	if err != nil {
		return db.CreatePropertyParams{}, "", fmt.Errorf("invalid price")
	}
	areaSqm, err := requiredNumericParam(body.AreaSqm)
	if err != nil {
		return db.CreatePropertyParams{}, "", fmt.Errorf("invalid area_sqm")
	}

	params := db.CreatePropertyParams{
		TenantID:    tenantID,
		Description: textParam(body.Description),
		Price:       price,
		Location:    textParam(body.Location),
		AreaSqm:     areaSqm,
		Type:        propertyTypeParam(body.Type),
		City:        textParam(body.City),
		Governorate: textParam(body.Governorate),
		Bedrooms:    body.Bedrooms,
		Bathrooms:   body.Bathrooms,
		Status:      propertyStatusParam(body.Status),
	}
	return params, embeddingContent(body), nil
}

func updateParams(tenantID pgtype.UUID, propertyID int64, body propertyRequest) (db.UpdatePropertyParams, string, error) {
	create, content, err := createParams(tenantID, body)
	if err != nil {
		return db.UpdatePropertyParams{}, "", err
	}

	return db.UpdatePropertyParams{
		ID:          propertyID,
		Description: create.Description,
		Price:       create.Price,
		Location:    create.Location,
		AreaSqm:     create.AreaSqm,
		Type:        create.Type,
		City:        create.City,
		Governorate: create.Governorate,
		Bedrooms:    create.Bedrooms,
		Bathrooms:   create.Bathrooms,
		Status:      create.Status,
		TenantID:    tenantID,
	}, content, nil
}

func parseTenantID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return pgtype.UUID{}, false
	}
	return tenantID, true
}

func parseTenantAndPropertyID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, int64, bool) {
	tenantID, ok := parseTenantID(w, r)
	if !ok {
		return pgtype.UUID{}, 0, false
	}

	propertyID, err := httpx.ParseInt64(chi.URLParam(r, "property_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid property id")
		return pgtype.UUID{}, 0, false
	}

	return tenantID, propertyID, true
}

func textParam(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func numericParam(value string) (pgtype.Numeric, error) {
	var numeric pgtype.Numeric
	if value == "" {
		return numeric, nil
	}
	if err := numeric.Scan(value); err != nil {
		return pgtype.Numeric{}, err
	}
	return numeric, nil
}

func requiredNumericParam(value string) (pgtype.Numeric, error) {
	numeric, err := numericParam(value)
	if err != nil {
		return pgtype.Numeric{}, err
	}
	numeric.Valid = true
	return numeric, nil
}

func propertyTypeParam(value db.PropertyType) db.NullPropertyType {
	return db.NullPropertyType{PropertyType: value, Valid: value != ""}
}

func propertyStatusParam(value db.PropertyStatus) db.NullPropertyStatus {
	return db.NullPropertyStatus{PropertyStatus: value, Valid: value != ""}
}

func embeddingContent(property propertyRequest) string {
	parts := []string{
		property.Description,
		property.Location,
		property.City,
		property.Governorate,
		string(property.Type),
		fmt.Sprintf("%d bedrooms", property.Bedrooms),
		fmt.Sprintf("%d bathrooms", property.Bathrooms),
	}
	if property.Price != "" {
		parts = append(parts, "price "+property.Price)
	}
	if property.AreaSqm != "" {
		parts = append(parts, "area "+property.AreaSqm+" sqm")
	}
	return strings.Join(parts, "\n")
}

func validateEmbedding(embedding []float32) error {
	if len(embedding) == 0 {
		return nil
	}
	if len(embedding) != embeddingDimensions {
		return fmt.Errorf("embedding must contain %d dimensions", embeddingDimensions)
	}
	return nil
}

func decodeImportProperties(data []byte) ([]propertyRequest, error) {
	var properties []propertyRequest
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&properties); err == nil {
		return properties, nil
	}

	var wrapper importRequest
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&wrapper); err != nil {
		return nil, err
	}
	return wrapper.Properties, nil
}
