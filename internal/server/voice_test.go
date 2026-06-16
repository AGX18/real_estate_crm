package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"

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
			{"step":1,"text":"هل ممكن التجمع الخامس؟","hesitation":false,"intent":"general_question","dominant_emotion":"neutral","confidence":NaN,"sentiment":"neutral","sentiment_score":NaN}
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
		t.Fatalf("expected %d got %d body %s", http.StatusCreated, w.Code, w.Body.String())
	}
	if mock.createLeadArg.Phone != "0101813355540" {
		t.Fatalf("expected extracted normalized phone got %q", mock.createLeadArg.Phone)
	}
	if mock.createLeadArg.Status.LeadStatus != db.LeadStatusFollowUp {
		t.Fatalf("expected callback to map to follow_up got %q", mock.createLeadArg.Status.LeadStatus)
	}
	if !mock.createLeadArg.Description.Valid || !bytes.Contains([]byte(mock.createLeadArg.Description.String), []byte("التجمع الخامس")) {
		t.Fatalf("expected lead description to include qualification/summary, got %q", mock.createLeadArg.Description.String)
	}
	if mock.createCallArg.Outcome.CallOutcome != db.CallOutcomeFollowUp {
		t.Fatalf("expected call outcome follow_up got %q", mock.createCallArg.Outcome.CallOutcome)
	}
	if !mock.createCallArg.Transcript.Valid || !bytes.Contains([]byte(mock.createCallArg.Transcript.String), []byte("010-181-335-5540")) {
		t.Fatalf("expected transcript to include useful turn text, got %q", mock.createCallArg.Transcript.String)
	}
	if bytes.Contains([]byte(mock.createCallArg.Details.String), []byte("confidence")) {
		t.Fatalf("expected details to ignore confidence fields, got %q", mock.createCallArg.Details.String)
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
		t.Fatalf("expected %d got %d body %s", http.StatusCreated, w.Code, w.Body.String())
	}
	if mock.updateLeadArg.Phone != "+201001234567" {
		t.Fatalf("expected existing lead to be updated with summary phone, got %q", mock.updateLeadArg.Phone)
	}
	if mock.updateLeadArg.Status.LeadStatus != db.LeadStatusQualified {
		t.Fatalf("expected qualified status got %q", mock.updateLeadArg.Status.LeadStatus)
	}
	if mock.createCallArg.Outcome.CallOutcome != db.CallOutcomeQualified {
		t.Fatalf("expected qualified outcome got %q", mock.createCallArg.Outcome.CallOutcome)
	}
	if mock.createCallArg.DurationSecs.Int32 != 120 || !mock.createCallArg.DurationSecs.Valid {
		t.Fatalf("expected duration 120 got %+v", mock.createCallArg.DurationSecs)
	}
}

func TestListCalls(t *testing.T) {
	mock := &mockQueries{
		calls: []db.Call{
			{ID: 1},
			{ID: 2},
		},
	}
	s := newTestServer(mock)
	r := newVoiceTestRouter(s)

	req := httptest.NewRequest(http.MethodGet, "/tenants/"+testTenantID+"/calls", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected %d got %d body %s", http.StatusOK, w.Code, w.Body.String())
	}
}
