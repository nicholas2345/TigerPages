package main

import (
	"errors"
	"fmt"
)

// Different from upsertInfo, ensureConsistency ensures that if a user or club updates
// their name, it is reflected in the followers table to ensure consistency across
// tables
func ensureConsistency(form map[string]string) (success bool, err error) {
	// if map empty
	if len(form) == 0 {
		fmt.Println("No arguments passed")
		return false, errors.New("No arguments passed")
	}

	// depending on if updated table is clubs or users, change query string values
	baseString := "UPDATE followers SET %s=$1 WHERE %s=$2;"
	var queryString string
	switch {
	case form["table"] == "clubs":
		queryString = fmt.Sprintf(baseString, "club_name", "club_id")
	case form["table"] == "users":
		queryString = fmt.Sprintf(baseString, "name", "net_id")
	default:
		return false, errors.New("Improper table name specified")
	}

	// Using executeTransaction from the transaction.go file, safely change information in the
	// tables via the use of transactions
	args := []string{form["name"], form["id"]}
	err = executeTransaction(queryString, args)
	if err != nil {
		return false, err
	}
	return true, nil
}
