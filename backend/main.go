package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

var db *sql.DB

type Match struct {
	ID        int    `json:"id"`
	HomeTeam  string `json:"homeTeam"`
	AwayTeam  string `json:"awayTeam"`
	MatchDate string `json:"matchDate"`
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "liga.db")
	if err != nil {
		log.Fatal(err)
	}
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS matches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		home_team TEXT NOT NULL,
		away_team TEXT NOT NULL,
		match_date TEXT NOT NULL
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func getAllMatches(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, home_team, away_team, match_date FROM matches")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var m Match
		err := rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		matches = append(matches, m)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

func getMatchByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	row := db.QueryRow("SELECT id, home_team, away_team, match_date FROM matches WHERE id = ?", id)
	var m Match
	err := row.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate)
	if err == sql.ErrNoRows {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func createMatch(w http.ResponseWriter, r *http.Request) {
	var m Match
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = db.Exec("INSERT INTO matches (home_team, away_team, match_date) VALUES (?, ?, ?)", m.HomeTeam, m.AwayTeam, m.MatchDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func updateMatch(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var m Match
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = db.Exec("UPDATE matches SET home_team = ?, away_team = ?, match_date = ? WHERE id = ?", m.HomeTeam, m.AwayTeam, m.MatchDate, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func deleteMatch(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	_, err := db.Exec("DELETE FROM matches WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func main() {
	initDB()
	r := mux.NewRouter()
	r.HandleFunc("/ping", ping).Methods("GET")
	r.HandleFunc("/api/matches", getAllMatches).Methods("GET")
	r.HandleFunc("/api/matches/{id}", getMatchByID).Methods("GET")
	r.HandleFunc("/api/matches", createMatch).Methods("POST")
	r.HandleFunc("/api/matches/{id}", updateMatch).Methods("PUT")
	r.HandleFunc("/api/matches/{id}", deleteMatch).Methods("DELETE")

	http.ListenAndServe(":8080", r)
}
