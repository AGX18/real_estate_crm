package server

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	db "real_estate_crm/internal/db/sqlc"

	"github.com/go-chi/chi/v5"
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

func testEmbeddingJSON() string {
	values := make([]string, 1536)
	for i := range values {
		values[i] = "0.1"
	}
	return strings.Join(values, ",")
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
