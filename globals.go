package main


var OK = []byte("{}\n")
var NULL = []byte(": null")

var users [1100000]User
var users_emails = make(map[string]bool)

var locations [1100000]Location

var visits [11000000]Visit
var visits_by_user [11000000]map[int]bool
var visits_by_location [11000000]map[int]bool

var users_counter int
var locations_counter int
var visits_counter int
