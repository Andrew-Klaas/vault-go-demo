package config

import (

	// "database/sql"
	// "net/http"
	// "github.com/vault/api"

	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// DBuser holds DB user info
type DBuser struct {
	Username string
	Password string
}

// DB Connection
var DB *sql.DB
var UserDB *sql.DB

var AppDBuser DBuser
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

var Conf = &oauth2.Config{
	ClientID:     "", //Set in init. Read from Vault
	ClientSecret: "", //Set in init. Read from Vault
	Endpoint:     google.Endpoint,
	RedirectURL:  "http://aklaas.sbx.hashidemos.io/oauth2/google/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
}

// Vclient ...
var Vclient, _ = api.NewClient(&api.Config{Address: "http://vault-ui.default.svc.cluster.local:8200", HttpClient: httpClient})

//var Vclient, _ = api.NewClient(&api.Config{Address: "http://localhost:8200", HttpClient: httpClient})

var tokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
var K8sAuthRole = "vault_go_demo"
var K8sAuthPath = "auth/kubernetes/login"

// FAKE
var AccessKeyId = "ASIAIOSFODNN7EXAMPLE"
var SecretAccessKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

func init() {
	//Vault
	//K8s
	fmt.Printf("Vault client init\n")
	buf, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		log.Fatal(err)
	}
	jwt := string(buf)
	fmt.Printf("K8s Service Account JWT: %v\n", jwt)

	config := map[string]interface{}{
		"jwt":  jwt,
		"role": K8sAuthRole,
	}

	secret, err1 := Vclient.Logical().Write(K8sAuthPath, config)
	fmt.Printf("Secret: %v\n", secret)
	if err1 != nil {
		log.Fatal(err1)
	}
	token := secret.Auth.ClientToken

	//Local
	// token := "password"

	Vclient.SetToken(token)

	data, err := Vclient.Logical().Read("database/creds/vault_go_demo")
	if err != nil {
		log.Fatal(err)
	}
	username := data.Data["username"]
	password := data.Data["password"]
	SQLQuery := "postgres://" + username.(string) + ":" + password.(string) + "@pq-postgresql.default.svc.cluster.local:5432/vault_go_demo?sslmode=disable"
	//SQLQuery := "postgres://" + username.(string) + ":" + password.(string) + "@localhost:5432/vault_go_demo?sslmode=disable"

	AppDBuser.Username = username.(string)
	AppDBuser.Password = password.(string)

	fmt.Printf("\nDB Username: %v\n", AppDBuser.Username)
	fmt.Printf("DB Password: %v\n\n", AppDBuser.Password)

	//DB setup
	DB, err = sql.Open("postgres", SQLQuery)
	if err != nil {
		log.Fatal(err)
	}
	if err = DB.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("Connected to database")

	// fmt.Println("Populating DB with example users")
	SQLQuery = "DROP TABLE vault_go_demo;"
	DB.Exec(SQLQuery)
	SQLQuery = "CREATE TABLE vault_go_demo (CUST_NO SERIAL PRIMARY KEY, FIRST TEXT NOT NULL, LAST TEXT NOT NULL, SSN TEXT NOT NULL, ADDR CHAR(50), BDAY DATE DEFAULT '1900-01-01', SALARY REAL DEFAULT 25500.00);"
	DB.Exec(SQLQuery)
	SQLQuery = "INSERT INTO vault_go_demo (FIRST, LAST, SSN, ADDR, BDAY, SALARY) VALUES('John', 'Doe', '435-59-5123', '456 Main Street', '1980-01-01', 60000.00);"
	DB.Exec(SQLQuery)
	SQLQuery = "INSERT INTO vault_go_demo (FIRST, LAST, SSN, ADDR, BDAY, SALARY) VALUES('Jane', 'Smith', '765-24-2083', '331 Johnson Street', '1985-02-02', 120000.00);"
	DB.Exec(SQLQuery)
	SQLQuery = "INSERT INTO vault_go_demo (FIRST, LAST, SSN, ADDR, BDAY, SALARY) VALUES('Ben', 'Franklin', '111-22-8084', '222 Chicago Street', '1985-02-02', 180000.00);"
	DB.Exec(SQLQuery)
	// SQLQuery = "INSERT INTO vault_go_demo (FIRST, LAST, SSN, ADDR, BDAY, SALARY) VALUES('Bill', 'Franklin', '111-22-8084', '222 Chicago Street', '1985-02-02', 180000.00);"
	// DB.Exec(SQLQuery)
	//test

	//setup Oauth2 config
	oauth2VaultResp, err := Vclient.Logical().Read("secret/data/oauth2/config")
	if err != nil {
		log.Fatal(err)
	}
	oauth2Data := oauth2VaultResp.Data["data"].(map[string]interface{})
	Conf.ClientID = oauth2Data["client_id"].(string)
	Conf.ClientSecret = oauth2Data["client_secret"].(string)

	//Create UserDB
	// SQLQuery = "DROP TABLE user_db;"
	// SQLQueryConnString := "postgres://" + username.(string) + ":" + password.(string) + "@127.0.0.1:5432/user_db?sslmode=disable"
	// UserDB, err = sql.Open("postgres", SQLQueryConnString)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// err = UserDB.Ping()
	// if err != nil {
	// 	panic(err)
	// }

	// // Prepare the SQL insert statement
	// sqlStatement := `
	// INSERT INTO users (username, password)
	// VALUES ($1, $2)
	// RETURNING user_id`

	// var exampleEmail string = "admin"
	// user := SystemUser{
	// 	Username: "myUsername",
	// 	Password: encryptedPw,
	// }

	// // Execute the SQL statement with user struct values as insert arguments
	// var userID int
	// err = UserDB.QueryRow(sqlStatement, user.Username, user.Password).Scan(&userID)
	// if err != nil {
	// 	panic(err)
	// }

}

/*
UserDB[email] = SystemUser{
	Username: name,
}

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
_, err = config.DB.Exec("INSERT INTO vault_go_demo (FIRST, LAST, SSN, ADDR, BDAY, SALARY) VALUES ($1, $2, $3, $4, $5, $6)", u.First, u.Last, u.Ssn, u.Addr, u.Bday, u.Salary)
*/

// Create Table vault-go-demo (
// 	CUST_NO SERIAL PRIMARY KEY,
// 	FIRST               TEXT NOT NULL,
// 	LAST                TEXT NOT NULL,
// 	SSN                 TEXT NOT NULL,
// 	ADDR                CHAR(50),
// 	BDAY			    DATE DEFAULT '1900-01-01',
// 	SALARY              REAL DEFAULT 25500.00
// );
