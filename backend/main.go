package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	//"strconv"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

type Match struct {
	ID        int    `json:"id"`
	HomeTeam  string `json:"homeTeam"`
	AwayTeam  string `json:"awayTeam"`
	MatchDate string `json:"matchDate"`
}

var db *sql.DB

func initDB() {
	var err error

	db, err = sql.Open("sqlite", "liga.db")
	if err != nil {
		log.Fatal("Error al abrir la base de datos:", err)
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
		log.Fatal("Error al crear la tabla:", err)
	}

	fmt.Println("Base de datos conectada")
}

func getMatches(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, home_team, away_team, match_date FROM matches")
	if err != nil {
		http.Error(w, "Error al consultar los partidos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var matches []Match

	for rows.Next() {
		var match Match
		err := rows.Scan(&match.ID, &match.HomeTeam, &match.AwayTeam, &match.MatchDate)
		if err != nil {
			http.Error(w, "Error al leer los datos", http.StatusInternalServerError)
			return
		}
		matches = append(matches, match)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func main() {

	initDB()
	router := mux.NewRouter()
	router.HandleFunc("/api/matches", getMatches).Methods("GET")

	fmt.Println("Servidor corriendo en puerto 8080")
	log.Fatal(http.ListenAndServe(":8080", router))

}
