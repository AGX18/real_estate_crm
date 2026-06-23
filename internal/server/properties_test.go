package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"
	"github.com/AGX18/real_estate_crm/internal/properties"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func newPropertyTestRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	s.registerPropertyRoutes(r)
	return r
}

func TestCreateProperty(t *testing.T) {
	mock := &mockQueries{property: db.Property{ID: 1, Bedrooms: 2, Bathrooms: 2}}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	body := bytes.NewBufferString(`{
		"description":"Sea view apartment",
		"price":"2500000.00",
		"location":"North Coast",
		"area_sqm":"120.50",
		"type":"شقة",
		"city":"Marina",
		"governorate":"Matrouh",
		"bedrooms":2,
		"bathrooms":2,
		"status":"available"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected %d got %d", http.StatusCreated, w.Code)
	}
}

func TestCreatePropertyGeneratesEmbedding(t *testing.T) {
	tenantID := mustParseUUID(t, testTenantID)
	mock := &mockQueries{property: db.Property{ID: 1, TenantID: tenantID, Bedrooms: 2, Bathrooms: 2}}
	s := newPropertyEmbeddingTestServer(mock, &fakeEmbedder{embedding: testEmbedding(0.25)})
	r := newPropertyTestRouter(s)

	body := bytes.NewBufferString(`{
		"description":"Sea view apartment",
		"price":"2500000.00",
		"location":"North Coast",
		"area_sqm":"120.50",
		"type":"شقة",
		"city":"Marina",
		"governorate":"Matrouh",
		"bedrooms":2,
		"bathrooms":2,
		"status":"available"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d got %d body %s", http.StatusCreated, w.Code, w.Body.String())
	}
	if mock.createPropertyEmbeddingCalls != 1 {
		t.Fatalf("expected one embedding write got %d", mock.createPropertyEmbeddingCalls)
	}
	if got := len(mock.createPropertyEmbeddingArg.Embedding.Slice()); got != 1536 {
		t.Fatalf("expected 1536 dimensions got %d", got)
	}
	if !strings.Contains(mock.createPropertyEmbeddingArg.Content, "Sea view apartment") {
		t.Fatalf("expected embedding content to include property description, got %q", mock.createPropertyEmbeddingArg.Content)
	}
}

func TestCreatePropertyEmbeddingFailureDoesNotCreateProperty(t *testing.T) {
	mock := &mockQueries{}
	s := newPropertyEmbeddingTestServer(mock, &fakeEmbedder{err: errors.New("provider down")})
	r := newPropertyTestRouter(s)

	body := bytes.NewBufferString(`{
		"description":"Sea view apartment",
		"price":"2500000.00",
		"location":"North Coast",
		"area_sqm":"120.50",
		"type":"شقة",
		"city":"Marina",
		"governorate":"Matrouh",
		"bedrooms":2,
		"bathrooms":2,
		"status":"available"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d got %d body %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
	if mock.createPropertyArg.AreaSqm.Valid {
		t.Fatal("expected property not to be created when embedding generation fails")
	}
	if mock.createPropertyEmbeddingCalls != 0 {
		t.Fatalf("expected no embedding writes got %d", mock.createPropertyEmbeddingCalls)
	}
}

func TestCreatePropertiesBulk(t *testing.T) {
	mock := &mockQueries{property: db.Property{ID: 1, Bedrooms: 2, Bathrooms: 2}}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	body := bytes.NewBufferString(`[
		{
			"description":"Apartment one",
			"price":"2500000.00",
			"location":"New Cairo",
			"area_sqm":"120",
			"type":"شقة",
			"city":"Cairo",
			"governorate":"Cairo",
			"bedrooms":2,
			"bathrooms":2,
			"status":"available"
		}
	]`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties/bulk", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected %d got %d", http.StatusCreated, w.Code)
	}
}

func TestImportPropertiesArrayFile(t *testing.T) {
	mock := &mockQueries{property: db.Property{ID: 1, Bedrooms: 2, Bathrooms: 2}}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	body, contentType := propertiesImportBody(t, "properties.json", `[
		{
			"description":"Apartment one",
			"price":"2500000.00",
			"location":"New Cairo",
			"area_sqm":"120",
			"type":"شقة",
			"city":"Cairo",
			"governorate":"Cairo",
			"bedrooms":2,
			"bathrooms":2,
			"status":"available"
		}
	]`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties/import", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected %d got %d", http.StatusCreated, w.Code)
	}
}

func TestImportPropertiesSkipsDuplicateFile(t *testing.T) {
	mock := &mockQueries{property: db.Property{ID: 1, Bedrooms: 2, Bathrooms: 2}}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	content := `[
		{
			"description":"Apartment one",
			"price":"2500000.00",
			"location":"New Cairo",
			"area_sqm":"120",
			"type":"شقة",
			"city":"Cairo",
			"governorate":"Cairo",
			"bedrooms":2,
			"bathrooms":2,
			"status":"available"
		}
	]`
	body, contentType := propertiesImportBody(t, "properties.json", content)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties/import", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d got %d", http.StatusCreated, w.Code)
	}

	body, contentType = propertiesImportBody(t, "properties.json", content)
	req = httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties/import", body)
	req.Header.Set("Content-Type", contentType)
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected duplicate import to return %d got %d", http.StatusOK, w.Code)
	}
	if mock.createPropertyCalls != 1 {
		t.Fatalf("expected duplicate import to create 1 property got %d", mock.createPropertyCalls)
	}
}

func TestImportPropertiesWrappedFile(t *testing.T) {
	mock := &mockQueries{property: db.Property{ID: 1, Bedrooms: 2, Bathrooms: 2}}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	body, contentType := propertiesImportBody(t, "properties.json", `{
		"properties": [
			{
				"description":"Apartment one",
				"price":"2500000.00",
				"location":"New Cairo",
				"area_sqm":"120",
				"type":"شقة",
				"city":"Cairo",
				"governorate":"Cairo",
				"bedrooms":2,
				"bathrooms":2,
				"status":"available"
			}
		]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties/import", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected %d got %d", http.StatusCreated, w.Code)
	}
}

func TestImportPropertiesAcceptsTestDataShape(t *testing.T) {
	mock := &mockQueries{property: db.Property{ID: 1, Bedrooms: 4, Bathrooms: 4}}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	body, contentType := propertiesImportBody(t, "test_data.json", `[
		{
			"description":"خصم 20% على فيلا توين هاوس متشطبة",
			"price":15000000,
			"area_sqm":212,
			"type":"فیلا",
			"project":"كومباوند ازار",
			"city":"التجمع الخامس",
			"Governorate":"القاهرة",
			"bedrooms":4,
			"bathrooms":4
		}
	]`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties/import", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d got %d body %s", http.StatusCreated, w.Code, w.Body.String())
	}
	if !mock.createPropertyArg.Price.Valid {
		t.Fatal("expected numeric price to be accepted")
	}
	if !mock.createPropertyArg.AreaSqm.Valid {
		t.Fatal("expected numeric area_sqm to be accepted")
	}
	if mock.createPropertyArg.Type.PropertyType != db.PropertyTypeValue5 {
		t.Fatalf("expected Persian Yeh villa spelling to normalize to %q got %q", db.PropertyTypeValue5, mock.createPropertyArg.Type.PropertyType)
	}
	if !mock.createPropertyArg.Location.Valid || mock.createPropertyArg.Location.String != "كومباوند ازار" {
		t.Fatalf("expected project to map to location got %+v", mock.createPropertyArg.Location)
	}
	if !mock.createPropertyArg.Governorate.Valid || mock.createPropertyArg.Governorate.String != "القاهرة" {
		t.Fatalf("expected capitalized Governorate to map to governorate got %+v", mock.createPropertyArg.Governorate)
	}
}

func TestImportPropertiesAcceptsTestData2Shape(t *testing.T) {
	mock := &mockQueries{property: db.Property{ID: 1, Bedrooms: 2, Bathrooms: 2}}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	body, contentType := propertiesImportBody(t, "test_data_2.json", `[
		{
			"description":"ارخص سعر في الماركت موقع متميز شقة غرفتين للبيع في كمباوند ماونتن فيو هايد بارك التجمع الخامس القاهرة الجديدة",
			"price":7400000,
			"area_sqm":141,
			"type":"شقة",
			"project":"كومباوند ماونتن فيو هايد بارك",
			"city":"التجمع الخامس",
			"Governorate":"القاهرة",
			"bedrooms":2,
			"bathrooms":3
		},
		{
			"description":"شالية صف اول علي الممشي السياحي امام ابراج العلمين مباشرة متشطب جاهز للاستلام بخصم الكاش 50% Down Town",
			"price":6200000,
			"area_sqm":164,
			"type":"شاليه",
			"project":"داون تاون",
			"city":"العلمين",
			"Governorate":"مطروح",
			"bedrooms":2,
			"bathrooms":2
		},
		{
			"description":"للبيع بمدينة نور فيلا استاند الون نموذج B يسعر لقطه",
			"price":22546000,
			"area_sqm":488,
			"type":"فیلا",
			"project":"مدينة نور",
			"city":"العاصمة الإدارية الجديدة",
			"Governorate":"القاهرة",
			"bedrooms":5,
			"bathrooms":5
		},
		{
			"description":"شقه للبيع الابراهيميه خطوات من شارع ابوقير",
			"price":2250000,
			"area_sqm":120,
			"type":"شقة",
			"project":"الابراهيمية",
			"Governorate":"الإسكندرية",
			"bedrooms":2,
			"bathrooms":1
		}
	]`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties/import", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected %d got %d body %s", http.StatusCreated, w.Code, w.Body.String())
	}
	if mock.createPropertyCalls != 4 {
		t.Fatalf("expected 4 properties to be created got %d", mock.createPropertyCalls)
	}
	if !mock.createPropertyArg.Governorate.Valid || mock.createPropertyArg.Governorate.String != "الإسكندرية" {
		t.Fatalf("expected final governorate to import got %+v", mock.createPropertyArg.Governorate)
	}
}

func TestListProperties(t *testing.T) {
	mock := &mockQueries{
		properties: []db.Property{
			{ID: 1, Bedrooms: 2, Bathrooms: 2},
			{ID: 2, Bedrooms: 3, Bathrooms: 2},
		},
	}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	req := httptest.NewRequest(http.MethodGet, "/tenants/"+testTenantID+"/properties", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func propertiesImportBody(t *testing.T, filename string, content string) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	file, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := file.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}
	return body, writer.FormDataContentType()
}

func TestSearchProperties(t *testing.T) {
	mock := &mockQueries{
		propertySearchResults: []db.SearchPropertyEmbeddingsRow{
			{PropertyID: 1, Content: "Sea view apartment"},
		},
	}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	body := bytes.NewBufferString(fmt.Sprintf(`{"embedding":[%s],"limit":5}`, testEmbeddingJSON()))
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties/search", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestSearchPropertiesWithQueryGeneratesEmbedding(t *testing.T) {
	mock := &mockQueries{
		propertySearchResults: []db.SearchPropertyEmbeddingsRow{
			{PropertyID: 1, Content: "Sea view apartment"},
		},
	}
	embedder := &fakeEmbedder{embedding: testEmbedding(0.5)}
	s := newPropertyEmbeddingTestServer(mock, embedder)
	r := newPropertyTestRouter(s)

	body := bytes.NewBufferString(`{"query":"sea view apartment in north coast","limit":3}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties/search", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d got %d body %s", http.StatusOK, w.Code, w.Body.String())
	}
	if embedder.lastText != "sea view apartment in north coast" {
		t.Fatalf("expected query to be embedded, got %q", embedder.lastText)
	}
	if got := len(mock.searchPropertyEmbeddingsArg.Embedding.Slice()); got != 1536 {
		t.Fatalf("expected 1536 dimensions got %d", got)
	}
	if mock.searchPropertyEmbeddingsArg.Limit != 3 {
		t.Fatalf("expected search limit 3 got %d", mock.searchPropertyEmbeddingsArg.Limit)
	}
}

func TestSearchPropertiesQueryEmbeddingFailureDoesNotSearch(t *testing.T) {
	mock := &mockQueries{}
	s := newPropertyEmbeddingTestServer(mock, &fakeEmbedder{err: errors.New("provider down")})
	r := newPropertyTestRouter(s)

	body := bytes.NewBufferString(`{"query":"sea view apartment","limit":3}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties/search", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d got %d body %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
	if len(mock.searchPropertyEmbeddingsArg.Embedding.Slice()) != 0 {
		t.Fatal("expected vector search not to run when query embedding fails")
	}
}

func TestSearchPropertiesRejectsWrongEmbeddingDimensions(t *testing.T) {
	mock := &mockQueries{}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	body := bytes.NewBufferString(`{"embedding":[0.1,0.2],"limit":5}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/properties/search", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d body %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func testEmbeddingJSON() string {
	values := make([]string, 1536)
	for i := range values {
		values[i] = "0.1"
	}
	return strings.Join(values, ",")
}

func testEmbedding(value float32) []float32 {
	values := make([]float32, 1536)
	for i := range values {
		values[i] = value
	}
	return values
}

func TestUpdateProperty(t *testing.T) {
	mock := &mockQueries{property: db.Property{ID: 1, Bedrooms: 3, Bathrooms: 2}}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	body := bytes.NewBufferString(`{
		"description":"Updated apartment",
		"price":"3000000.00",
		"location":"New Cairo",
		"area_sqm":"140",
		"type":"شقة",
		"city":"Cairo",
		"governorate":"Cairo",
		"bedrooms":3,
		"bathrooms":2,
		"status":"available"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/tenants/"+testTenantID+"/properties/1", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteProperty(t *testing.T) {
	mock := &mockQueries{}
	s := newTestServer(mock)
	r := newPropertyTestRouter(s)

	req := httptest.NewRequest(http.MethodDelete, "/tenants/"+testTenantID+"/properties/1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, w.Code)
	}
}

type fakeEmbedder struct {
	embedding []float32
	err       error
	lastText  string
}

func (f *fakeEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	f.lastText = text
	return f.embedding, f.err
}

func newPropertyEmbeddingTestServer(q *mockQueries, embedder *fakeEmbedder) *Server {
	s := newTestServer(q)
	propertyService := properties.NewServiceWithEmbedder(q, embedder)
	s.propertyService = propertyService
	s.propertyHandler = properties.NewHandler(propertyService)
	return s
}

func mustParseUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()

	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		t.Fatalf("failed to parse uuid: %v", err)
	}
	return id
}
