package main

import (
	"net/http"
	"time"
)

func sessionTimeoutHandler(w http.ResponseWriter, r *http.Request) {
	// Unset JWT and netID cookie by making them expire via negative MaxAge and Expires values
	http.SetCookie(w, &http.Cookie{
		Name: tokenName,
		Value: "",
		Path: "/",
		MaxAge: -1,
		Expires: time.Now().Add(-100 * time.Hour),
	})
	http.ServeFile(w, r, "static/sessionTimeoutPage.html")
}
