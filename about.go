package main

import "net/http"

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/aboutPage.html")
}
