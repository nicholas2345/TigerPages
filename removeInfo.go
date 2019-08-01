package main

import (
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

// removes row/rows from the given table in the given database
func remove(form map[string]string) (success bool, err error) {

	// if map empty
	if len(form) == 0 {
		fmt.Println("No arguments passed")
		return false, errors.New("No arguments passed")
	}

	// criteria for each table that represents a UNIQUE row
	// clubs = club_id, links = club_id and link_type, admins = net_id and club_id,
	// users = net_id, categories = club_id and category, postings = posting_id
	deleteCriteria := map[string][]string{
		"clubs":      {"club_id"},
		"links":      {"club_id", "link_type"},
		"followers":  {"id"},
		"users":      {"net_id"},
		"categories": {"club_id", "category"},
		"postings":   {"posting_id"},
		"requests":   {"rid"},
	}

	// build query string
	queryString := fmt.Sprintf("DELETE FROM %s where ", form["table"])

	// add criteria to your delete, making an interface
	criteria := deleteCriteria[form["table"]]
	var formValues []string
	for i, value := range criteria {
		queryString += fmt.Sprintf("%s=$%d AND ", value, i+1)
		formValues = append(formValues, form[value])
	}

	queryString = queryString[0 : len(queryString)-5]

	// Using executeTransaction from the transaction.go file, safely change information in the
	// tables via the use of transactions
	err = executeTransaction(queryString, formValues)
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}
