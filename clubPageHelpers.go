package main

import (
	"fmt"
)

// helper functions for clubpage

// makes given netID follow given clubID
func followClub(netID string, clubID string) {

	// from netID and clubID get person name and clubID name
	var name string
	var clubName string
	stmt, err := db.Prepare("SELECT name FROM users WHERE users.net_id=$1;")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = stmt.QueryRow(netID).Scan(&name)
	if err != nil {
		fmt.Println(err)
		return
	}
	stmt, err = db.Prepare("SELECT name FROM clubs WHERE clubs.club_id=$1;")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = stmt.QueryRow(clubID).Scan(&clubName)
	if err != nil {
		fmt.Println(err)
		return
	}
	// build map from new values and call upsert
	form := map[string]string{
		"table":     "followers",
		"id":        netID + clubID,
		"net_id":    netID,
		"club_id":   clubID,
		"admin":     "f",
		"member":    "f",
		"name":      name,
		"club_name": clubName,
	}
	success, err := upsert(form)
	if !success {
		fmt.Println(err)
	}
}

// removes netID from given clubID
func unfollowClub(netID string, clubID string) {
	// build map from new values and call remove
	form := map[string]string{
		"table": "followers",
		"id":    netID + clubID,
	}
	success, err := remove(form)
	if !success {
		fmt.Println(err)
	}
}

// Adds a request to the Postgres db for netID to be a member of a club
func reqMember(netID string, clubID string) {
	// build map from new values and call upsert
	form := map[string]string{
		"table":   "requests",
		"rid":     netID + clubID,
		"net_id":  netID,
		"club_id": clubID,
		"type":    "member",
	}
	success, err := upsert(form)
	if !success {
		fmt.Println(err)
	}
}

// Adds a request to the Postgres db for netID to be a admin of a club
func reqAdmin(netID string, clubID string) {
	// build map from new values and call upsert
	form := map[string]string{
		"table":   "requests",
		"rid":     netID + clubID,
		"net_id":  netID,
		"club_id": clubID,
		"type":    "admin",
	}
	success, err := upsert(form)
	if !success {
		fmt.Println(err)
	}
}

// makes netID member of given clubID club
func addMember(netID string, clubID string) bool {
	// from netID and clubID get person name and clubID name
	var name string
	var clubName string
	stmt, err := db.Prepare("SELECT name FROM users WHERE users.net_id=$1;")
	if err != nil {
		fmt.Println(err)
		return false
	}
	err = stmt.QueryRow(netID).Scan(&name)
	if err != nil {
		fmt.Println(err)
		return false
	}
	stmt, err = db.Prepare("SELECT name FROM clubs WHERE clubs.club_id=$1;")
	if err != nil {
		fmt.Println(err)
		return false
	}
	err = stmt.QueryRow(clubID).Scan(&clubName)
	if err != nil {
		fmt.Println(err)
		return false
	}
	// build map from new values and call upsert
	form := map[string]string{
		"table":     "followers",
		"id":        netID + clubID,
		"net_id":    netID,
		"club_id":   clubID,
		"admin":     "f",
		"member":    "t",
		"name":      name,
		"club_name": clubName,
	}
	success, err := upsert(form)
	if !success {
		fmt.Println(err)
		return false
	}
	return true
}

// makes netID member of given clubID club
func leaveMember(netID string, clubID string) {
	// from netID and clubID get person name and clubID name
	var name string
	var clubName string
	stmt, err := db.Prepare("SELECT name FROM users WHERE users.net_id=$1;")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = stmt.QueryRow(netID).Scan(&name)
	if err != nil {
		fmt.Println(err)
		return
	}
	stmt, err = db.Prepare("SELECT name FROM clubs WHERE clubs.club_id=$1;")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = stmt.QueryRow(clubID).Scan(&clubName)
	if err != nil {
		fmt.Println(err)
		return
	}
	// build map from new values and call upsert
	form := map[string]string{
		"table":     "followers",
		"id":        netID + clubID,
		"net_id":    netID,
		"club_id":   clubID,
		"admin":     "f",
		"member":    "f",
		"name":      name,
		"club_name": clubName,
	}
	success, err := upsert(form)
	if !success {
		fmt.Println(err)
	}
}

// makes netID admin of given clubID club
func addAdmin(netID string, clubID string) bool {
	// from netID and clubID get person name and clubID name
	var name string
	var clubName string
	stmt, err := db.Prepare("SELECT name FROM users WHERE users.net_id=$1;")
	if err != nil {
		fmt.Println(err)
		return false
	}
	err = stmt.QueryRow(netID).Scan(&name)
	if err != nil {
		fmt.Println(err)
		return false
	}
	stmt, err = db.Prepare("SELECT name FROM clubs WHERE clubs.club_id=$1;")
	if err != nil {
		fmt.Println(err)
		return false
	}
	err = stmt.QueryRow(clubID).Scan(&clubName)
	if err != nil {
		fmt.Println(err)
		return false
	}
	// build map from new values and call upsert
	form := map[string]string{
		"table":     "followers",
		"id":        netID + clubID,
		"net_id":    netID,
		"club_id":   clubID,
		"admin":     "t",
		"member":    "t",
		"name":      name,
		"club_name": clubName,
	}
	success, err := upsert(form)
	if !success {
		fmt.Println("upsert error")
		fmt.Println(err)
		return false
	}
	return true
}
