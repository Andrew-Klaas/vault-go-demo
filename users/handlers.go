package users

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Andrew-Klaas/vault-go-demo/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type SystemUser struct {
	Username string
	Password []byte
}

var UserDB = map[string]SystemUser{
	"admin": SystemUser{
		Username: "admin",
		Password: []byte("admin"),
	},
}

type CustomClaims struct {
	jwt.StandardClaims
	SID string
	EXP time.Time
}

// Sessions tracks active sessions
// TODO create JWT tokens
// Key is sessionID and value is email address
var sessions = map[string]string{}

// OAuth Conns determines whether a user has been converted to our system from Oauth users
// Key is oAuth User ID and Value is email address in our system
// (our system, not Google - register)
var oAuthConns = map[string]string{}

// OAuthExp determines whether the Oauth user's token is still valid
// Key is our system email address and value is time of expiration for their JWT tokens
var oAuthExp = map[string]time.Time{}

type googleResp struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

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
	// if req.Method != http.MethodPost {
	// 	msg := url.QueryEscape("Method not allowed")
	// 	http.Redirect(w, req, "/?msg="+msg, http.StatusSeeOther)
	// 	return
	// }

	fmt.Printf("performing google login\n")
	sv := uuid.New()
	url := config.Conf.AuthCodeURL(sv.String())
	oAuthExp[sv.String()] = time.Now().Add(time.Hour)
	http.Redirect(w, req, url, http.StatusSeeOther)
}

func GoogleCallback(w http.ResponseWriter, req *http.Request) {
	fmt.Println("google callback")
	// get the code
	ctx := context.Background()
	code := req.FormValue("code")
	if code == "" {
		msg := url.QueryEscape("Code not found")
		http.Redirect(w, req, "/?msg="+msg, http.StatusSeeOther)
		return
	}

	state := req.FormValue("state")
	if state == "" {
		msg := url.QueryEscape("State not found")
		http.Redirect(w, req, "/?msg="+msg, http.StatusSeeOther)
		return
	}

	token, err := config.Conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	ts := config.Conf.TokenSource(req.Context(), token)

	client := oauth2.NewClient(req.Context(), ts)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo?alt=json")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatalf("Status code error: %v\n", resp.StatusCode)
		return
	}

	fmt.Printf("Response: %v\n", resp)
	//Read the response body into a byte array then convert to a json struct
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var gr googleResp
	err = json.Unmarshal(body, &gr)
	if err != nil {
		log.Fatal(err)
	}

	// Check if the user is already in our system
	userEmail, ok := oAuthConns[gr.ID]
	fmt.Printf(" User email: %v\n", userEmail)
	fmt.Printf("oAuthConns: %v\n", oAuthConns)
	if !ok || userEmail == "" {
		jwt := createToken(gr.ID)

		// If not, we need to add user email address to our system
		oAuthConns[gr.ID] = gr.Email

		v := url.Values{}
		v.Add("email", gr.Email)
		v.Add("sst", jwt)
		v.Add("name", gr.Name)
		// We need to redirect them to the register page
		http.Redirect(w, req, "/register?"+v.Encode(), http.StatusSeeOther)
	}
	fmt.Printf("User email found: %v\n", userEmail)
	// The user existed so create a session
	err = createSession(userEmail, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	c, err := req.Cookie("sessionID")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Cookie: %v\n", c)

	http.Redirect(w, req, "/addrecord", http.StatusSeeOther)
}

func Register(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPost {
	// 	msg := url.QueryEscape("Method not allowed")
	// 	http.Redirect(w, r, "/?msg="+msg, http.StatusSeeOther)
	// 	return
	// }

	sst := r.FormValue("sst")
	name := r.FormValue("name")
	email := r.FormValue("email")

	uID, err := parseToken(sst)
	if err != nil {
		http.Redirect(w, r, "/?msg="+url.QueryEscape("Error parsing token"), http.StatusSeeOther)
	}

	UserDB[email] = SystemUser{
		Username: name,
	}

	oAuthConns[uID] = email
	fmt.Printf("email: %v", email)
	fmt.Printf("oAuthConn: %v", oAuthConns)
	err = createSession(email, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/addrecord", http.StatusSeeOther)
}

func createSession(email string, w http.ResponseWriter) error {
	fmt.Printf("Creating Session and Cookie with email: %v\n", email)
	// create a session ID
	sID := uuid.New().String()
	sessions[sID] = email

	// create a JWT token
	token := createToken(sID)

	// set the JWT token as a cookie on the client
	// HttpOnly is set to false for demo purposes
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionID",
		Value:    token,
		Path:     "/",
		HttpOnly: false,
	})

	return nil
}

func createToken(sID string) string {
	cc := CustomClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
		},
		SID: sID,
	}
	fmt.Printf("Custom Claims sID: %v\n", cc.SID)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cc)
	fmt.Printf("Token: %v\n", token)
	//Add Vault Transit
	st, err := token.SignedString([]byte("secret"))
	if err != nil {
		log.Fatal(err)
	}
	return st
}

func parseToken(ss string) (string, error) {
	token, err := jwt.ParseWithClaims(ss, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("error parsing token")
		}
		return []byte("secret"), nil
	})
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		fmt.Printf("claims: %v\n", claims)
	} else {
		fmt.Printf("error: %v\n", err)
		log.Fatal(err)
	}
	fmt.Printf("parsed claims: %v\n", token.Claims.(*CustomClaims))
	return token.Claims.(*CustomClaims).SID, nil
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

	fmt.Printf("users records BEFORE decrypt %v\n", cRecords)
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

func Logout(w http.ResponseWriter, req *http.Request) {
	// get cookie
	fmt.Printf("\nLogging out\n")

	c, err := req.Cookie("sessionID")
	if err != nil {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	// delete the session
	sID, err := parseToken(c.Value)
	if err != nil {
		log.Fatal(err)
	}
	delete(sessions, sID)
	// remove the cookie
	c = &http.Cookie{
		Name:   "sessionID",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
	http.Redirect(w, req, "/", http.StatusSeeOther)
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

		_, err = config.DB.Exec("INSERT INTO vault_go_demo (FIRST, LAST, SSN, ADDR, BDAY, SALARY) VALUES ($1, $2, $3, $4, $5, $6)", u.First, u.Last, u.Ssn, u.Addr, u.Bday, u.Salary)
		if err != nil {
			log.Fatal(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		http.Redirect(w, req, "/records", http.StatusSeeOther)
	}

	var sID string
	var e string
	c, err := req.Cookie("sessionID")
	if err != nil {
		fmt.Printf("Cookie was empty: %v\n", err)
		// c = &http.Cookie{
		// 	Name:  "sessionID",
		// 	Value: "",
		// }
	} else if err == nil {
		sID, err := parseToken(c.Value)
		fmt.Printf("parse token sID: %v\n", sID)
		if err != nil {
			log.Println("index parseToken error: ", err)
		}
	}
	if sID != "" {
		e = sessions[sID]
	}

	err = config.TPL.ExecuteTemplate(w, "addrecord.gohtml", e)
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
