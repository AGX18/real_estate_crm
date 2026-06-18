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
	Summary     *agentSummary  `json:"summary"`
	SummaryText string         `json:"-"`
	Timeline    []timelineItem `json:"timeline"`
	Turns       []turnItem     `json:"turns"`

	Phone        string           `json:"phone"`
	PhoneNumber  string           `json:"phone_number"`
	Description  string           `json:"description"`
	Status       db.LeadStatus    `json:"status"`
	LeadStatus   db.LeadStatus    `json:"lead_status"`
	Transcript   string           `json:"transcript"`
	Details      string           `json:"details"`
	CallSummary  string           `json:"call_summary"`
	Sentiment    db.CallSentiment `json:"sentiment"`
	Outcome      db.CallOutcome   `json:"outcome"`
	DurationSecs int32            `json:"duration_secs"`
}

func (r *createCallRequest) UnmarshalJSON(data []byte) error {
	type requestAlias createCallRequest
	var body struct {
		requestAlias
		Summary json.RawMessage `json:"summary"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		return err
	}

	*r = createCallRequest(body.requestAlias)
	if len(body.Summary) == 0 || string(body.Summary) == "null" {
		return nil
	}

	var summaryText string
	if err := json.Unmarshal(body.Summary, &summaryText); err == nil {
		r.SummaryText = summaryText
		return nil
	}

	var summary agentSummary
	if err := json.Unmarshal(body.Summary, &summary); err != nil {
		return err
	}
	r.Summary = &summary
	return nil
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

func (h *Handler) ListCalls(w http.ResponseWriter, r *http.Request) {
	tenantID, err := httpx.ParseUUID(chi.URLParam(r, "tenant_id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid tenant id")
		return
	}

	calls, err := h.service.ListCalls(r.Context(), tenantID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to fetch calls")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, calls)
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
	status := firstLeadStatus(r.Status, r.LeadStatus)
	outcome := r.Outcome
	sentiment := r.Sentiment
	durationSecs := r.DurationSecs
	phone := normalizePhone(firstNonEmpty(r.Phone, r.PhoneNumber))
	description := r.Description
	transcript := r.Transcript
	details := r.Details
	summary := firstNonEmpty(r.CallSummary, r.SummaryText)
	var structured structuredSummary

	if r.Summary != nil {
		structured = parseStructuredSummary(r.Summary.Summary)
	}
	if details != "" {
		detailsStructured := parseStructuredSummary(details)
		if len(detailsStructured.Fields) > 0 {
			structured = detailsStructured
		}
	}
	if len(structured.Fields) == 0 && summary != "" {
		structured = parseStructuredSummary(summary)
	}
	if phone == "" {
		phone = normalizePhone(firstNonEmpty(structured.Fields["phone number"], structured.Fields["phone"]))
	}
	if status == "" {
		if value := firstNonEmpty(string(r.LeadStatus), string(r.Outcome), structured.Fields["call outcome"]); value != "" {
			status = normalizeLeadStatus(value)
		}
	}
	if outcome == "" {
		if value := firstNonEmpty(string(r.Status), string(r.LeadStatus), structured.Fields["call outcome"]); value != "" {
			outcome = normalizeCallOutcome(value)
		}
	}
	if sentiment == "" {
		if value := structured.Fields["sentiment"]; value != "" {
			sentiment = normalizeCallSentiment(value)
		}
	}
	if description == "" && len(structured.Fields) > 0 {
		description = structuredLeadDescription(structured)
	}
	if len(structured.Fields) > 0 {
		summary = conciseCallSummary(phone, outcome)
		details = mergeDetails(structuredDetails(structured), detailsWithoutAssistantPrompts(details))
	}

	if r.Summary != nil {
		if phone == "" && r.Summary.Phone != nil {
			phone = normalizePhone(*r.Summary.Phone)
		}
		if phone == "" {
			phone = normalizePhone(firstNonEmpty(structured.Fields["phone number"], structured.Fields["phone"]))
		}
		if description == "" {
			description = leadDescription(*r.Summary)
			if isStructuredSummary(description) {
				description = structuredLeadDescription(structured)
			}
		}
		if summary == "" {
			summary = firstNonEmpty(r.Summary.LLMSummary, structuredCallSummary(structured), r.Summary.Summary)
		}
		if status == "" {
			status = normalizeLeadStatus(firstNonEmpty(r.Summary.Classification, r.Summary.CallOutcome, structured.Fields["call outcome"], r.Summary.OverallIntent))
		}
		if outcome == "" {
			outcome = normalizeCallOutcome(firstNonEmpty(r.Summary.Classification, r.Summary.CallOutcome, structured.Fields["call outcome"], r.Summary.OverallIntent))
		}
		if sentiment == "" {
			sentiment = normalizeCallSentiment(firstNonEmpty(r.Summary.DominantEmotion, structured.Fields["sentiment"]))
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
		details = mergeDetails(structuredDetails(structured), buildDetails(r.Summary, r.Timeline))
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

type structuredSummary struct {
	Fields map[string]string
}

func parseStructuredSummary(value string) structuredSummary {
	fields := map[string]string{}
	value = strings.TrimSpace(strings.ReplaceAll(value, "\r\n", "\n"))
	if value == "" {
		return structuredSummary{Fields: fields}
	}

	labelPattern := regexp.MustCompile(`(?i)(phone number|phone|budget|rooms|location|property type|sentiment|call outcome)\s*:`)
	matches := labelPattern.FindAllStringSubmatchIndex(value, -1)
	for index, match := range matches {
		label := strings.ToLower(strings.TrimSpace(value[match[2]:match[3]]))
		start := match[1]
		end := len(value)
		if index+1 < len(matches) {
			end = matches[index+1][0]
		}
		fieldValue := strings.TrimSpace(value[start:end])
		fieldValue = strings.Trim(fieldValue, "\n\t ")
		if fieldValue != "" {
			fields[label] = fieldValue
		}
	}

	return structuredSummary{Fields: fields}
}

func isStructuredSummary(value string) bool {
	return len(parseStructuredSummary(value).Fields) > 0
}

func structuredCallSummary(summary structuredSummary) string {
	if len(summary.Fields) == 0 {
		return ""
	}

	outcome := humanize(firstNonEmpty(summary.Fields["call outcome"], "call"))
	phone := firstNonEmpty(summary.Fields["phone number"], summary.Fields["phone"])
	if phone != "" {
		return fmt.Sprintf("%s from %s.", outcome, normalizePhone(phone))
	}
	return outcome + "."
}

func conciseCallSummary(phone string, outcome db.CallOutcome) string {
	summary := "Call"
	if outcome != "" {
		summary = humanize(string(outcome)) + " call"
	}
	if phone != "" {
		summary += " with " + phone
	}
	return summary + "."
}

func structuredLeadDescription(summary structuredSummary) string {
	parts := []string{}
	qualification := map[string]string{}
	for _, key := range []string{"budget", "rooms", "location", "property type"} {
		value := meaningfulStructuredValue(summary.Fields[key])
		if value != "" {
			qualification[key] = value
		}
	}
	if len(qualification) > 0 {
		data, err := json.Marshal(qualification)
		if err == nil {
			parts = append(parts, "Qualification: "+string(data))
		}
	}
	if outcome := meaningfulStructuredValue(summary.Fields["call outcome"]); outcome != "" {
		parts = append(parts, "Intent: "+outcome)
	}
	if len(parts) == 0 {
		return structuredCallSummary(summary)
	}
	parts = append(parts, structuredCallSummary(summary))
	return strings.Join(parts, "\n")
}

func structuredDetails(summary structuredSummary) string {
	if len(summary.Fields) == 0 {
		return ""
	}

	parts := []string{}
	for _, key := range []string{"budget", "rooms", "location", "property type", "sentiment", "call outcome"} {
		value := meaningfulStructuredValue(summary.Fields[key])
		if value != "" {
			parts = append(parts, strings.ReplaceAll(key, " ", "_")+": "+value)
		}
	}
	return strings.Join(parts, "\n")
}

func meaningfulStructuredValue(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(strings.ToLower(value), "assistant:") {
		return ""
	}
	return value
}

func detailsWithoutAssistantPrompts(value string) string {
	if isStructuredSummary(value) {
		return ""
	}
	lines := []string{}
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(strings.ToLower(line), ": assistant:") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func mergeDetails(values ...string) string {
	parts := []string{}
	for _, value := range values {
		for _, line := range strings.Split(value, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				parts = append(parts, line)
			}
		}
	}
	return strings.Join(parts, "\n")
}

func firstLeadStatus(values ...db.LeadStatus) db.LeadStatus {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func humanize(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "_", " "))
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
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
