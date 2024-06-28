package main

import (
	"encoding/json"
	"net/http"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) postNewUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to decode params")
		return
	}

	email, password := params.Email, params.Password

	newUser, err := cfg.db.CreateUser(email, password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "issue creating user, try again later")
	}

	respondWithJSON(w, http.StatusCreated, User{
		Id:       newUser.Id,
		Email:    newUser.Email,
		Password: newUser.Password,
	})
}
