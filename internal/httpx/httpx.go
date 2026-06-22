package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AGX18/real_estate_crm/internal/logger"

	"github.com/jackc/pgx/v5/pgtype"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message string, causes ...error) {
	if len(causes) > 0 && causes[0] != nil {
		l := logger.Log
		if l == nil {
			l = slog.Default()
		}
		l.Error("endpoint error cause",
			"status", status,
			"message", message,
			"error", causes[0],
		)
	}
	WriteJSON(w, status, map[string]string{"error": message})
}

func ParseUUID(id string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(id)
	return uuid, err
}

func ParseInt64(id string) (int64, error) {
	return strconv.ParseInt(id, 10, 64)
}
