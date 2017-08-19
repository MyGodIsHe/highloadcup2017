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
	"strconv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)


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
		cacheStorage = make(map[string]Cache)

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
			users_emails[rec.Email] = true
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

		users[rec.Id] = rec

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
		cacheStorage = make(map[string]Cache)

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

		locations[rec.Id] = rec

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
		cacheStorage = make(map[string]Cache)

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

		visitSetEvent(rec)

		w.Write(OK)
	})

	router.GET("/users/:id/visits", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
		if err != nil {
			w.WriteHeader(400);
			return
		}

		toDate, toDateValue, err := getIntFromQuery(r.URL.Query().Get("toDate"))
		if err != nil {
			w.WriteHeader(400);
			return
		}

		country := r.URL.Query().Get("country")
		var l Location
		/*{
			is_found := false
			if country != "" {
				for _, x := range locations {
					if x.Country == country {
						l = x
						is_found = true
						break
					}
				}
				if !is_found {
					w.WriteHeader(404)
					return
				}
			}
		}*/

		toDistance, toDistanceValue, err := getIntFromQuery(r.URL.Query().Get("toDistance"))
		if err != nil {
			w.WriteHeader(400);
			return
		}

		result := ShortVisits{}
		for _, v := range visits_by_user[id] {
			if fromDate != "" && v.VisitedAt <= fromDateValue {
				continue
			}
			if toDate != "" && v.VisitedAt >= toDateValue {
				continue
			}
			l = locations[v.Location]
			if country != "" && l.Country != country {
				continue
			}
			//if country == "" {
			//	l = locations[v.Location]
			//} else if v.Location != l.Id { continue }
			if toDistance != "" && l.Distance >= toDistanceValue {
				continue
			}
			result = append(result, ShortVisit{Mark: v.Mark, Place: l.Place, VisitedAt: v.VisitedAt})
		}
		sort.Sort(result)
		json.NewEncoder(w).Encode(DataShortVisit{Visits: result})
	})

	router.GET("/locations/:id/avg", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
		if err != nil {
			w.WriteHeader(400);
			return
		}

		toDate, toDateValue, err := getIntFromQuery(r.URL.Query().Get("toDate"))
		if err != nil {
			w.WriteHeader(400);
			return
		}

		fromAge, fromAgeValue, err := getIntFromQuery(r.URL.Query().Get("fromAge"))
		if err != nil {
			w.WriteHeader(400);
			return
		}

		toAge, toAgeValue, err := getIntFromQuery(r.URL.Query().Get("toAge"))
		if err != nil {
			w.WriteHeader(400);
			return
		}

		gender := r.URL.Query().Get("gender")
		if gender != "" && gender != "f" && gender != "m" {
			w.WriteHeader(400)
			return
		}

		now := time.Now().UTC()

		avgCount := 0
		avgSum := 0
		for _, v := range visits {
			if v.Location != id {
				continue
			}
			if fromDate != "" && v.VisitedAt <= fromDateValue {
				continue
			}
			if toDate != "" && v.VisitedAt >= toDateValue {
				continue
			}
			u := users[v.User]
			if gender != "" && u.Gender != gender {
				continue
			}
			age := diff(time.Unix(int64(u.BirthDate), 0).UTC(), now)
			if fromAge != "" && age < fromAgeValue {
				continue
			}
			if toAge != "" && age >= toAgeValue {
				continue
			}
			avgCount++
			avgSum += v.Mark
		}
		var avg float64
		if avgCount != 0 {
			avg = float64(avgSum) / float64(avgCount)
		}
		avg, _ = strconv.ParseFloat(fmt.Sprintf("%.5f", avg), 64)
		//fmt.Println("avg", r.URL.String(), avg, )
		json.NewEncoder(w).Encode(DataAvg{Avg: avg})
	})

	fmt.Println("Good luck ^-^")

	err := http.ListenAndServe(":80", router)
	if err != nil {
		log.Fatal(err)
	}
}
