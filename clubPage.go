package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// struct that represents the information displayed on a club's page
type clubPage struct {
	NetID      string            `json:"NetID"`
	ClubID     string            `json:"ClubID"`
	IsFollower bool              `json:"IsFollower"`
	IsMember   bool              `json:"IsMember"`
	ReqType    string            `json:"ReqType"`
	Name       string            `json:"Name"`
	Bio        string            `json:"Bio"`
	Categories []string          `json:"Categories"`
	Members    [][2]string       `json:"Members"`
	Links      map[string]string `json:"Links"`
	Postings   [][7]string       `json:"Postings"`
}

// handles requests for club pages
func clubPageHandler(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	clubID := vars["clubID"]
	fmt.Println("-----------------------------------------------")
	fmt.Println("Received request to access club page of " + clubID)

	// check if they are an admin of the page. If so, we redirect them to admin page
	if isAdmin(clubID, netID) {
		fmt.Println("User is an admin and will be redirected to admin version of page.")
		fmt.Println("-----------------------------------------------")
		// get club info for admin version of club pages
		http.Redirect(w, r, "/club/"+clubID+"/admin/", http.StatusFound)

	} else {
		// get club info for regular page
		success, pageInfo, err := getClubInfo(clubID, netID)
		if !success {
			http.Redirect(w, r, "/error/", http.StatusFound)
			fmt.Println(err.Error())
			fmt.Println("-----------------------------------------------")
			return
		}
		servePage(w, "clubPage.html", pageInfo)
		fmt.Println("Successfully served request")
		fmt.Println("-----------------------------------------------")
	}
}

// handles interactions with the NON-ADMIN version of a club page
// pertaining to the following and membership of a club
func clubInteractionHandler(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	clubID := vars["clubID"]
	// from POST form get action
	r.ParseForm()
	action := r.PostFormValue("action")

	switch {
	case action == "follow":
		followClub(netID, clubID)
	case action == "unfollow":
		unfollowClub(netID, clubID)
	case action == "reqmember":
		reqMember(netID, clubID)
	case action == "reqadmin":
		reqAdmin(netID, clubID)
	default:
		leaveMember(netID, clubID)
	}
	http.Redirect(w, r, "/club/"+clubID, http.StatusFound)
}

// given a clubID return their associated links
func getLinks(clubID string, resultChan chan map[string]string, errorChan chan error, group *sync.WaitGroup) {
	// defer wg done
	defer group.Done()
	// use clubID to get links
	stmt, err := db.Prepare("SELECT link_type, descriptor FROM links WHERE links.club_id=$1;")
	if err != nil {
		errorChan <- err
		return
	}
	rows, err := stmt.Query(clubID)
	defer rows.Close()
	if err != nil {
		errorChan <- err
		return
	}
	links := map[string]string{}
	for rows.Next() {
		var linkType string
		var descriptor string

		err = rows.Scan(&linkType, &descriptor)
		if err != nil {
			errorChan <- err
			return
		}
		links[linkType] = descriptor
	}
	resultChan <- links
}

// given clubID return members of that club
func getMembers(clubID string) (members [][2]string, err error) {
	stmt, err := db.Prepare("SELECT name, net_id FROM followers WHERE followers.club_id=$1;")
	if err != nil {
		return [][2]string{}, err
	}
	rows, err := stmt.Query(clubID)
	defer rows.Close()
	if err != nil {
		return [][2]string{}, err
	}
	for rows.Next() {
		var name [2]string
		err = rows.Scan(&name[0], &name[1])
		if err != nil {
			return [][2]string{}, err
		}
		members = append(members, name)
	}
	return members, nil
}

// given clubID, return postings made by that club
func getPostings(clubID string, resultChan chan [][7]string, errorChan chan error, group *sync.WaitGroup) {
	// defer wg done
	defer group.Done()
	queryString := "SELECT title, blurb, long_blurb, creation_time, member_post, posting_id, has_image FROM postings WHERE postings.club_id=$1 " +
		"ORDER BY creation_time DESC;"
	stmt, err := db.Prepare(queryString)
	if err != nil {
		errorChan <- err
		return
	}
	rows, err := stmt.Query(clubID)
	defer rows.Close()
	if err != nil {
		errorChan <- err
		return
	}
	var postings [][7]string
	for rows.Next() {
		posting := [7]string{}
		err := rows.Scan(&posting[0], &posting[1], &posting[2], &posting[3], &posting[4], &posting[5], &posting[6])
		dt, _ := time.Parse("2006-01-02T15:04:05Z", posting[3])
		posting[3] = dt.Format("Jan 2 '06 at 15:04")
		if err != nil {
			errorChan <- err
			return
		}
		postings = append(postings, posting)
	}
	resultChan <- postings
}

// given clubID, get the categories
func getCategories(clubID string, resultChan chan []string, errorChan chan error, group *sync.WaitGroup) {
	// defer wg done
	defer group.Done()
	stmt, err := db.Prepare("SELECT category FROM categories WHERE categories.club_id=$1;")
	if err != nil {
		errorChan <- err
		return
	}
	rows, err := stmt.Query(clubID)
	defer rows.Close()
	if err != nil {
		errorChan <- err
		return
	}
	// append categories and send to channel
	var categories []string
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			errorChan <- err
			return
		}
		categories = append(categories, category)
	}
	resultChan <- categories
}

// gets the name, bio and potentially their members if the club chooses to display their members given a club id
func getBasicsAndMembers(clubID string, resultChan chan [][2]string, errorChan chan error, group *sync.WaitGroup) {
	// defer wg done
	defer group.Done()
	// first get club_id and bio
	stmt, err := db.Prepare("SELECT name, bio, show_members FROM clubs WHERE clubs.club_id=$1;")
	if err != nil {
		errorChan <- err
		return
	}
	rows, err := stmt.Query(clubID)
	defer rows.Close()
	if err != nil {
		errorChan <- err
		return
	}
	// set to variable
	var basics [][2]string
	for rows.Next() {
		var name string
		var bio string
		var showMembers bool
		err = rows.Scan(&name, &bio, &showMembers)
		if err != nil {
			errorChan <- err
			return
		}
		nameSlice := [2]string{name, ""}
		bioSlice := [2]string{bio, ""}
		basics = append(basics, nameSlice, bioSlice)
		// if show members is true, get members and append
		if showMembers {
			members, err := getMembers(clubID)
			if err != nil {
				errorChan <- err
				return
			}
			for _, member := range members {
				basics = append(basics, member)
			}
		}
	}
	resultChan <- basics
}

// given a user's netID, find if they have sent an admin or member request for the pages
func getRequests(netID string, clubID string, reqChan chan string, errChan chan error, group *sync.WaitGroup) {
	// defer wg done
	defer group.Done()
	// first get club_id and bio
	stmt, err := db.Prepare("SELECT type FROM requests WHERE requests.rid=$1;")
	if err != nil {
		errChan <- err
		return
	}

	// query for row and attempt to scan value, if error return nil
	var reqType string
	switch err := stmt.QueryRow(netID + clubID).Scan(&reqType); err {
	// no requests have been sent
	case sql.ErrNoRows:
		reqChan <- "none"
	// if no send req type to chan
	case nil:
		reqChan <- reqType
	default:
		errChan <- err
	}
}

// given a user's netID, find what role they have for the page(follower, member or neither)
// and if they have sent an admin or member request
func getRole(netID string, clubID string, roleChan chan [2]bool, errChan chan error, group *sync.WaitGroup) {
	// defer wg done
	defer group.Done()
	// first get club_id and bio
	stmt, err := db.Prepare("SELECT member FROM followers WHERE followers.net_id=$1 AND followers.club_id=$2;")
	if err != nil {
		errChan <- err
		return
	}

	// query for row and attempt to scan value, if error return false false
	var role [2]bool
	switch err := stmt.QueryRow(netID, clubID).Scan(&role[0]); err {
	case sql.ErrNoRows:
		role[0] = false
		role[1] = false
		roleChan <- role
	// if no error, set role[1] to true (bc if there exists a row, they are a follower) and send
	case nil:
		role[1] = true
		roleChan <- role
	default:
		errChan <- err
	}
}

// get club page info for normal page
func getClubInfo(clubID string, netID string) (success bool, page clubPage, err error) {

	// make all the channels
	errChan := make(chan error, 1)
	basicsChan := make(chan [][2]string, 1)
	roleChan := make(chan [2]bool, 1)
	categoryChan := make(chan []string, 1)
	linkChan := make(chan map[string]string, 1)
	postChan := make(chan [][7]string, 1)
	reqChan := make(chan string, 1)

	// run go routines for all parts
	// create wait group

	var wg sync.WaitGroup
	wg.Add(6)
	go getBasicsAndMembers(clubID, basicsChan, errChan, &wg)
	go getRole(netID, clubID, roleChan, errChan, &wg)
	go getRequests(netID, clubID, reqChan, errChan, &wg)
	go getLinks(clubID, linkChan, errChan, &wg)
	go getCategories(clubID, categoryChan, errChan, &wg)
	go getPostings(clubID, postChan, errChan, &wg)
	wg.Wait()

	// Using a select statement, check if there is an error in error channel. If so act accordingly
	// Otherwise set fields and return
	select {

	case err := <-errChan:
		return false, *new(clubPage), err

	default:
		page.NetID = netID
		page.ClubID = clubID

		role := <-roleChan
		page.IsMember = role[0]
		page.IsFollower = role[1]

		page.ReqType = <-reqChan

		basics := <-basicsChan
		page.Name = basics[0][0]
		page.Bio = basics[1][0]
		if len(basics) > 2 {
			page.Members = basics[2:]
		}

		page.Links = <-linkChan

		page.Categories = <-categoryChan

		page.Postings = <-postChan

		return true, page, nil
	}

}
