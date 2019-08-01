package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type personPage struct {
	NetID   string      `json:"NetID"`
	Name    string      `json:"Name"`
	Bio     string      `json:"Bio"`
	MyClubs [][4]string `json:"MyClubs"`
}

// handles POST requests when you edit your profile and hit save changes
func updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, netIDJWT := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}
	// first parse the form. Also limits file size if there is one to 32 MiB
	r.ParseMultipartForm(32 << 20)
	// Then get form
	netID := r.PostFormValue("net_id")
	name := r.PostFormValue("name")
	bio := r.PostFormValue("bio")
	// as a security measure, ensure that netIDCookieValue matches
	// netID from POST. if it doesn't you don't want someone else editing
	// someone's info. As such if they don't match, redirect to your profile
	if netIDJWT != netID {
		http.Redirect(w, r, "/profile/", http.StatusFound)
	}

	//// If err != nil, this means the user has updated their profile picture.
	//// As such, upload pic to AWS
	file, _, err := r.FormFile("pic")
	if err == nil {
		// construct filename s3 does not require a filename
		filename := "users/" + netID + "/picture"
		fmt.Println(filename)
		go uploadImage(file, filename)
	}

	// build map from new values and call upsert
	form := map[string]string{
		"table":  "users",
		"net_id": netID,
		"name":   name,
		"bio":    bio,
	}
	success, err := upsert(form)
	if !success {
		fmt.Println(err)
	}

	// additionally, ensure the name in users is the same for that in followers
	// by calling ensureConsistency
	consistencyForm := map[string]string{
		"table": "users",
		"name":  name,
		"id":    netID,
	}
	success, err = ensureConsistency(consistencyForm)
	if !success {
		fmt.Println(err)
	}

	// redirect to profile
	http.Redirect(w, r, "/profile/", http.StatusFound)
}

// handles requests for your profile
func profilePageHandler(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	fmt.Println("-----------------------------------------------")
	fmt.Println("Received request to access the profile page of " + netID)
	success, err, pageInfo := getPersonInfo(netID)
	if !success {
		http.Redirect(w, r, "/error/", http.StatusFound)
		fmt.Println(err.Error())
		fmt.Println("-----------------------------------------------")
		return
	}
	page := personPage{NetID: netID, Name: pageInfo.Name, Bio: pageInfo.Bio, MyClubs: pageInfo.MyClubs}
	servePage(w, "profilePage.html", page)
	fmt.Println("Successfully served request")
	fmt.Println("-----------------------------------------------")
}

// handles requests for people
func studentPageHandler(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}
	vars := mux.Vars(r)
	person := vars["netID"]

	// if person is you aka person == netID, redirect to profile page
	if person == netID {
		http.Redirect(w, r, "/profile/", http.StatusFound)
		return
	}
	fmt.Println("-----------------------------------------------")
	fmt.Println("Received request to access the personal page of " + person)
	success, err, pageInfo := getPersonInfo(person)
	if !success {
		http.Redirect(w, r, "/error/", http.StatusFound)
		fmt.Println(err.Error())
		fmt.Println("-----------------------------------------------")
		return
	}
	page := personPage{Name: pageInfo.Name, Bio: pageInfo.Bio, MyClubs: pageInfo.MyClubs}
	servePage(w, "personPage.html", page)
	fmt.Println("Successfully served request")
	fmt.Println("-----------------------------------------------")
}

// get person info from netID
func getNameBio(netID string, rc chan [2]string, ec chan error, group *sync.WaitGroup) {
	defer group.Done()
	// make statement
	stmt, err := db.Prepare("SELECT name, bio FROM users WHERE users.net_id=$1;")
	if err != nil {
		ec <- err
		return
	}
	rows, err := stmt.Query(netID)
	defer rows.Close()
	if err != nil {
		ec <- err
		return
	}

	// set name and bio
	var nameBio [2]string
	for rows.Next() {
		err = rows.Scan(&nameBio[0], &nameBio[1])
		if err != nil {
			ec <- err
			return
		}
	}
	rc <- nameBio
}

// get their clubs
func getPersonsClubs(netID string, rc chan [][4]string, ec chan error, group *sync.WaitGroup) {
	defer group.Done()
	// get info about the clubs they're in
	stmt, err := db.Prepare("SELECT club_id, club_name, member, admin FROM followers WHERE followers.net_id=$1;")
	if err != nil {
		ec <- err
		return
	}
	rows, err := stmt.Query(netID)
	defer rows.Close()
	if err != nil {
		ec <- err
		return
	}
	var results [][4]string
	for rows.Next() {
		row := [4]string{}
		err = rows.Scan(&row[0], &row[1], &row[2], &row[3])
		if err != nil {
			ec <- err
			return
		}
		results = append(results, row)
	}
	rc <- results
}

// given a netID, return the page for that person
func getPersonInfo(netID string) (success bool, err error, page personPage) {

	errChan := make(chan error, 1)
	basicsChan := make(chan [2]string, 1)
	clubChan := make(chan [][4]string, 1)

	var wg sync.WaitGroup
	wg.Add(2)
	go getNameBio(netID, basicsChan, errChan, &wg)
	go getPersonsClubs(netID, clubChan, errChan, &wg)
	wg.Wait()

	select {

	case err := <-errChan:
		return false, err, *new(personPage)

	default:
		basics := <-basicsChan
		page.Name = basics[0]
		page.Bio = basics[1]

		page.MyClubs = <-clubChan
	}

	return true, nil, page
}
