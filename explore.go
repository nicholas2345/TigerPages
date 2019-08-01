package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// struct used for random clubs
type ClubInfo struct {
	Clubs [][3]string
}

// handles explore page
func exploreHandler(w http.ResponseWriter, r *http.Request) {

	// validate user
	valid, _ := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}

	// Select 3 random rows from clubs page in Postgres
	queryString := "Select club_id, name, bio from clubs order by random() limit 3;"
	stmt, err := db.Prepare(queryString)
	if err != nil {
		http.Redirect(w, r, "/error/", http.StatusFound)
		fmt.Println(err.Error())
		return
	}

	rows, err := stmt.Query()
	defer rows.Close()
	if err != nil {
		http.Redirect(w, r, "/error/", http.StatusFound)
		fmt.Println(err.Error())
		return
	}

	clubs := [][3]string{}
	for rows.Next() {
		club := [3]string{}
		err = rows.Scan(&club[0], &club[1], &club[2])
		if err != nil {
			http.Redirect(w, r, "/error/", http.StatusFound)
			fmt.Println(err.Error())
			return
		}
		clubs = append(clubs, club)

	}
	clubInfo := ClubInfo{Clubs: clubs}
	servePage(w, "explorePage.html", clubInfo)
}

//// handles explore results page
func exploreResultsHandler(w http.ResponseWriter, r *http.Request) {
	// validate user
	valid, _ := validateUser(w, r)

	if !valid {
		http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		return
	}
	// get categories
	vars := mux.Vars(r)
	categoriesString := vars["categories"]
	categories := strings.Split(categoriesString, "&")
	fmt.Println("-----------------------------------------------")
	fmt.Println("Received request to access pages matching to the following categories:", categories)
	success, err, results := getSearchFormResults(categories)
	if !success {
		http.Redirect(w, r, "/error/", http.StatusFound)
		fmt.Println(err.Error())
		fmt.Println("-----------------------------------------------")
		return
	}
	type Results struct {
		Data [][4]string
	}
	servePage(w, "exploreResultsPage.html", Results{Data: results})
	fmt.Println("Successfully served request")
	fmt.Println("-----------------------------------------------")
}

// For specific categories return clubs that match
func getSearchFormResults(categories []string) (success bool, err error, results [][4]string) {

	// Query string
	queryString := "SELECT categories.category, clubs.name, clubs.bio, clubs.club_id FROM categories " +
		"LEFT JOIN clubs ON clubs.club_id = categories.club_id WHERE "
	for _, category := range categories {
		queryString += fmt.Sprintf("categories.category = '%s' OR ", category)
	}
	queryString = queryString[0:len(queryString)-4] + ";"
	stmt, err := db.Prepare(queryString)
	if err != nil {
		return false, err, nil
	}
	rows, err := stmt.Query()
	defer rows.Close()
	if err != nil {
		return false, err, nil
	}

	results = [][4]string{}
	for rows.Next() {
		var result [4]string
		err := rows.Scan(&result[0], &result[1], &result[2], &result[3])
		if err != nil {
			return false, err, nil
		}
		results = append(results, result)
	}
	return true, nil, results
}
