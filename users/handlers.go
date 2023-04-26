package users

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Andrew-Klaas/vault-go-demo/config"
)

// Index ...
func Index(w http.ResponseWriter, req *http.Request) {

	fmt.Printf("username: %v, password %v\n", config.AppDBuser.Username, config.AppDBuser.Password)
	err := config.TPL.ExecuteTemplate(w, "index.gohtml", config.AppDBuser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Create a golang function that does a google oauth2 login
func GoogleLogin(w http.ResponseWriter, req *http.Request) {
	// fmt.Println("google login")

}

// Dbview ...
func DbView(w http.ResponseWriter, req *http.Request) {

	if req.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
	}

	cRecords, err := GetRecords()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("users records %v\n", cRecords)

	err = config.TPL.ExecuteTemplate(w, "dbview.gohtml", cRecords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Records ...
func Records(w http.ResponseWriter, req *http.Request) {

	if req.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
	}

	cRecords, err := GetRecords()
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("users records BEFORE decrypt %v\n", cRecords)
	for i := 3; i < len(cRecords); i++ {
		u := cRecords[i]
		data := map[string]interface{}{
			"ciphertext": string(u.Ssn),
		}
		response, err := config.Vclient.Logical().Write("transit/decrypt/my-key", data)
		if err != nil {
			log.Fatal(err)
		}
		ptxt := strings.Split(response.Data["plaintext"].(string), ":")
		ssn, err := base64.StdEncoding.DecodeString(ptxt[0])
		if err != nil {
			log.Fatal(err)
		}
		cRecords[i].Ssn = string(ssn)
	}

	err = config.TPL.ExecuteTemplate(w, "records.gohtml", cRecords)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// DbUserView ...
func DbUserView(w http.ResponseWriter, req *http.Request) {

	if req.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
	}

	dbUsers, err := GetUsers()
	if err != nil {
		log.Fatal(err)
	}

	err = config.TPL.ExecuteTemplate(w, "dbusers.gohtml", dbUsers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Addrecord ...
func Addrecord(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		f := req.FormValue("first")
		l := req.FormValue("last")
		ssn := req.FormValue("ssn")
		adr := req.FormValue("address")
		bd := req.FormValue("birthday")
		slry := req.FormValue("salary")

		// convert form values
		f64, err := strconv.ParseFloat(slry, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		conSlry := float32(f64)

		u := User{
			Cust_no: "",
			First:   f,
			Last:    l,
			Ssn:     ssn,
			Addr:    adr,
			Bday:    bd,
			Salary:  conSlry,
		}
		fmt.Printf("User record to add: %v\n", u)

		//HashiCorp Vault encryption
		data := map[string]interface{}{
			"plaintext": base64.StdEncoding.EncodeToString([]byte(u.Ssn)),
		}
		response, err := config.Vclient.Logical().Write("transit/encrypt/my-key", data)
		if err != nil {
			log.Fatal(err)
		}
		ctxt := response.Data["ciphertext"].(string)
		fmt.Printf("Vault encrypted ssn: %v\n", ctxt)

		u.Ssn = ctxt
		fmt.Printf("user record to add post encrypt: %v\n", u)

		/*
			SQLQuery = "INSERT INTO vault_go_demo (FIRST, LAST, SSN, ADDR, BDAY, SALARY) VALUES('Bill', 'Franklin', '111-22-8084', '222 Chicago Street', '1985-02-02', 180000.00);"
			DB.Exec(SQLQuery)
		*/
		_, err = config.DB.Exec("INSERT INTO vault_go_demo (FIRST, LAST, SSN, ADDR, BDAY, SALARY) VALUES ($1, $2, $3, $4, $5, $6)", u.First, u.Last, u.Ssn, u.Addr, u.Bday, u.Salary)
		if err != nil {
			log.Fatal(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		http.Redirect(w, req, "/records", http.StatusSeeOther)
	}
	err := config.TPL.ExecuteTemplate(w, "addrecord.gohtml", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// UpdateRecord ...
func UpdateRecord(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		cn := req.FormValue("cust_no")
		f := req.FormValue("first")
		l := req.FormValue("last")
		ssn := req.FormValue("ssn")
		adr := req.FormValue("address")
		bd := req.FormValue("birthday")
		slry := req.FormValue("salary")

		// convert form values
		f64, err := strconv.ParseFloat(slry, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		conSlry := float32(f64)

		u := User{
			Cust_no: cn,
			First:   f,
			Last:    l,
			Ssn:     ssn,
			Addr:    adr,
			Bday:    bd,
			Salary:  conSlry,
		}
		// fmt.Printf("User record to update: %v\n", u)

		//HashiCorp Vault encryption
		data := map[string]interface{}{
			"plaintext": base64.StdEncoding.EncodeToString([]byte(u.Ssn)),
		}
		response, err := config.Vclient.Logical().Write("transit/encrypt/my-key", data)
		if err != nil {
			log.Fatal(err)
		}
		ctxt := response.Data["ciphertext"].(string)
		fmt.Printf("Vault encrypted ssn: %v\n", ctxt)

		u.Ssn = ctxt
		fmt.Printf("user record to update (post encrypt): %v\n", u)

		/*
			_, err = db.Exec("UPDATE books SET isbn = $1, title=$2, author=$3, price=$4 WHERE isbn=$1;", bk.Isbn, bk.Title, bk.Author, bk.Price)
			if err != nil {
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
		*/
		convcn, err := strconv.Atoi(u.Cust_no)
		if err != nil {
			log.Fatal(err)
		}
		_, err = config.DB.Exec("UPDATE vault_go_demo SET CUST_NO=$1, FIRST=$2, LAST=$3, SSN=$4, ADDR=$5, BDAY=$6, SALARY=$7 WHERE CUST_NO=$1;", convcn, u.First, u.Last, u.Ssn, u.Addr, u.Bday, u.Salary)
		if err != nil {
			log.Fatal(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		http.Redirect(w, req, "/records", http.StatusSeeOther)
	}
	err := config.TPL.ExecuteTemplate(w, "updaterecord.gohtml", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
