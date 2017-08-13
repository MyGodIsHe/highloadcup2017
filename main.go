/*
export GOPATH="$HOME/go"
export GOBIN=$GOPATH/bin

curl -i http://localhost:80/users/2
curl -i http://localhost:80/users/new -d '{"first_name": "Пётр", "last_name": "Фетатосян", "birth_date": -1720915200, "gender": "m", "id": 10, "email": "wibylcudestiwuk@icloud.com"}'

*/
//w.Header().Set("Content-Type", "application/json; charset=utf-8")

package main

import (
	"sort"
	"strings"
	"strconv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"archive/zip"

	"github.com/julienschmidt/httprouter"
)


type User struct {
	Id			int		`json:"id"`
	Email		string	`json:"email"`
	FirstName	string	`json:"first_name"`
	LastName	string	`json:"last_name"`
	BirthDate	int		`json:"birth_date"`
}

type DataUser struct {
	Users	[]User	`json:"users"`
}

type Location struct {
	Id			int		`json:"id"`
	Place		string	`json:"place"`
	Country		string	`json:"country"`
	City		string	`json:"city"`
	Distance	int		`json:"distance"`
}

type DataLocation struct {
	Locations	[]Location	`json:"locations"`
}

type Visit struct {
	Id			int		`json:"id"`
	Location	int		`json:"location"`
	User		int		`json:"user"`
	VisitedAt	int		`json:"visited_at"`
	Mark		int		`json:"mark"`
}

type DataVisit struct {
	Visits	[]Visit	`json:"visits"`
}

type ShortVisit struct {
	Mark		int		`json:"mark"`
	Place		string	`json:"place"`
	VisitedAt	int		`json:"visited_at"`
}

type ShortVisits []ShortVisit
func (s ShortVisits) Len() int {
    return len(s)
}
func (s ShortVisits) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ShortVisits) Less(i, j int) bool {
    return s[i].VisitedAt < s[j].VisitedAt
}

type DataShortVisit struct {
	Visits	ShortVisits	`json:"visits"`
}

var OK = []byte("{}\n")

var users = make(map[int]User)
var users_max_id = 0
var users_emails = make(map[string]bool)

var locations = make(map[int]Location)
var locations_max_id = 0

var visits = make(map[int]Visit)
var visits_max_id = 0


func checkEmail(w http.ResponseWriter,e string) {
	_, ok := users_emails[e]
	if !ok {
		w.WriteHeader(400)
		return
	}
}

func getIntFromQuery(sv string) (string, int, interface{}) {
	var v int
	var err interface{}
	if sv != "" {
		v, err = strconv.Atoi(sv)
	}
	return sv, v, err
}


func loadData(fname string) {
	r, err := zip.OpenReader(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		fmt.Printf("%s loading..", f.Name)
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		if strings.HasPrefix(f.Name, "users") {
			var recs DataUser
			err = json.NewDecoder(rc).Decode(&recs)
			if err != nil {
				log.Fatal(err)
			}
			for _, rec := range recs.Users {
				users[rec.Id] = rec
				users_emails[rec.Email] = true
				if users_max_id < rec.Id {
					users_max_id = rec.Id
				}
			}
		}
		if strings.HasPrefix(f.Name, "locations") {
			var recs DataLocation
			err = json.NewDecoder(rc).Decode(&recs)
			if err != nil {
				log.Fatal(err)
			}
			for _, rec := range recs.Locations {
				locations[rec.Id] = rec
				if locations_max_id < rec.Id {
					locations_max_id = rec.Id
				}
			}
		}
		if strings.HasPrefix(f.Name, "visits") {
			var recs DataVisit
			err = json.NewDecoder(rc).Decode(&recs)
			if err != nil {
				log.Fatal(err)
			}
			for _, rec := range recs.Visits {
				visits[rec.Id] = rec
				if visits_max_id < rec.Id {
					visits_max_id = rec.Id
				}
			}
		}
		rc.Close()
		fmt.Println("done")
	}
}


func main() {
	loadData("/home/elias/projects/traveler-go/data.zip")
	router := httprouter.New()
	router.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	})
	router.POST("/users/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
			users_emails[rec.Email] = true
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
	})
	router.GET("/locations/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	})
	router.POST("/locations/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	})
	router.GET("/visits/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	})
	router.POST("/visits/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	})
	router.GET("/users/:id/visits", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
/*
fromDate - посещения с visited_at > fromDate
toDate - посещения с visited_at < toDate
country - название страны, в которой находятся интересующие достопримечательности
toDistance - возвращать только те места, у которых расстояние от города меньше этого параметра
*/
		var id int
		var err interface{}

		id, err = strconv.Atoi(ps.ByName("id"))
		if err != nil {
			w.WriteHeader(400)
			return
		}

		fromDate, fromDateValue, err := getIntFromQuery(r.URL.Query().Get("fromDate"))
		if err != nil { w.WriteHeader(400); return }

		toDate, toDateValue, err := getIntFromQuery(r.URL.Query().Get("toDate"))
		if err != nil { w.WriteHeader(400); return }

		country := r.URL.Query().Get("country")

		toDistance, toDistanceValue, err := getIntFromQuery(r.URL.Query().Get("toDistance"))
		if err != nil { w.WriteHeader(400); return }

		result := ShortVisits{}
		for _, v := range visits {
			l := locations[v.Location]
			if v.User != id { continue }
			if fromDate != "" && v.VisitedAt <= fromDateValue { continue }
			if toDate != "" && v.VisitedAt >= toDateValue { continue }
			if country != "" && l.Country != country { continue }
			if toDistance != "" && l.Distance >= toDistanceValue { continue }
			result = append(result, ShortVisit{Mark: v.Mark, Place: l.Place, VisitedAt: v.VisitedAt})
		}
		sort.Sort(result)
		json.NewEncoder(w).Encode(DataShortVisit{Visits: result})
	})
	router.GET("/locations/:id/avg", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
/*
fromDate - учитывать оценки только с visited_at > fromDate
toDate - учитывать оценки только с visited_at < toDate
fromAge - учитывать только путешественников, у которых возраст (считается от текущего timestamp) больше этого параметра
toAge - как предыдущее, но наоборот
gender - учитывать оценки только мужчин или женщин
*/

	})
	err := http.ListenAndServe(":80", router)
	if err != nil {
		log.Fatal(err)
	}
}
