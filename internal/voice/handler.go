package voice

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	db "github.com/AGX18/real_estate_crm/internal/db/sqlc"
	"github.com/AGX18/real_estate_crm/internal/httpx"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	service *Service
}

type createCallRequest struct {
	Summary  *agentSummary  `json:"summary"`
	Timeline []timelineItem `json:"timeline"`
	Turns    []turnItem     `json:"turns"`

	Phone        string           `json:"phone"`
	Description  string           `json:"description"`
	Status       db.LeadStatus    `json:"status"`
	Transcript   string           `json:"transcript"`
	Details      string           `json:"details"`
	CallSummary  string           `json:"call_summary"`
	Sentiment    db.CallSentiment `json:"sentiment"`
	Outcome      db.CallOutcome   `json:"outcome"`
	DurationSecs int32            `json:"duration_secs"`
}

type agentSummary struct {
	CallID          string         `json:"call_id"`
	ClientName      string         `json:"client_name"`
	CallOutcome     string         `json:"call_outcome"`
	CallDurationS   int32          `json:"call_duration_s"`
	TotalUserTurns  int32          `json:"total_user_turns"`
	DominantEmotion string         `json:"dominant_emotion"`
	OverallIntent   string         `json:"overall_intent"`
	HesitationRate  string         `json:"hesitation_rate"`
	Qualification   map[string]any `json:"qualification"`
	LLMSummary      string         `json:"llm_summary"`
	Timestamp       string         `json:"timestamp"`
	Phone           *string        `json:"phone"`
	Summary         string         `json:"summary"`
	Classification  string         `json:"classification"`
	CalledAt        string         `json:"called_at"`
}

type timelineItem struct {
	Step            int     `json:"step"`
	Text            string  `json:"text"`
	Hesitation      bool    `json:"hesitation"`
	Intent          string  `json:"intent"`
	Timestamp       string  `json:"timestamp"`
	DominantEmotion string  `json:"dominant_emotion"`
	Confidence      float64 `json:"confidence"`
	Sentiment       string  `json:"sentiment"`
	SentimentScore  float64 `json:"sentiment_score"`
}

type turnItem struct {
	Role      string `json:"role"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateCall(w http.ResponseWriter, r *http.Request) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	body, err := decodeCreateCallRequest(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params := body.toParams(tenantID)
	if params.Phone == "" {
		httpx.WriteError(w, http.StatusBadRequest, "phone is required")
		return
	}

	result, err := h.service.CreateCall(r.Context(), params)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create call")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, result)
}

func decodeCreateCallRequest(r *http.Request) (createCallRequest, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return createCallRequest{}, err
	}
	raw = []byte(strings.ReplaceAll(string(raw), ": NaN", ": null"))
	raw = []byte(strings.ReplaceAll(string(raw), ":NaN", ":null"))

	var body createCallRequest
	if err := json.Unmarshal(raw, &body); err != nil {
		return createCallRequest{}, err
	}
	return body, nil
}

func (r createCallRequest) toParams(tenantID pgtype.UUID) CreateCallParams {
	status := r.Status
	outcome := r.Outcome
	sentiment := r.Sentiment
	durationSecs := r.DurationSecs
	phone := r.Phone
	description := r.Description
	transcript := r.Transcript
	details := r.Details
	summary := r.CallSummary

	if r.Summary != nil {
		if phone == "" && r.Summary.Phone != nil {
			phone = normalizePhone(*r.Summary.Phone)
		}
		if description == "" {
			description = leadDescription(*r.Summary)
		}
		if summary == "" {
			summary = firstNonEmpty(r.Summary.LLMSummary, r.Summary.Summary)
		}
		if status == "" {
			status = normalizeLeadStatus(firstNonEmpty(r.Summary.Classification, r.Summary.CallOutcome, r.Summary.OverallIntent))
		}
		if outcome == "" {
			outcome = normalizeCallOutcome(firstNonEmpty(r.Summary.Classification, r.Summary.CallOutcome, r.Summary.OverallIntent))
		}
		if sentiment == "" {
			sentiment = normalizeCallSentiment(r.Summary.DominantEmotion)
		}
		if durationSecs == 0 {
			durationSecs = r.Summary.CallDurationS
		}
	}

	if phone == "" {
		phone = extractPhone(r.Turns)
	}
	if transcript == "" {
		transcript = buildTranscript(r.Turns)
	}
	if details == "" {
		details = buildDetails(r.Summary, r.Timeline)
	}
	if status == "" {
		status = db.LeadStatusFollowUp
	}
	if outcome == "" {
		outcome = db.CallOutcomeFollowUp
	}

	return CreateCallParams{
		TenantID:     tenantID,
		Phone:        phone,
		Description:  description,
		Status:       status,
		Transcript:   transcript,
		Details:      details,
		Summary:      summary,
		Sentiment:    sentiment,
		Outcome:      outcome,
		DurationSecs: durationSecs,
	}
}

func leadDescription(summary agentSummary) string {
	parts := []string{}
	if summary.ClientName != "" {
		parts = append(parts, "Client: "+summary.ClientName)
	}
	if len(summary.Qualification) > 0 {
		qualification, err := json.Marshal(summary.Qualification)
		if err == nil {
			parts = append(parts, "Qualification: "+string(qualification))
		}
	}
	if summary.OverallIntent != "" {
		parts = append(parts, "Intent: "+summary.OverallIntent)
	}
	if summary.LLMSummary != "" {
		parts = append(parts, summary.LLMSummary)
	}
	if len(parts) == 0 {
		return summary.Summary
	}
	return strings.Join(parts, "\n")
}

func buildTranscript(turns []turnItem) string {
	lines := make([]string, 0, len(turns))
	var previousRole string
	var previousText string
	for _, turn := range turns {
		text := strings.TrimSpace(turn.Text)
		if text == "" || text == "ابدأ المكالمة" {
			continue
		}
		role := firstNonEmpty(turn.Role, "unknown")
		if role == previousRole && text == previousText {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %s", role, text))
		previousRole = role
		previousText = text
	}
	return strings.Join(lines, "\n")
}

func buildDetails(summary *agentSummary, timeline []timelineItem) string {
	parts := []string{}
	if summary != nil {
		if len(summary.Qualification) > 0 {
			qualification, err := json.Marshal(summary.Qualification)
			if err == nil {
				parts = append(parts, "qualification: "+string(qualification))
			}
		}
		parts = append(parts,
			"classification: "+summary.Classification,
			"call_outcome: "+summary.CallOutcome,
			"overall_intent: "+summary.OverallIntent,
			"dominant_emotion: "+summary.DominantEmotion,
		)
	}
	for _, intent := range notableIntents(timeline) {
		parts = append(parts, "notable_intent: "+intent)
	}
	return strings.Join(compact(parts), "\n")
}

func notableIntents(timeline []timelineItem) []string {
	seen := make(map[string]struct{})
	intents := []string{}
	for _, item := range timeline {
		intent := strings.TrimSpace(item.Intent)
		if intent == "" || intent == "general_question" {
			continue
		}
		if _, ok := seen[intent]; ok {
			continue
		}
		seen[intent] = struct{}{}
		intents = append(intents, intent)
	}
	return intents
}

func extractPhone(turns []turnItem) string {
	phonePattern := regexp.MustCompile(`\+?\d[\d\s-]{6,}\d`)
	for _, turn := range turns {
		match := phonePattern.FindString(turn.Text)
		if match == "" {
			continue
		}
		return normalizePhone(match)
	}
	return ""
}

func normalizePhone(value string) string {
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")
	return replacer.Replace(value)
}

func normalizeLeadStatus(value string) db.LeadStatus {
	switch strings.ToLower(value) {
	case "qualified":
		return db.LeadStatusQualified
	case "closed":
		return db.LeadStatusClosed
	case "unqualified", "not_interested", "no_answer":
		return db.LeadStatusUnqualified
	default:
		return db.LeadStatusFollowUp
	}
}

func normalizeCallOutcome(value string) db.CallOutcome {
	switch strings.ToLower(value) {
	case "qualified":
		return db.CallOutcomeQualified
	case "closed":
		return db.CallOutcomeClosed
	case "unqualified", "not_interested":
		return db.CallOutcomeUnqualified
	case "no_answer":
		return db.CallOutcomeNoAnswer
	default:
		return db.CallOutcomeFollowUp
	}
}

func normalizeCallSentiment(value string) db.CallSentiment {
	switch strings.ToLower(value) {
	case "positive", "happy", "interested":
		return db.CallSentimentPositive
	case "negative", "angry", "frustrated":
		return db.CallSentimentNegative
	case "neutral":
		return db.CallSentimentNeutral
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func compact(values []string) []string {
	compacted := values[:0]
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			compacted = append(compacted, value)
		}
	}
	return compacted
}
