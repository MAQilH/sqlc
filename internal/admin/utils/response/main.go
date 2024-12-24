package response

import (
	"encoding/json"
	"github.com/sqlc-dev/sqlc/internal/admin/model"
	"net/http"
)

func SendSucJsonMessage(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&model.Response{message}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
