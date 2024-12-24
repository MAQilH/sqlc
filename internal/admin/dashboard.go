package admin

import (
	"github.com/go-chi/chi/v5"
	"github.com/sqlc-dev/sqlc/internal/admin/utils/jwt"
	"github.com/sqlc-dev/sqlc/internal/admin/utils/response"
	"log"
	"strings"
)

import (
	"encoding/json"
	"fmt"
	"github.com/sqlc-dev/sqlc/internal/admin/model"
	"io"
	"net/http"
	"strconv"
)

func handleTableNames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	tablesNames, _ := dr.GetTablesNames()

	if err := json.NewEncoder(w).Encode(tablesNames); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func handleDropTable(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var dropTableRequest model.DropTableRequest
	err = json.Unmarshal(body, &dropTableRequest)

	fmt.Println(dropTableRequest.TableName)

	err = dr.DropTable(dropTableRequest.TableName)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	response.SendSucJsonMessage(w, "Table dropped successfully!")
}

func handleCreateDocument(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var documentInfo model.DocumentInfo
	err = json.Unmarshal(body, &documentInfo)

	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}

	fmt.Println("Create new document in table", documentInfo.TableName, " with data:", documentInfo.FieldsInfo)

	err = dr.CreateDocument(documentInfo)

	if err != nil {
		http.Error(w, "Unable to create new document because: "+err.Error(), http.StatusBadRequest)
		return
	}

	response.SendSucJsonMessage(w, "Document created successfully!")
}

func handleGetTableInfo(w http.ResponseWriter, r *http.Request) {
	tableName := r.URL.Query().Get("table_name")

	tableColumns, err := dr.GetTableColumns(tableName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	columnMap := make(map[string]model.ColumnInfo)
	for _, column := range tableColumns {
		columnMap[column.ColumnName] = column
	}

	tableSize, err := dr.GetTableSize(tableName)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tableInfo := &model.TableInfoResponse{
		tableName,
		columnMap,
		tableSize,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tableInfo); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func handleGetTableDocuments(w http.ResponseWriter, r *http.Request) {
	tableName := r.URL.Query().Get("table_name")

	pageNumber, err := strconv.Atoi(r.URL.Query().Get("page_number"))
	if err != nil {
		http.Error(w, "Unable to parse page_number", http.StatusBadRequest)
		return
	}

	documentPerPage, err := strconv.Atoi(r.URL.Query().Get("document_per_page"))
	if err != nil {
		http.Error(w, "Unable to parse document_per_page", http.StatusBadRequest)
		return
	}

	documents, err := dr.GetTableDocuments(tableName, pageNumber, documentPerPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(documents); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	return
}

func handleEditDocument(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var editDocumentRequest model.EditDocumentRequest
	err = json.Unmarshal(body, &editDocumentRequest)
	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}

	err = dr.EditDocument(editDocumentRequest.TableName, editDocumentRequest.PrevDocument, editDocumentRequest.UpdatedDocument)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response.SendSucJsonMessage(w, "Document edited successfully!")
}

func handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var documentInfo model.DocumentInfo
	err = json.Unmarshal(body, &documentInfo)
	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}

	err = dr.DeleteDocument(documentInfo.TableName, documentInfo.FieldsInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response.SendSucJsonMessage(w, "Document deleted successfully!")
}

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		log.Println("auth Header: ", authHeader)
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		_, err := jwt.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Your Token was expired!", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RunDashboard(r chi.Router) {
	r.Use(JWTMiddleware)

	r.Get("/tableNames", handleTableNames)
	r.Delete("/dropTable", handleDropTable)
	r.Post("/createDocument", handleCreateDocument)
	r.Get("/tableInfo", handleGetTableInfo)
	r.Get("/tableDocuments", handleGetTableDocuments)
	r.Post("/editDocument", handleEditDocument)
	r.Delete("/deleteDocument", handleDeleteDocument)
}
