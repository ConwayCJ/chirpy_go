package main

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type apiConfig struct {
	fileserverHits int
	db             *DB
}

func main() {
	// create database/point to existing db
	db, err := NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}

	cfg := &apiConfig{
		fileserverHits: 0,
		db:             db,
	}

	const port = "8080"
	filepathRoot := "."

	mux := http.NewServeMux()
	mux.Handle("/app/*", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))))
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.HandleFunc("/api/reset", cfg.srvResetCount)
	mux.HandleFunc("GET /api/healthz", srvrHealth)
	mux.HandleFunc("GET /admin/metrics", cfg.srvMetrics)
	// chirps
	mux.HandleFunc("GET /api/chirps", cfg.retrieveChirps)
	mux.HandleFunc("GET /api/chirps/{id}", cfg.retrieveChirp)
	mux.HandleFunc("POST /api/chirps", cfg.postNewChirp)
	// users
	mux.HandleFunc("POST /api/users", cfg.postNewUser)
	mux.HandleFunc("/api/login", cfg.handleLogin)
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	dbStruct, err := cfg.db.loadDB()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to load db")
		return
	}

	var usr parameters
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&usr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "issue decoding json")
		return
	}

	dbUsr, exists := findExistingUser(dbStruct, usr.Email)
	if !exists {
		respondWithError(w, http.StatusNotFound, "invalid email address")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUsr.Password), []byte(usr.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect password")
		return
	}

	respondWithJSON(w, 200, dbUsr)
}
