package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	db "real_estate_crm/internal/db/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func newVoiceTestRouter(s *Server) http.Handler {
	r := chi.NewRouter()
	s.registerVoiceRoutes(r)
	return r
}

func TestCreateCallCreatesLead(t *testing.T) {
	mock := &mockQueries{
		lead:           db.Lead{ID: 1, Phone: "+201001234567"},
		call:           db.Call{ID: 1},
		leadByPhoneErr: pgx.ErrNoRows,
	}
	s := newTestServer(mock)
	r := newVoiceTestRouter(s)

	body := bytes.NewBufferString(`{
		"summary":{
			"call_id":"20260607_165802",
			"client_name":"",
			"call_outcome":"callback",
			"call_duration_s":7,
			"total_user_turns":8,
			"dominant_emotion":"neutral",
			"overall_intent":"interest_show",
			"hesitation_rate":"0.0%",
			"qualification":{"area":"التجمع الخامس"},
			"llm_summary":"عميل يبحث عن سكن في التجمع الخامس ويريد التفاصيل على الواتساب.",
			"phone":null,
			"summary":"عميل مهتم بالتجمع الخامس.",
			"classification":"callback",
			"called_at":"2026-06-07T16:58:02.541134"
		},
		"timeline":[
			{"step":1,"text":"هل ممكن التجمع الخامس؟","hesitation":false,"intent":"general_question","dominant_emotion":"neutral","confidence":0.9,"sentiment":"neutral","sentiment_score":0.99}
		],
		"turns":[
			{"role":"user","text":"هل ممكن التجمع الخامس؟","timestamp":"2026-06-07T16:58:02.541284"},
			{"role":"user","text":"تمام، 010-181-335-5540.","timestamp":"2026-06-07T16:58:02.541339"}
		]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/calls", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected %d got %d", http.StatusCreated, w.Code)
	}
}

func TestCreateCallUpdatesExistingLead(t *testing.T) {
	mock := &mockQueries{
		lead: db.Lead{ID: 1, Phone: "+201001234567"},
		call: db.Call{ID: 1},
	}
	s := newTestServer(mock)
	r := newVoiceTestRouter(s)

	body := bytes.NewBufferString(`{
		"summary":{
			"call_id":"20260607_165802",
			"call_outcome":"qualified",
			"call_duration_s":120,
			"dominant_emotion":"positive",
			"overall_intent":"interest_show",
			"qualification":{"area":"New Cairo","budget":"6M"},
			"llm_summary":"Qualified buyer wants New Cairo apartment.",
			"phone":"+201001234567",
			"classification":"qualified"
		},
		"timeline":[],
		"turns":[
			{"role":"agent","text":"How can I help?"},
			{"role":"user","text":"I want an apartment in New Cairo."}
		]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/tenants/"+testTenantID+"/calls", body)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected %d got %d", http.StatusCreated, w.Code)
	}
}
