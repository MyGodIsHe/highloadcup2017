package main


var OK = []byte("{}\n")

var users = make(map[int]User)
var users_emails = make(map[string]bool)

var locations = make(map[int]Location)

var visits = make(map[int]Visit)
var visits_by_user = make(map[int]map[int]Visit)