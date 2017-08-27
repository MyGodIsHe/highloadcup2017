package main

var OK = []byte("{}\n")
var NULL = []byte(": null")

var users [1100000]User
var users_emails = make(map[string]bool)

var locations [1100000]Location

var visits [11000000]Visit
var visits_by_user [1100000][]int
var visits_by_location [1100000][]int

var users_counter int
var locations_counter int
var visits_counter int
