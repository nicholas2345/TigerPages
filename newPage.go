package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
)

type Alert struct {
	DoAlert bool
}

func newPageHandler(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, _ := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}
	servePage(w, "passwordSubmission.html", Alert{DoAlert: false})
}

func newPagePostHandler(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, netID := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	r.ParseMultipartForm(32 << 20)

	// if the post request is a password submission
	if r.PostFormValue("type") == "password" {
		// if password correct, give them clubSubmission page
		if r.PostFormValue("password") == createPagePW {
			http.ServeFile(w, r, "static/clubSubmission.html")
			return
		} else {
			servePage(w, "passwordSubmission.html", Alert{DoAlert: true})
			return
		}
	}

	// if post request is for club submission

	// create club_id
	clubID := strconv.Itoa(rand.Intn(10000000))

	//// If err != nil, this means the person added a photo
	//// As such, upload pic to AWS
	file, _, err := r.FormFile("pic")
	if err == nil {
		// construct filename s3 does not require a filename
		filename := "clubs/" + clubID + "/picture"
		fmt.Println(filename)
		go uploadImage(file, filename)
	}

	// sumit name bio
	clubInfoForm := map[string]string{
		"table":        "clubs",
		"club_id":      clubID,
		"name":         r.PostFormValue("name"),
		"bio":          r.PostFormValue("bio"),
		"show_members": "true",
	}
	success, err := upsert(clubInfoForm)
	if !success {
		fmt.Println(err.Error())
	}

	// next add user as admin of this club
	addAdmin(netID, clubID)

	// next update links for page
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

	// next set their Categories
	for _, category := range r.Form["categories"] {
		categoryForm := map[string]string{
			"table":    "categories",
			"club_id":  clubID,
			"category": category,
		}
		success, err := upsert(categoryForm)
		if !success {
			fmt.Println(err.Error())
		}
	}
	http.Redirect(w, r, "/club/"+clubID, http.StatusFound)
}
