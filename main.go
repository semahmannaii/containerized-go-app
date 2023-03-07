package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Manga struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Chapters int    `json:"chapters"`
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS mangas (id SERIAL PRIMARY KEY, title TEXT, chapters INT)")

	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/mangas", GetMangas(db)).Methods("GET")
	router.HandleFunc("/mangas/{id}", GetManga(db)).Methods("GET")
	router.HandleFunc("/mangas", CreateManga(db)).Methods("POST")
	router.HandleFunc("/mangas/{id}", UpdateManga(db)).Methods("PUT")
	router.HandleFunc("/mangas/{id}", DeleteManga(db)).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", jsonContentTypeMiddleware(router)))
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func GetMangas(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM mangas")

		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		mangas := []Manga{}

		for rows.Next() {
			var manga Manga
			if err := rows.Scan(&manga.ID, &manga.Title, &manga.Chapters); err != nil {
				log.Fatal(err)
			}
			mangas = append(mangas, manga)
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(mangas)
	}
}

func GetManga(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]

		var manga Manga
		err := db.QueryRow("SELECT * FROM mangas WHERE id = $1", id).Scan(&manga.ID, &manga.Title, &manga.Chapters)

		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(manga)
	}
}

func CreateManga(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var manga Manga

		json.NewDecoder(r.Body).Decode(&manga)

		err := db.QueryRow("INSERT INTO mangas (title, chapters) VALUES ($1, $2) RETURNING id", manga.Title, manga.Chapters).Scan(&manga.ID)

		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(manga)
	}
}

func UpdateManga(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var manga Manga
		json.NewDecoder(r.Body).Decode(&manga)

		params := mux.Vars(r)
		id := params["id"]

		_, err := db.Exec("UPDATE mangas SET title = $1, chapters = $2 WHERE id = $3", manga.Title, manga.Chapters, id)

		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(manga)
	}
}

func DeleteManga(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]

		var manga Manga
		err := db.QueryRow("SELECT * FROM mangas WHERE id = $1", id).Scan(&manga.ID, &manga.Title, &manga.Chapters)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := db.Exec("DELETE FROM mangas WHERE id = $1", id)

			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			json.NewEncoder(w).Encode("Manga has been deleted")
		}

	}
}
