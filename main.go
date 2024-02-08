package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Andrew-Klaas/vault-go-demo/users"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func getAlbums(w http.ResponseWriter, r *http.Request) {
	log.Println("getAlbums")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(albums)
}

func main() {
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("public"))))

	http.HandleFunc("/", users.Index)
	http.HandleFunc("/records", users.Records)
	http.HandleFunc("/dbview", users.DbView)
	http.HandleFunc("/addrecord", users.Addrecord)
	http.HandleFunc("/updaterecord", users.UpdateRecord)
	http.HandleFunc("/dbusers", users.DbUserView)
	http.HandleFunc("/api", getAlbums)

	http.Handle("/favicon.ico", http.NotFoundHandler())

	// http.HandleFunc("/oauth2/google/login", users.GoogleLogin)
	// http.HandleFunc("/oauth2/google/callback", users.GoogleCallback)
	// http.HandleFunc("/register", users.Register)
	// http.HandleFunc("/logout", users.Logout)

	log.Println("Listening on port 9090...")
	http.ListenAndServe(":9090", nil)
}
