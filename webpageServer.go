package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
)

// Taking in an HTTP.ResponseWriter, a filepath to a html file and a
// struct, write to the ResponseWriter the html file filled in with data
func servePage(w http.ResponseWriter, filepath string, st interface{}) {
	c, err := ioutil.ReadFile("static/"  + filepath)
	s := string(c)
	t := template.New("page")
	t, err = t.Parse(s)
	if err != nil {
		fmt.Println(err)
	}
	err = t.ExecuteTemplate(w, "page", st)
	if err != nil {
		fmt.Println(err)
	}
}
