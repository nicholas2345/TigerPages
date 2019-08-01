package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// handles logging in to TigerPages via CAS
func login(w http.ResponseWriter, r *http.Request) {
	// set CAS url
	CASurl := "https://fed.princeton.edu/cas/login?service="

	// The url of the request is the service we are trying to reach
	service := r.URL.Scheme + "://" + r.Host + "/login/"
	if r.URL.Scheme == "" {
		service = "http://" + r.Host + "/login/"
	}

	// if ticket is not none, validate it
	ticket := r.URL.Query().Get("ticket")
	if ticket != "" {
		valid, netID := validate(service, ticket)
		if valid {
			token, err := getToken(netID)
			if err != nil {
				fmt.Println(err)
				http.Redirect(w, r, "/", http.StatusFound)

			}
			// use cookies to set token
			http.SetCookie(w, &http.Cookie{
				Name: tokenName,
				Value: string(token),
				Path: "/",
				RawExpires: "0",
			})

			http.Redirect(w, r, "/home/", http.StatusFound)
		}
	}
	// Send a GET request to CAS in the form of a redirect
	http.Redirect(w, r, CASurl + service, http.StatusSeeOther)
}

// sees if tickets for CAS are validated. If so, return true and the netID of a user. Otherwise return false
// and the empty string
func validate(service string, ticket string) (validated bool, netID string) {
	validationURL := "https://fed.princeton.edu/cas/validate?service=" + url.QueryEscape(service) +
		"&ticket=" + url.QueryEscape(ticket)
	r, err := http.Get(validationURL)
	if err != nil {
		fmt.Println(err)
	}
	defer r.Body.Close()
	if r.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		response := strings.Split(string(bodyBytes), "\n")

		if response[0] == "yes" {
			return true, response[1]
		} else {
			return false, ""
		}

	} else {
		return false,""

	}
}
