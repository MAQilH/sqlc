package admin

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/sqlc-dev/sqlc/internal/admin/model"
	"github.com/sqlc-dev/sqlc/internal/admin/utils/jwt"
	"github.com/sqlc-dev/sqlc/internal/admin/utils/response"
	"io"
	"net/http"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handle login request")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fmt.Println("handle login request1")

	var loginRequest model.LoginRequest
	err = json.Unmarshal(body, &loginRequest)
	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}
	fmt.Println("handle login request2", loginRequest.Username)

	admin, err := dr.FindAdminWithUsername(loginRequest.Username)
	if err != nil {
		http.Error(w, "Invalid Username or Password", http.StatusNotFound)
		return
	}
	fmt.Println("find admin")

	if admin.Password != loginRequest.Password {
		fmt.Println("User enter invalid information!")
		http.Error(w, "Invalid Username or Password", http.StatusNotFound)
		return
	}
	fmt.Println("create token was called")
	token, err := jwt.CreateTokenForAdmin(admin.Username)
	if err != nil {
		http.Error(w, "Unable to create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&model.LoginResponse{
		"Your Logged in successfully!",
		token,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var registerRequest model.RegisterRequest
	err = json.Unmarshal(body, &registerRequest)
	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}

	_, err = dr.FindAdminWithUsername(registerRequest.Username)
	if err == nil {
		http.Error(w, "This username already taken!", http.StatusNotFound)
		return
	}

	_, err = dr.InsertAdmin(model.AdminSchema{
		registerRequest.Username,
		registerRequest.Password,
		registerRequest.Email,
		registerRequest.TelegramID,
	})
	if err != nil {
		http.Error(w, "Unable to insert admin", http.StatusInternalServerError)
		return
	}

	response.SendSucJsonMessage(w, "You are registered successfully!")
}

func handleVerifyToken(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var verifyTokenRequest model.VerifyTokenRequest
	err = json.Unmarshal(body, &verifyTokenRequest)
	if err != nil {
		http.Error(w, "Unable to parse body", http.StatusBadRequest)
		return
	}

	_, err = jwt.ValidateToken(verifyTokenRequest.Token)
	if err != nil {
		http.Error(w, "Invalid Token", http.StatusNotFound)
	}

	response.SendSucJsonMessage(w, "You are verified successfully!")
}

func RunAuth(r chi.Router) {
	r.Post("/login", handleLogin)
	r.Post("/register", handleRegister)
	r.Post("/verifyToken", handleVerifyToken)
}
