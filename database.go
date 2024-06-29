package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

type DB struct {
	path string
	mx   *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func NewDB(path string) (*DB, error) {
	newDB := &DB{
		path: path,
		mx:   &sync.RWMutex{},
	}
	err := newDB.ensureDB()
	return newDB, err
}

func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)
	dbStruct := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
	}

	// checks if file doesn't exist, creates if true
	if errors.Is(err, os.ErrNotExist) {
		db.writeToDB(dbStruct)
	}

	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mx.Lock()
	defer db.mx.Unlock()

	DBStructure := DBStructure{}

	dat, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return DBStructure, err
	}

	err = json.Unmarshal(dat, &DBStructure)
	if err != nil {
		return DBStructure, err
	}

	return DBStructure, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStr, err := db.loadDB()
	if err != nil {
		log.Fatal("issue reading from database")
	}

	chirps := make([]Chirp, 0, len(dbStr.Chirps))
	for _, chirp := range dbStr.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) GetUsers() ([]User, error) {
	dbStr, err := db.loadDB()
	if err != nil {
		log.Fatal("issue reading from database")
	}

	users := make([]User, 0, len(dbStr.Users))
	for _, user := range dbStr.Users {
		users = append(users, user)
	}

	return users, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	allChirps, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	if chirp, exists := allChirps.Chirps[id]; !exists {
		return chirp, nil
	}

	return Chirp{}, os.ErrNotExist

}

func (db *DB) writeToDB(dbStr DBStructure) error {
	db.mx.RLock()
	defer db.mx.RUnlock()

	dat, err := json.Marshal(dbStr)
	if err != nil {
		return nil
	}

	err = os.WriteFile(db.path, dat, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbstruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbstruct.Chirps) + 1
	chirp := Chirp{
		Id:   id,
		Body: body,
	}

	dbstruct.Chirps[id] = chirp

	err = db.writeToDB(dbstruct)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func findExistingUser(db DBStructure, email string) (User, bool) {
	users := db.Users

	for _, user := range users {
		if user.Email == email {
			return user, true
		}
	}

	return User{}, false
}

func (db *DB) CreateUser(email, password string) (User, error) {
	dbstruct, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if _, ok := findExistingUser(dbstruct, email); ok {
		return User{}, errors.New("email already exists")
	}

	id := len(dbstruct.Users) + 1
	user := User{
		Id:       id,
		Email:    email,
		Password: password,
	}

	dbstruct.Users[id] = user

	err = db.writeToDB(dbstruct)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
