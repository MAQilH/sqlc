package admin

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

var dr DatabaseRepository

func RunAdminServer() {
	loadEnvVariable()
	dr = NewDatabaseService()
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Group(RunDashboard)
	r.Group(RunAuth)

	staticFiles := http.FileServer(http.Dir("./build"))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("./build" + r.RequestURI); os.IsNotExist(err) {
			http.ServeFile(w, r, "./build/index.html")
		} else {
			staticFiles.ServeHTTP(w, r)
		}
	})

	fmt.Println("Your admin server listen on port 6060")
	err := http.ListenAndServe(":6060", r)
	if err != nil {
		fmt.Println("Its failed to create an server on port 6060")
		return
	}
}

func loadEnvVariable() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
