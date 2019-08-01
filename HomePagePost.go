package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Post struct {
	Post [7]string
}

type myError struct {
	what string
}

func (e *myError) Error() string {
	return e.what
}

func homePagePostHandler(w http.ResponseWriter, r *http.Request) {

	//validate the user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	//gets postID
	postID := strings.TrimPrefix(r.URL.Path, "/home/")

	// set cookie
	fmt.Println("-----------------------------------------------")
	fmt.Println("Received request to access post of " + netID)
	success, err, pageInfo := getPost(postID, netID)
	if !success {
		http.Redirect(w, r, "/error/", http.StatusFound)
		fmt.Println(err.Error())
		fmt.Println("-----------------------------------------------")
		return
	}

	servePage(w, "homePagePost.html", pageInfo)
	fmt.Println("Request successfully served")
	fmt.Println("-----------------------------------------------")
}

func getPost(postID, netID string) (bool, error, Post) {
	queryString := "SELECT postings.title, postings.club_id, postings.member_post, postings.blurb, postings.long_blurb, postings.creation_time, clubs.name, postings.posting_id FROM postings, clubs WHERE posting_id = $1 AND postings.club_id = clubs.club_id;"
	var columns Post
	var memberPost bool
	// accesses database
	stmt, err := db.Prepare(queryString)
	if err != nil {
		return false, err, columns
	}
	err = stmt.QueryRow(postID).Scan(&columns.Post[0], &columns.Post[1], &memberPost, &columns.Post[2], &columns.Post[3], &columns.Post[4], &columns.Post[5], &columns.Post[6])
	if err != nil {
		return false, err, columns
	}
	dt, _ := time.Parse("2006-01-02T15:04:05Z", columns.Post[4])
	columns.Post[4] = dt.Format("Jan 2 '06 at 15:04")
	// gets info from database and checks if post is a member only post or not
	if err != nil {
		return false, err, columns
	}
	if memberPost {
		return validateMember(netID, columns.Post[1], columns)
	} else {
		return true, nil, columns
	}
}

func validateMember(netID, club_id string, columns Post) (bool, error, Post) {
	fmt.Println("In validate member")
	fmt.Println(netID)
	fmt.Println(club_id)
	queryString := "SELECT member FROM followers WHERE net_id = $1 AND club_id = $2;"
	var member bool
	stmt, err := db.Prepare(queryString)
	if err != nil {
		return false, err, columns
	}
	err = stmt.QueryRow(netID, club_id).Scan(&member)
	if err != nil {
		return false, err, columns
	}
	if member {
		return true, nil, columns
	} else {
		return false, &myError{"User is not Valid"}, columns
	}
}
