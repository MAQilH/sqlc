package model

type GetTableDocumentsRequest struct {
	TableName       string `json:"table_name"`
	PageNumber      int    `json:"page_number"`
	DocumentPerPage int    `json:"document_per_page"`
}

type EditDocumentRequest struct {
	TableName       string            `json:"table_name"`
	PrevDocument    map[string]string `json:"prev_document"`
	UpdatedDocument map[string]string `json:"updated_document"`
}

type DropTableRequest struct {
	TableName string `json:"table_name"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	TelegramID string `json:"telegram_id"`
	Email      string `json:"email"`
}

type VerifyTokenRequest struct {
	Token string `json:"token"`
}
