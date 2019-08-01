package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gbrlsnchs/jwt"
)

// the struct that holds the token
type Token struct {
	*jwt.JWT
	IsLoggedIn bool   `json:"IsLoggedIn"`
	NetID      string `json:"NetID"`
}

// Issues a token for the given netID
func getToken(netID string) (token []byte, err error) {

	now := time.Now()
	accessToken := &Token{
		JWT: &jwt.JWT{
			Issuer:         "TigerPages",
			Subject:        netID,
			Audience:       "TigerPages",
			ExpirationTime: now.Add(200 * time.Minute).Unix(),
			IssuedAt:       now.Unix(),
			ID:             "AccessToken",
		},
		IsLoggedIn: true,
		NetID:      netID,
	}
	accessToken.SetAlgorithm(hs256)

	// marshal to get payload
	payload, err := jwt.Marshal(accessToken)
	if err != nil {
		return nil, err
	}

	// sign off on payload to get token
	token, err = hs256.Sign(payload)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// verifies a the accessToken oookie contains a valid JWT
func validateUser(w http.ResponseWriter, r *http.Request) (valid bool, netID string) {

	// first get the cookie for the token
	tokenCookie, err := r.Cookie(tokenName)
	if err != nil {
		return false, ""
	}

	// parse the token to get payload, signature and if err occurred
	payload, sig, err := jwt.Parse(tokenCookie.Value)
	if err != nil {
		return false, ""
	}

	// Verify token
	if err = hs256.Verify(payload, sig); err != nil {
		return false, ""
	}

	// unmarshal it to get Token struct
	var token Token
	if err = jwt.Unmarshal(payload, &token); err != nil {
		return false, ""
	}

	// Validate fields
	iatValidator := jwt.IssuedAtValidator(time.Now())
	expValidator := jwt.ExpirationTimeValidator(time.Now())
	audValidator := jwt.AudienceValidator("TigerPages")
	if err = token.Validate(iatValidator, expValidator, audValidator); err != nil {
		switch err {
		case jwt.ErrIatValidation:
			fmt.Println(token.NetID + ": " + err.Error())
			http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		case jwt.ErrExpValidation:
			fmt.Println(token.NetID + ": " + err.Error())
			http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)

		case jwt.ErrAudValidation:
			fmt.Println(token.NetID + ": " + err.Error())
			http.Redirect(w, r, "/sessiontimeout/", http.StatusFound)
		}
	}
	return true, token.NetID
}
