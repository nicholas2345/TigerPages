package main

import (
	"net/http"
	"time"
)

// logs you out of CAS and remove yours cookies
func logout(w http.ResponseWriter, r *http.Request) {

	// set CAS url
	CASurl := "https://fed.princeton.edu/cas/logout?service="

	// The url of the request is the service we are trying to reach
	service := r.URL.Scheme + "://" + r.Host
	if r.URL.Scheme == "" {
		service = "http://" + r.Host
	}

	// Unset JWT and netID cookie by making them expire via negative MaxAge and Expires values
	http.SetCookie(w, &http.Cookie{
		Name: tokenName,
		Value: "",
		Path: "/",
		MaxAge: -1,
		Expires: time.Now().Add(-100 * time.Hour),
	})

	// Finally redirect to CAS URL. This will redirect us back to login page
	http.Redirect(w, r, CASurl + service, http.StatusSeeOther)
}