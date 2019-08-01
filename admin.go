package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// struct for club page for admins
type adminPage struct {
	ClubID         string            `json:"ClubID"`
	Name           string            `json:"Name"`
	Bio            string            `json:"Bio"`
	Categories     []string          `json:"Categories"`
	Members        [][2]string       `json:"Members"`
	Links          map[string]string `json:"Links"`
	Postings       [][7]string       `json:"Postings"`
	MemberRequests []string          `json:"MemberRequests"`
	AdminRequests  []string          `json:"AdminRequests"`
}

// handles request for admins
func adminPageHandler(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	// Get admin relevant information and return page
	vars := mux.Vars(r)
	clubID := vars["clubID"]

	// if not an admin, redirect them to normal page
	if !isAdmin(clubID, netID) {
		http.Redirect(w, r, "/club/"+clubID, http.StatusFound)
	}

	fmt.Println("-----------------------------------------------")
	fmt.Println("Received request to access admin page of " + clubID)
	success, pageInfo, err := getAdminInfo(clubID)
	if !success {
		http.Redirect(w, r, "/error/", http.StatusFound)
		fmt.Println(err.Error())
		fmt.Println("-----------------------------------------------")
		return
	}
	servePage(w, "adminClubPage.html", pageInfo)
	fmt.Println("Successfully served request")
	fmt.Println("-----------------------------------------------")
}

// handles POST requests when admins edit club profile and hit save changes
func adminActionHandler(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, _ := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	// first parse the form
	r.ParseMultipartForm(32 << 20)
	// figure out what type of post it is then act accordingly
	action := r.PostFormValue("action")
	clubID := r.PostFormValue("club_id")
	switch {
	case action == "createpost":
		createpost(w, r)
	case action == "editpost":
		editpost(w, r)
	case action == "deletepost":
		deletepost(w, r)
	case action == "memberRequests":
		changeMembership(w, r, clubID)
	case action == "adminRequests":
		changeAdmins(w, r, clubID)
	default:
		editinfo(w, r)
	}
	http.Redirect(w, r, "/club/"+clubID+"/admin/", http.StatusFound)
}

// Similar to getClubInfo, getAdminInfo gets the relevant information of a
// club for the admin version of that club page. The functions that build
// part of the adminPage struct can be found in clubPage.go
func getAdminInfo(clubID string) (success bool, page adminPage, err error) {
	// make all the channels
	errChan := make(chan error, 1)
	basicsChan := make(chan [][2]string, 1)
	categoryChan := make(chan []string, 1)
	linkChan := make(chan map[string]string, 1)
	postChan := make(chan [][7]string, 1)
	memReqChan := make(chan []string, 1)
	adReqChan := make(chan []string, 1)

	// run go routines for all parts
	// create wait group
	var wg sync.WaitGroup
	wg.Add(5)
	go getBasicsAndMembers(clubID, basicsChan, errChan, &wg)
	go getLinks(clubID, linkChan, errChan, &wg)
	go getCategories(clubID, categoryChan, errChan, &wg)
	go getPostings(clubID, postChan, errChan, &wg)
	go getSentRequests(clubID, memReqChan, adReqChan, errChan, &wg)
	wg.Wait()

	// Using a select statement, check if there is an error in error channel. If so act accordingly
	// Otherwise set fields and return
	select {

	case err := <-errChan:
		return false, *new(adminPage), err

	default:
		page.ClubID = clubID

		basics := <-basicsChan
		page.Name = basics[0][0]
		page.Bio = basics[1][0]
		if len(basics) > 2 {
			page.Members = basics[2:]
		}

		page.Links = <-linkChan
		page.Categories = <-categoryChan
		page.Postings = <-postChan
		page.MemberRequests = <-memReqChan
		page.AdminRequests = <-adReqChan

		return true, page, nil
	}
}

// given a clubID, send lists of the membership and admin requests
// for the club to the specified channels
func getSentRequests(clubID string, memReqChan chan []string, adReqChan chan []string, errChan chan error, group *sync.WaitGroup) {
	// defer wg done
	defer group.Done()
	// prepare statement
	stmt, err := db.Prepare("SELECT net_id, type FROM requests WHERE requests.club_id=$1;")
	if err != nil {
		errChan <- err
		return
	}

	// query for rows
	rows, err := stmt.Query(clubID)
	defer rows.Close()
	if err != nil {
		errChan <- err
		return
	}

	// iterate over all rows
	var memberList []string
	var adminList []string
	for rows.Next() {
		var netID string
		var requestType string
		err := rows.Scan(&netID, &requestType)
		if err != nil {
			errChan <- err
			return
		}
		// depending on type of request, add to member or adminList
		if requestType == "member" {
			memberList = append(memberList, netID)
		} else {
			adminList = append(adminList, netID)
		}
	}
	memReqChan <- memberList
	adReqChan <- adminList
}

// creates a post
func createpost(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	clubID := r.PostFormValue("club_id")
	// ensure admin status
	if !isAdmin(clubID, netID) {
		http.Redirect(w, r, "/club/"+clubID, http.StatusFound)
	}

	hasImage := r.PostFormValue("has_image")
	postID := strconv.Itoa(rand.Int())
	if hasImage == "true" {
		file, _, _ := r.FormFile("pic")
		// construct filename s3 does not require a filename
		filename := "clubs/" + clubID + "/posts/" + postID
		fmt.Println(filename)
		go uploadImage(file, filename)
	}
	// get title, blurb, long blurb and member post
	form := map[string]string{
		"table": "postings",
		// posting ID is a random integer
		"posting_id":    postID,
		"club_id":       clubID,
		"title":         r.PostFormValue("title"),
		"blurb":         r.PostFormValue("blurb"),
		"long_blurb":    r.PostFormValue("long_blurb"),
		"member_post":   r.PostFormValue("member_post"),
		"creation_time": time.Now().Format(time.RFC3339),
		"has_image":     hasImage,
	}
	success, err := upsert(form)
	if !success {
		fmt.Println(err)
	}
}

// confirms, deletes or saves for later the membership requests for a given club
func changeMembership(w http.ResponseWriter, r *http.Request, clubID string) {
	// iterate over all keys and values. Keys == netID's, Values == confirm, Deny
	// or later. Confirm means add member, deny means delete request, save for Later
	// means do nothing
	for key, values := range r.PostForm {
		if key == "action" || key == "club_id" || values[0] == "later" {
			continue
		}
		fmt.Println(key + ": " + values[0])
		fmt.Println(clubID)
		fmt.Println(key + clubID)
		deleteRequest := true
		if values[0] == "confirm" {
			deleteRequest = addMember(key, clubID)
		}
		if deleteRequest {
			form := map[string]string{
				"table": "requests",
				"rid":   key + clubID,
			}
			success, err := remove(form)
			if !success {
				fmt.Println(err)
			}
		}
	}
}

// confirms, deletes or saves for later the admin requests for a given club
func changeAdmins(w http.ResponseWriter, r *http.Request, clubID string) {
	// iterate over all keys and values. Keys == netID's, Values == confirm, Deny
	// or later. Confirm means add member, deny means delete request, save for Later
	// means do nothing
	for key, values := range r.PostForm {
		if key == "action" || key == "club_id" || values[0] == "later" {
			continue
		}
		fmt.Println(key + ": " + values[0])
		fmt.Println(clubID)
		fmt.Println(key + clubID)
		deleteRequest := true
		if values[0] == "confirm" {
			deleteRequest = addAdmin(key, clubID)
		}
		if deleteRequest {
			form := map[string]string{
				"table": "requests",
				"rid":   key + clubID,
			}
			success, err := remove(form)
			if !success {
				fmt.Println(err)
			}
		}
	}
}

// edits a post
func editpost(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	clubID := r.PostFormValue("club_id")
	// ensure admin status
	if !isAdmin(clubID, netID) {
		http.Redirect(w, r, "/club/"+clubID, http.StatusFound)
	}

	// if there is an image, post it
	hasImage := r.PostFormValue("has_image")
	postID := r.PostFormValue("posting_id")
	file, _, err := r.FormFile("pic")

	// if there is no error getting pic and it has image, upload it
	if hasImage == "true" && err == nil {
		// construct filename s3 does not require a filename
		filename := "clubs/" + clubID + "/posts/" + postID
		fmt.Println(filename)
		go uploadImage(file, filename)
	}
	// if no image, delete photo
	fmt.Println(hasImage)
	if hasImage == "false" { // else delete the photo from the post
		go deletePhoto("clubs/" + clubID + "/posts/" + postID)
	}

	// get title, blurb, long blurb and member post
	form := map[string]string{
		"table":       "postings",
		"posting_id":  postID,
		"club_id":     clubID,
		"title":       r.PostFormValue("title"),
		"blurb":       r.PostFormValue("blurb"),
		"long_blurb":  r.PostFormValue("long_blurb"),
		"member_post": r.PostFormValue("member_post"),
		"has_image":   hasImage,
	}
	success, err := upsert(form)
	if !success {
		fmt.Println(err)
	}
}

// deletes a post
func deletepost(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}
	clubID := r.PostFormValue("club_id")
	postID := r.PostFormValue("posting_id")
	// ensure admin status
	if !isAdmin(clubID, netID) {
		http.Redirect(w, r, "/club/"+clubID, http.StatusFound)
	}
	// get title, blurb, long blurb and member post
	form := map[string]string{
		"table":      "postings",
		"posting_id": postID,
	}
	success, err := remove(form)
	if !success {
		fmt.Println(err)
	}
	// if post has an image
	if r.PostFormValue("has_image") == "true" {
		go deletePhoto("clubs/" + clubID + "/posts/" + postID)
	}
}

// edits name bio
func editinfo(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	clubID := r.PostFormValue("club_id")
	// ensure admin status
	if !isAdmin(clubID, netID) {
		http.Redirect(w, r, "/club/"+clubID, http.StatusFound)
	}

	file, _, err := r.FormFile("pic")
	if err == nil {
		// construct filename s3 does not require a filename
		filename := "clubs/" + clubID + "/picture"
		fmt.Println(filename)
		go uploadImage(file, filename)
	}

	form := map[string]string{
		"table":        "clubs",
		"club_id":      clubID,
		"name":         r.PostFormValue("name"),
		"bio":          r.PostFormValue("bio"),
		"show_members": "true",
	}
	success, err := upsert(form)
	if !success {
		fmt.Println(err)
	}
	// additionally, ensure the name in users is the same for that in followers
	// by calling ensureConsistency
	consistencyForm := map[string]string{
		"table": "clubs",
		"name":  r.PostFormValue("name"),
		"id":    clubID,
	}
	success, err = ensureConsistency(consistencyForm)
	if !success {
		fmt.Println(err)
	}

	// update links for page
	linkForm := map[string]string{
		"facebook":   r.PostFormValue("facebook"),
		"instagram":  r.PostFormValue("instagram"),
		"twitter":    r.PostFormValue("twitter"),
		"youtube":    r.PostFormValue("youtube"),
		"applemusic": r.PostFormValue("applemusic"),
		"spotify":    r.PostFormValue("spotify"),
		"email":      r.PostFormValue("email"),
		"website":    r.PostFormValue("website"),
	}
	changeLinks(clubID, linkForm)
	if !success {
		fmt.Println(err)
	}
}

// changeLinks is a function that takes in a map of the values for links and updates them
// you want to update each link as a single upsert
func changeLinks(clubID string, form map[string]string) {
	for linkType, value := range form {
		linkForm := map[string]string{
			"table":      "links",
			"link_id":    clubID + linkType,
			"club_id":    clubID,
			"link_type":  linkType,
			"descriptor": value,
		}
		success, err := upsert(linkForm)
		if !success {
			fmt.Println(err)
		}
	}
}
