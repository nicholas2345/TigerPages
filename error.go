package main

import "net/http"

func errorHandler(w http.ResponseWriter, r *http.Request) {
	servePage(w, "errorPage.html", nil)
}
