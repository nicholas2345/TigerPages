package main

// a collection of functions that check things

// Checks if the given user is an admin of the page. Takes in clubID and netID and returns bool
func isAdmin(clubID string, netID string) bool {
	stmt, err := db.Prepare("SELECT net_id, admin FROM followers WHERE followers.club_id=$1;")
	if err != nil {
		return false
	}
	rows, err := stmt.Query(clubID)
	defer rows.Close()
	// iterate through all returned rows. If netID is an admin, return true
	for rows.Next() {
		var user string
		var admin bool

		err := rows.Scan(&user, &admin)
		if err != nil {
			return false
		}

		if netID == user && admin {
			return true
		}
	}

	// if it gets to here, given netID was not an admin of club with given clubID. Return false
	return false
}
