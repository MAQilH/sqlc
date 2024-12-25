package model

type ColumnInfo struct {
	ColumnName    string      `json:"column_name"`
	DataType      string      `json:"data_type"`
	IsNullable    string      `json:"is_nullable"`
	ColumnDefault interface{} `json:"column_default"`
}

type TableInfoResponse struct {
	TableName string                `json:"table_name"`
	Columns   map[string]ColumnInfo `json:"columns"`
	TableSize int64                 `json:"table_size"`
}

type DocumentInfo struct {
	TableName  string            `json:"table_name"`
	FieldsInfo map[string]string `json:"document_data"`
}

type Response struct {
	Message string `json:"message"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

type DocumentResponse struct {
	Message  string                 `json:"message"`
	Document map[string]interface{} `json:"document"`
}
