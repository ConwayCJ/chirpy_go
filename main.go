package main

import (
	"log"
	"net/http"
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

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
