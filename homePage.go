package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
	"time"
	"encoding/json"
)

// create a struct used for templates
type Postings struct {
	Postings [][9]string `json:"postings"` 
}

// handles requests for people
func homePageHandler(w http.ResponseWriter, r *http.Request) {

	// validate user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	// set cookie
	fmt.Println("-----------------------------------------------")
	fmt.Println("Received request to access home page of " + netID)
	index := r.PostFormValue("index")
	// check if user has logged into TigerPages before. If not, serve them the newUserPage
	if userList[netID] != "true" {
		// upsert their info into users table
		form := map[string]string{
			"table":  "users",
			"net_id": netID,
			"name":   netID,
			"bio":    "Click edit info to change bio",
		}
		success, err := upsert(form)
		if !success {
			fmt.Println(err)
		}
		// Add them to userList
		userList[netID] = "true"
		http.ServeFile(w, r, "static/newUserHomePage.html")
	} else  if (index == "") {
		success, err, pageInfo := getInitialFeed(netID)
		if !success {
			http.Redirect(w, r, "/error/", http.StatusFound)
			fmt.Println(err.Error())
			fmt.Println("-----------------------------------------------")
			return
		}

		// make a template
		postings := Postings{Postings: pageInfo}
		servePage(w, "homePage.html", postings)
		// print out success
		fmt.Println("Request successfully served")
		fmt.Println("-----------------------------------------------")
	} else {
		success, err, pageInfo := getMoreFeed(netID, index)
		if !success {
			http.Redirect(w, r, "/error/", http.StatusFound)
			fmt.Println(err.Error())
			fmt.Println("-----------------------------------------------")
			return
		}
		// make a template
		postings := Postings{Postings:pageInfo}
		json, err := json.Marshal(postings)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			fmt.Println(err.Error())
			fmt.Println("-----------------------------------------------")
			return
		}
		// taken from https://appliedgo.net/json/
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
		
		// print out success
		fmt.Println("Successfully sent more information")
	}
}

// given a netID, get all postings
func getInitialFeed(netID string) (success bool, err error, postings [][9]string) {
	queryString := "SELECT DISTINCT title, blurb, long_blurb, creation_time, member_post, followers.member, clubs.name, postings.club_id, postings.posting_id, postings.has_image FROM postings " +
		"LEFT JOIN followers ON followers.club_id = postings.club_id " +
		"LEFT JOIN clubs ON postings.club_id = clubs.club_id " +
		"WHERE followers.net_id=$1 ORDER BY creation_time DESC LIMIT 4;"
	stmt, err := db.Prepare(queryString)
	if err != nil {
		return false, err, nil
	}
	rows, err := stmt.Query(netID)
	defer rows.Close()
	if err != nil {
		return false, err, nil
	}

	for rows.Next() {
		columns := [10]string{}
		var memberPost bool
		var isMember bool
		err := rows.Scan(&columns[0], &columns[1], &columns[2], &columns[3], &memberPost, &isMember, &columns[6], &columns[7], &columns[8], &columns[9])
		if err != nil {
			return false, err, nil
		}
		// ensures no non member views a member post
		if memberPost && !isMember {
			continue
		}

		// transform datetime NOTE: the seemingly random date is just a quirk of Go where Jan 2 2006 at 15:04 is the
		// reference date. It has nothing to do with the actual values
		dt, _ := time.Parse("2006-01-02T15:04:05Z", columns[3])
		columns[3] = dt.Format("Jan 2 '06 at 15:04")
		if memberPost {
			postingInfo := [9]string{columns[0], columns[1], columns[2], columns[3], "Member Post", columns[6], columns[7], columns[8], columns[9]}
			postings = append(postings, postingInfo)
		} else {
			postingInfo := [9]string{columns[0], columns[1], columns[2], columns[3], "Followers Post", columns[6], columns[7], columns[8], columns[9]}
			postings = append(postings, postingInfo)
		}
	}
	return true, nil, postings
}

func getMoreFeed(netID, i string) (success bool, err error, postings [][9]string) {
	queryString := "SELECT DISTINCT title, blurb, long_blurb, creation_time, member_post, followers.member, clubs.name, postings.club_id, postings.posting_id, postings.has_image FROM postings " +
		"LEFT JOIN followers ON followers.club_id = postings.club_id " +
		"LEFT JOIN clubs ON postings.club_id = clubs.club_id " +
		"WHERE followers.net_id=$1 ORDER BY creation_time DESC LIMIT 2 OFFSET $2;"
	stmt, err := db.Prepare(queryString)
	if err != nil {	return false, err, nil}
	rows, err := stmt.Query(netID, i)
	defer rows.Close()
	if err != nil {	return false, err, nil}
	for rows.Next() {
		columns := [10]string{}
		var memberPost bool
		var isMember bool
		err := rows.Scan(&columns[0], &columns[1], &columns[2], &columns[3], &memberPost, &isMember, &columns[6], &columns[7], &columns[8], &columns[9])
		if err != nil {	return false, err, nil}
		if memberPost && !isMember {
			continue
		}

		// transform datetime NOTE: the seemingly random date is just a quirk of Go where Jan 2 2006 at 15:04 is the
		// reference date. It has nothing to do with the actual values
		dt,_ := time.Parse("2006-01-02T15:04:05Z", columns[3])
		columns[3] = dt.Format("Jan 2 '06 at 15:04")
		if memberPost {
			postingInfo := [9]string{columns[0], columns[1], columns[2], columns[3], "Member Post", columns[6], columns[7], columns[8], columns[9]}
			postings = append(postings, postingInfo)
		} else {
			postingInfo := [9]string{columns[0], columns[1], columns[2], columns[3], "Followers Post", columns[6], columns[7], columns[8], columns[9]}
			postings = append(postings, postingInfo)
		}
	}
	return true, nil, postings
}

