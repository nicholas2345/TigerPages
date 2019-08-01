package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gbrlsnchs/jwt"
	"github.com/gorilla/mux"
	_ "github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

// Global variable for the database
var db *sql.DB

// Global variable for the AWS session
var awsSession *session.Session

// A list of netIDs of users who have used the sigType
var userList map[string]string

// location of the files used for signing and verification
const (
	tokenName    = "AccessToken"
	createPagePW = "goprincetontigers"
)

// HS256 verification key
var (
	hs256 = jwt.NewHS256(os.Getenv("SECRET_KEY"))
)

// initializes the database
func start() {

	var err error
	// connect to db and check for error
	// user := "sdduncan"
	// pw := "tpages2018"
	// dbname := "sdduncan"
	//
	// dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
	// 	user, pw, dbname)
	// db, err = sql.Open("postgres", dbinfo)
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	checkError(err)

	// ping
	err = db.Ping()
	checkError(err)
	fmt.Println("Database started")

	// Start AWS Session in us-east-2 region (Ohio)
	awsSession = session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	}))
	fmt.Println("AWS session initialized")

	// set random seed for new posts
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Random seed set")

	// set list of users
	userList = getUserList()
	fmt.Println("Got list of users")
}

// From https://stackoverflow.com/questions/41616975/how-to-redirect-http-to-https-in-gorilla-mux
// Basically what it does is it checks the header to see if a request is a HTTP one. If so, it redirects it to the same
// URL path but just HTTPS
func RedirectToHTTPSRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		proto := req.Header.Get("x-forwarded-proto")
		if proto == "http" || proto == "HTTP" {
			http.Redirect(res, req, fmt.Sprintf("https://%s%s", req.Host, req.URL), http.StatusPermanentRedirect)
			return
		}

		next.ServeHTTP(res, req)

	})
}

// Serves the login page
func loginHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/loginPage.html")

}

// gets all people who have logged into the site before
func getUserList() map[string]string {
	stmt, err := db.Prepare("Select net_id from users where true;")
	checkError(err)
	rows, err := stmt.Query()
	defer rows.Close()
	checkError(err)
	users := make(map[string]string)
	for rows.Next() {
		var netID string
		err = rows.Scan(&netID)
		checkError(err)
		users[netID] = "true"
	}
	return users
}

// main function
func main() {

	// connect to db and check for error
	start()

	// create a new mux router from Gorilla package
	r := mux.NewRouter()

	// Handles logging in and out of TigerPages
	r.HandleFunc("/", loginHandler)
	r.HandleFunc("/login/", login)
	r.HandleFunc("/logout/", logout)

	// Handles about page
	r.HandleFunc("/about/", aboutHandler)
	// Handles Expired Tokens
	r.HandleFunc("/sessiontimeout/", sessionTimeoutHandler)
	// Handles home page
	r.HandleFunc("/home/", homePageHandler)
	// Handles examining a post
	r.HandleFunc("/home/{postid}", homePagePostHandler)
	// error page for errors accessing a page
	r.HandleFunc("/error/", errorHandler)
	// post req create page
	r.HandleFunc("/createclub/", newPagePostHandler).Methods("POST")
	// Handles creating a page
	r.HandleFunc("/createclub/", newPageHandler)
	// Handles any student page
	r.HandleFunc("/person/{netID}", studentPageHandler)
	// handles non-admins interactions with a club pages
	r.HandleFunc("/club/{clubID}", clubInteractionHandler).Methods("POST")
	// Handles club page requests for non admins
	r.HandleFunc("/club/{clubID}", clubPageHandler)
	// handles post requests to a clubpage
	r.HandleFunc("/club/{clubID}/admin/", adminActionHandler).Methods("POST")
	// Handles club page requests for admins
	r.HandleFunc("/club/{clubID}/admin/", adminPageHandler)
	// Handles explore page
	r.HandleFunc("/explore/", exploreHandler)
	// Handles explore results page
	r.HandleFunc("/explore/results/{categories}", exploreResultsHandler)
	// handles post requests when updating profile info
	r.HandleFunc("/profile/", updateProfileHandler).Methods("POST")
	// Handles profile page for other requests
	r.HandleFunc("/profile/", profilePageHandler)
	// Set static file directory in the StaticFiles directory
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	// handle CTLDthanleclaire.com/blog/2014/08/24/handling-ctrl-c-interrupt-signal-in-golang-programs/
	sigC := make(chan os.Signal, 3)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sigType := <-sigC
		fmt.Println("Received", sigType.String(), "and now shutting down database")
		db.Close()
		os.Exit(0)
	}()

	// get the port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// log any fatal attempt at connecting to the server
	httpsRouter := RedirectToHTTPSRouter(r)
	log.Fatal(http.ListenAndServe(":"+port, httpsRouter))
}

// Checks for errors in operations
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
