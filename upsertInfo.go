package main

import (
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

// updates/inserts values into the given table in the given database
func upsert(form map[string]string) (success bool, err error) {

	// if map empty
	if len(form) == 0 {
		fmt.Println("No arguments passed")
		return false, errors.New("No arguments passed")
	}

	// get key and value strings
	keyString, valueString, upsertString := buildStrings(form)

	// Map with conflict criteria for each table
	conflictCriteria := map[string]string{
		"clubs":     "club_id",
		"postings":  "posting_id",
		"users":     "net_id",
		"followers": "id",
		"requests":  "rid",
		"links":     "link_id",
	}

	// build query string
	queryString := "INSERT INTO %s(%s)\n" +
		"VALUES(%s);"
	queryString = fmt.Sprintf(queryString, form["table"], keyString, valueString)

	// if table has a unique columns, add on conflict statement to the query string for
	// upsert functionality
	if _, ok := conflictCriteria[form["table"]]; ok {
		// prepare and execute the command and return
		queryString = queryString[0 : len(queryString)-1]
		queryString += " ON CONFLICT (%s) DO UPDATE SET %s;"
		queryString = fmt.Sprintf(queryString, conflictCriteria[form["table"]], upsertString)

	}

	// Using executeTransaction from the transaction.go file, safely change information in the
	// tables via the use of transactions
	err = executeTransaction(queryString, nil)
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

// From the form specified by the URL, build a string of all the keys,
// a string of all the values and string used for updating data if necessary
// (upsert string)
func buildStrings(values map[string]string) (k, v, u string) {

	// build key and values string + upsert string
	keyString := ""
	valueString := ""
	upsertString := ""
	for key, value := range values {

		if key == "table" {
			continue
		}
		value = strings.Replace(value, "'", "''", -1)
		value = "'" + value + "',"
		valueString += value

		keyvalue := key + " = " + value
		upsertString += keyvalue

		key += ","
		keyString += key
	}
	keyString = keyString[0 : len(keyString)-1]
	valueString = valueString[0 : len(valueString)-1]
	upsertString = upsertString[0 : len(upsertString)-1]
	return keyString, valueString, upsertString
}
