package main

import (
	_ "github.com/lib/pq"
)

// Using the BEGIN PREPARE EXECUTE ROLLBACK IF ERROR OR COMMIT pattern,
// executeTransaction is used by functions that change the data in our database
// such as remove or upsert to ensure ACID compliance via the use of transactions
func executeTransaction(queryString string, args []string) error {

	// begin transaction
	tx, err := db.Begin()
	if err != nil {return err}

	// prepare and rollback if necessary
	stmt, err := tx.Prepare(queryString)
	if err != nil {
		tx.Rollback()
		return err
	}

	// defer stmt closing
	defer stmt.Close()

	// execute and rollback if necessary
	if args == nil {
		_, err = stmt.Exec()
	} else if len(args) == 1{
		_, err = stmt.Exec(args[0])
	} else {
		_, err = stmt.Exec(args[0], args[1])
	}
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
