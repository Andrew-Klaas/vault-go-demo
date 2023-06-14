package main

import (
	"log"
	"net/http"

	"github.com/Andrew-Klaas/vault-go-demo/users"
)

//Testing

func main() {
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("public"))))

	http.HandleFunc("/", users.Index)
	http.HandleFunc("/records", users.Records)
	http.HandleFunc("/dbview", users.DbView)
	http.HandleFunc("/addrecord", users.Addrecord)
	http.HandleFunc("/updaterecord", users.UpdateRecord)
	http.HandleFunc("/dbusers", users.DbUserView)

	http.Handle("/favicon.ico", http.NotFoundHandler())

	http.HandleFunc("/oauth2/google/login", users.GoogleLogin)
	http.HandleFunc("/oauth2/google/callback", users.GoogleCallback)
	http.HandleFunc("/register", users.Register)
	http.HandleFunc("/logout", users.Logout)

	log.Println("Listening on port 9090...")
	http.ListenAndServe(":9090", nil)

}
