/*
export GOPATH="$HOME/go"
export GOBIN=$GOPATH/bin

curl -i http://localhost:80/users/2
curl -i http://localhost:80/users/new -d '{"first_name": "Пётр", "last_name": "Фетатосян", "birth_date": -1720915200, "gender": "m", "id": 10, "email": "wibylcudestiwuk@icloud.com"}'

*/
//w.Header().Set("Content-Type", "application/json; charset=utf-8")

package main

import (
	"strconv"
	"encoding/json"
//	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var OK = []byte("{}\n")

type User struct {
	Id			int		`json:"id"`
	Email		string	`json:"email"`
    FirstName	string	`json:"first_name"`
    LastName	string	`json:"last_name"`
    BirthDate	int		`json:"birth_date"`
}

type Location struct {
	Id			int		`json:"id"`
	Place		string	`json:"place"`
    Country		string	`json:"country"`
    City		string	`json:"city"`
    Distance	int		`json:"distance"`
}

type Visit struct {
	Id			int		`json:"id"`
	Location	string	`json:"location"`
    User		string	`json:"user"`
    VisitedAt	int		`json:"visited_at"`
    Mark		int		`json:"mark"`
}

var users = make(map[int]User)
var users_max_id = 0
var users_emails = make(map[string]bool)

var locations = make(map[int]Location)
var locations_max_id = 0

var visits = make(map[int]Visit)
var visits_max_id = 0

/* USER BEGIN */

func checkEmail(w http.ResponseWriter,e string) {
	_, ok := users_emails[e]
	if !ok {
		w.WriteHeader(400)
		return
	}
}

func getUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(400)
		return
	}
	rec, ok := users[id]
	if !ok {
		w.WriteHeader(404)
		return
	}
	json.NewEncoder(w).Encode(rec)
}

func updateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var id int
	var err interface{}
	var rec User

	err = json.NewDecoder(r.Body).Decode(&rec)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	param := ps.ByName("id")
	if param == "new" {
		checkEmail(w, rec.Email)
		users_max_id++
		id = users_max_id
	} else {
		id, err = strconv.Atoi(param)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		user, ok := users[id]
		if !ok {
			w.WriteHeader(404)
			return
		}
		if user.Email != rec.Email {
			checkEmail(w, rec.Email)
		}
	}
	
	rec.Id = id
	users[id] = rec
	w.Write(OK)
}

/* USER END */


/* LOCATIONS BEGIN */

func getLocation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(400)
		return
	}
	rec, ok := locations[id]
	if !ok {
		w.WriteHeader(404)
		return
	}
	json.NewEncoder(w).Encode(rec)
}

func updateLocation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var id int
	var err interface{}
	var rec Location

	err = json.NewDecoder(r.Body).Decode(&rec)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	param := ps.ByName("id")
	if param == "new" {
		locations_max_id++
		id = locations_max_id
	} else {
		id, err = strconv.Atoi(param)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		_, ok := locations[id]
		if !ok {
			w.WriteHeader(404)
			return
		}
	}
	
	rec.Id = id
	locations[id] = rec
	w.Write(OK)
}

/* LOCATIONS END */


/* VISITS BEGIN */

func getVisit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(400)
		return
	}
	rec, ok := visits[id]
	if !ok {
		w.WriteHeader(404)
		return
	}
	json.NewEncoder(w).Encode(rec)
}

func updateVisit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var id int
	var err interface{}
	var rec Visit

	err = json.NewDecoder(r.Body).Decode(&rec)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	param := ps.ByName("id")
	if param == "new" {
		visits_max_id++
		id = visits_max_id
	} else {
		id, err = strconv.Atoi(param)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		_, ok := visits[id]
		if !ok {
			w.WriteHeader(404)
			return
		}
	}
	
	rec.Id = id
	visits[id] = rec
	w.Write(OK)
}

/* VISITS END */


func main() {
	router := httprouter.New()
	router.GET("/users/:id", getUser)
	router.POST("/users/:id", updateUser)
	router.GET("/locations/:id", getLocation)
	router.POST("/locations/:id", updateLocation)
	router.GET("/visits/:id", getVisit)
	router.POST("/visits/:id", updateVisit)
	http.ListenAndServe(":80", router)
}
