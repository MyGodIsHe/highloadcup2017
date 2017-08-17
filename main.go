/*
export GOPATH="$HOME/go"
export GOBIN=$GOPATH/bin

curl -i http://localhost:80/users/2
curl -i http://localhost:80/users/new -d '{"first_name": "Пётр", "last_name": "Фетатосян", "birth_date": -1720915200, "gender": "m", "id": 10, "email": "wibylcudestiwuk@icloud.com"}'

*/
//w.Header().Set("Content-Type", "application/json; charset=utf-8")

package main

import (
	"time"
	"sort"
	"strings"
	"strconv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"archive/zip"

	"io"

	"github.com/julienschmidt/httprouter"
	"github.com/bearbin/go-age"
)

func parseId(dict map[string]interface{}) bool {
	v, ok := dict["id"]
    if ok && v == nil {
        return false
    }
    return true
}

func parseString(dict map[string]interface{}, value *string, name string, required bool) bool {
	v, ok := dict[name]
    if ok && v == nil {
        return false
    }
    if !ok {
    	return !required
    }
    *value = v.(string)
    return true
}

func parseInt(dict map[string]interface{}, value *int, name string, required bool) bool {
	v, ok := dict[name]
    if ok && v == nil {
        return false
    }
    if !ok {
        return !required
    }
    *value = int(v.(float64))
    return true
}


type User struct {
	Id			int		`json:"id"`
	Email		string	`json:"email"`
	FirstName	string	`json:"first_name"`
	LastName	string	`json:"last_name"`
	Gender		string	`json:"gender"`
	BirthDate	int		`json:"birth_date"`
}

func updateUser(body io.Reader, rec *User, required bool) bool {
	dict := make(map[string]interface{})
	if err := json.NewDecoder(body).Decode(&dict); err != nil {
        return false
    }

    if required && !parseId(dict) { return false }
    if !parseString(dict, &rec.Email, "email", required) || len(rec.Email) > 100 { return false }
    if !parseString(dict, &rec.FirstName, "first_name", required) || len(rec.FirstName) > 50 { return false }
    if !parseString(dict, &rec.LastName, "last_name", required) || len(rec.LastName) > 50 { return false }
    if !parseString(dict, &rec.Gender, "gender", required) || (rec.Gender != "f" && rec.Gender != "m") { return false }
    if !parseInt(dict, &rec.BirthDate, "birth_date", required) || (rec.BirthDate < -1262325600 || rec.BirthDate > 915123600) { return false }
    return true
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

func updateLocation(body io.Reader, rec *Location, required bool) bool {
	dict := make(map[string]interface{})
	if err := json.NewDecoder(body).Decode(&dict); err != nil {
        return false
    }

    if required && !parseId(dict) { return false }
    if !parseString(dict, &rec.Place, "place", required) { return false }
    if !parseString(dict, &rec.Country, "country", required) || len(rec.Country) > 50 { return false }
    if !parseString(dict, &rec.City, "city", required) || len(rec.City) > 50 { return false }
    if !parseInt(dict, &rec.Distance, "distance", required) { return false }
    return true
}

type DataLocation struct {
	Locations	[]Location	`json:"locations"`
}

type Visit struct {
	Id			int	`json:"id"`
	Location	int	`json:"location"`
	User		int	`json:"user"`
	VisitedAt	int	`json:"visited_at"`
	Mark		int	`json:"mark"`
}

func updateVisit(body io.Reader, rec *Visit, required bool) bool {
	dict := make(map[string]interface{})
	if err := json.NewDecoder(body).Decode(&dict); err != nil {
        return false
    }

    if required && !parseId(dict) { return false }
    if !parseInt(dict, &rec.Location, "location", required) { return false }
    if !parseInt(dict, &rec.User, "user", required) { return false }
    if !parseInt(dict, &rec.VisitedAt, "visited_at", required) || (rec.VisitedAt < 946659600 || rec.VisitedAt > 1420045200) { return false }
    if !parseInt(dict, &rec.Mark, "mark", required) || (rec.Mark < 0 || rec.Mark > 5) { return false }
    return true
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

type DataAvg struct {
	Avg		float64	`json:"avg"`
}

var OK = []byte("{}\n")

var users = make(map[int]User)
var users_max_id = 0
var users_emails = make(map[string]bool)

var locations = make(map[int]Location)
var locations_max_id = 0

var visits = make(map[int]Visit)
var visits_max_id = 0


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
	loadData("/tmp/data/data.zip")
	router := httprouter.New()
	router.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			w.WriteHeader(404)
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

		param := ps.ByName("id")
		is_insert := param == "new"
		if !is_insert {
			id, err = strconv.Atoi(param)
			if err != nil {
				w.WriteHeader(404)
				return
			}
			var ok bool
			rec, ok = users[id]
			if !ok {
				w.WriteHeader(404)
				return
			}
		}

		old_email := rec.Email

		if !updateUser(r.Body, &rec, is_insert) {
			w.WriteHeader(400)
			return
		}

		if is_insert {
			_, ok := users_emails[rec.Email]
			if ok {
				w.WriteHeader(400)
				return
			}
			users_max_id++
			id = users_max_id
			users_emails[rec.Email] = true
			// insert
			rec.Id = id
		} else {
			if old_email != rec.Email {
				_, ok := users_emails[rec.Email]
				if ok {
					w.WriteHeader(400)
					return
				}

				delete(users_emails, old_email)
				users_emails[rec.Email] = true
			}
		}
		users[id] = rec

		w.Write(OK)
	})
	router.GET("/locations/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			w.WriteHeader(404)
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

		param := ps.ByName("id")
		is_insert := param == "new"
		if !is_insert {
			id, err = strconv.Atoi(param)
			if err != nil {
				w.WriteHeader(404)
				return
			}
			var ok bool
			rec, ok = locations[id]
			if !ok {
				w.WriteHeader(404)
				return
			}
		}

		if !updateLocation(r.Body, &rec, is_insert) {
			w.WriteHeader(400)
			return
		}

		if is_insert {
			locations_max_id++
			id = locations_max_id
			rec.Id = id
		}
		locations[id] = rec

		w.Write(OK)
	})
	router.GET("/visits/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			w.WriteHeader(404)
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

		param := ps.ByName("id")
		is_insert := param == "new"
		if !is_insert {
			id, err = strconv.Atoi(param)
			if err != nil {
				w.WriteHeader(404)
				return
			}
			var ok bool
			rec, ok = visits[id]
			if !ok {
				w.WriteHeader(404)
				return
			}
		}

		if !updateVisit(r.Body, &rec, is_insert) {
			w.WriteHeader(400)
			return
		}

		if is_insert {
			visits_max_id++
			id = visits_max_id
			rec.Id = id
		}
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
			w.WriteHeader(404)
			return
		}

		_, ok := users[id]
		if !ok {
			w.WriteHeader(404)
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
		var id int
		var err interface{}

		id, err = strconv.Atoi(ps.ByName("id"))
		if err != nil {
			w.WriteHeader(404)
			return
		}

		_, ok := locations[id]
		if !ok {
			w.WriteHeader(404)
			return
		}

		fromDate, fromDateValue, err := getIntFromQuery(r.URL.Query().Get("fromDate"))
		if err != nil { w.WriteHeader(400); return }

		toDate, toDateValue, err := getIntFromQuery(r.URL.Query().Get("toDate"))
		if err != nil { w.WriteHeader(400); return }

		fromAge, fromAgeValue, err := getIntFromQuery(r.URL.Query().Get("fromAge"))
		if err != nil { w.WriteHeader(400); return }

		toAge, toAgeValue, err := getIntFromQuery(r.URL.Query().Get("toAge"))
		if err != nil { w.WriteHeader(400); return }

		gender := r.URL.Query().Get("gender")
		if gender != "" && gender != "f" && gender != "m" {
			w.WriteHeader(400)
			return
		}

		now := time.Now()

		avgCount := 0
		avgSum := 0
		for _, v := range visits {
			if v.Location != id { continue }
			if fromDate != "" && v.VisitedAt <= fromDateValue { continue }
			if toDate != "" && v.VisitedAt >= toDateValue { continue }
			u := users[v.User]
			if gender != "" && u.Gender != gender { continue }
			age := age.AgeAt(time.Unix(int64(u.BirthDate), 0), now)
			if fromAge != "" && age <= fromAgeValue { continue }
			if toAge != "" && age >= toAgeValue { continue }
			avgCount++
			avgSum += v.Mark
		}
		var avg float64
		if avgCount != 0 {
			avg = float64(avgSum)/float64(avgCount)
		}
		avg, _ = strconv.ParseFloat(fmt.Sprintf("%.5f", avg), 64)
		json.NewEncoder(w).Encode(DataAvg{Avg: avg})
	})
	err := http.ListenAndServe(":80", router)
	if err != nil {
		log.Fatal(err)
	}
}
