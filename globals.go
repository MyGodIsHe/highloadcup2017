package main


var OK = []byte("{}\n")
var NULL = []byte(": null")

var users [10000000]User
var users_emails = make(map[string]bool)

var locations [10000000]Location

var visits [10000000]Visit
var visits_by_user = make(map[int]map[int]Visit)
var visits_by_location = make(map[int]map[int]Visit)
