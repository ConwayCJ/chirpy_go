package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// GET
func (cfg *apiConfig) retrieveChirp(w http.ResponseWriter, r *http.Request) {
	chirpIdstr := r.PathValue("id")
	chirpId, err := strconv.Atoi(chirpIdstr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}

	dbChirp, err := cfg.db.GetChirp(chirpId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp")
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		Id:   dbChirp.Id,
		Body: dbChirp.Body,
	})
}

func (cfg *apiConfig) retrieveChirps(w http.ResponseWriter, r *http.Request) {

	chirps, err := cfg.db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
	}

	for _, chirp := range chirps {
		chirps = append(chirps, Chirp{
			Id:   chirp.Id,
			Body: chirp.Body,
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].Id < chirps[j].Id
	})

	respondWithJSON(w, http.StatusOK, chirps)
}

// POST
func (cfg *apiConfig) postNewChirp(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to decode params")
		return
	}

	msg, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	newChirp, err := cfg.db.CreateChirp(msg)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create new chirp - try again")
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		Id:   newChirp.Id,
		Body: newChirp.Body,
	})
}

func validateChirp(body string) (string, error) {
	maxChirpLength := 140

	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	return cleanString(body), nil
}

func cleanString(s string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(s, " ")
	sentence := []string{}

	for _, word := range words {
		found := false
		for _, badWord := range badWords {
			if strings.EqualFold(word, badWord) {
				found = true
				break
			}
		}
		if found {
			sentence = append(sentence, "****")
		} else {
			sentence = append(sentence, word)
		}
	}

	return strings.Join(sentence, " ")
}
