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

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)


func main() {
	loadData("/tmp/data/data.zip")

	router := fasthttprouter.New()

	router.GET("/users/:id", func(ctx *fasthttp.RequestCtx) {
		id, err := strconv.Atoi(ctx.UserValue("id").(string))
		if err != nil {
			ctx.SetStatusCode(404)
			return
		}
		rec, ok := users[id]
		if !ok {
			ctx.SetStatusCode(404)
			return
		}
		json.NewEncoder(ctx).Encode(rec)
	})

	router.POST("/users/:id", func(ctx *fasthttp.RequestCtx) {
		var id int
		var err interface{}
		var rec User

		param := ctx.UserValue("id").(string)
		is_insert := param == "new"
		if !is_insert {
			id, err = strconv.Atoi(param)
			if err != nil {
				ctx.SetStatusCode(404)
				return
			}
			var ok bool
			rec, ok = users[id]
			if !ok {
				ctx.SetStatusCode(404)
				return
			}
		}

		old_email := rec.Email

		if !updateUser(ctx, &rec, is_insert) {
			ctx.SetStatusCode(400)
			return
		}

		if is_insert {
			_, ok := users_emails[rec.Email]
			if ok {
				ctx.SetStatusCode(400)
				return
			}
			users_emails[rec.Email] = true
		} else {
			if old_email != rec.Email {
				_, ok := users_emails[rec.Email]
				if ok {
					ctx.SetStatusCode(400)
					return
				}

				delete(users_emails, old_email)
				users_emails[rec.Email] = true
			}
		}

		users[rec.Id] = rec

		ctx.Write(OK)
	})

	router.GET("/locations/:id", func(ctx *fasthttp.RequestCtx) {
		id, err := strconv.Atoi(ctx.UserValue("id").(string))
		if err != nil {
			ctx.SetStatusCode(404)
			return
		}
		rec, ok := locations[id]
		if !ok {
			ctx.SetStatusCode(404)
			return
		}
		json.NewEncoder(ctx).Encode(rec)
	})

	router.POST("/locations/:id", func(ctx *fasthttp.RequestCtx) {
		var id int
		var err interface{}
		var rec Location

		param := ctx.UserValue("id").(string)
		is_insert := param == "new"
		if !is_insert {
			id, err = strconv.Atoi(param)
			if err != nil {
				ctx.SetStatusCode(404)
				return
			}
			var ok bool
			rec, ok = locations[id]
			if !ok {
				ctx.SetStatusCode(404)
				return
			}
		}

		if !updateLocation(ctx, &rec, is_insert) {
			ctx.SetStatusCode(400)
			return
		}

		locations[rec.Id] = rec

		ctx.Write(OK)
	})

	router.GET("/visits/:id", func(ctx *fasthttp.RequestCtx) {
		id, err := strconv.Atoi(ctx.UserValue("id").(string))
		if err != nil {
			ctx.SetStatusCode(404)
			return
		}
		rec, ok := visits[id]
		if !ok {
			ctx.SetStatusCode(404)
			return
		}
		json.NewEncoder(ctx).Encode(rec)
	})

	router.POST("/visits/:id", func(ctx *fasthttp.RequestCtx) {
		var id int
		var err interface{}
		var rec Visit

		param := ctx.UserValue("id").(string)
		is_insert := param == "new"
		if !is_insert {
			id, err = strconv.Atoi(param)
			if err != nil {
				ctx.SetStatusCode(404)
				return
			}
			var ok bool
			rec, ok = visits[id]
			if !ok {
				ctx.SetStatusCode(404)
				return
			}
		}

		if !updateVisit(ctx, &rec, is_insert) {
			ctx.SetStatusCode(400)
			return
		}

		visitSetEvent(rec)

		ctx.Write(OK)
	})

	router.GET("/users/:id/visits", func(ctx *fasthttp.RequestCtx) {
		var id int
		var err interface{}

		id, err = strconv.Atoi(ctx.UserValue("id").(string))
		if err != nil {
			ctx.SetStatusCode(404)
			return
		}

		_, ok := users[id]
		if !ok {
			ctx.SetStatusCode(404)
			return
		}

		hasFromDate, fromDateValue, err := getIntFromQuery(ctx,"fromDate")
		if err != nil {
			ctx.SetStatusCode(400);
			return
		}

		hasToDate, toDateValue, err := getIntFromQuery(ctx, "toDate")
		if err != nil {
			ctx.SetStatusCode(400);
			return
		}

		country := string(ctx.URI().QueryArgs().Peek("country"))
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
					ctx.SetStatusCode(404)
					return
				}
			}
		}*/

		ctx.URI().QueryArgs().GetUintOrZero("toDistance")
		hasToDistance, toDistanceValue, err := getIntFromQuery(ctx, "toDistance")
		if err != nil {
			ctx.SetStatusCode(400);
			return
		}

		result := ShortVisits{}
		for _, v := range visits_by_user[id] {
			if hasFromDate && v.VisitedAt <= fromDateValue {
				continue
			}
			if hasToDate && v.VisitedAt >= toDateValue {
				continue
			}
			l = locations[v.Location]
			if country != "" && l.Country != country {
				continue
			}
			//if country == "" {
			//	l = locations[v.Location]
			//} else if v.Location != l.Id { continue }
			if hasToDistance && l.Distance >= toDistanceValue {
				continue
			}
			result = append(result, ShortVisit{Mark: v.Mark, Place: l.Place, VisitedAt: v.VisitedAt})
		}
		sort.Sort(result)
		json.NewEncoder(ctx).Encode(DataShortVisit{Visits: result})
	})

	router.GET("/locations/:id/avg", func(ctx *fasthttp.RequestCtx) {
		var id int
		var err interface{}

		id, err = strconv.Atoi(ctx.UserValue("id").(string))
		if err != nil {
			ctx.SetStatusCode(404)
			return
		}

		_, ok := locations[id]
		if !ok {
			ctx.SetStatusCode(404)
			return
		}


		hasFromDate, fromDateValue, err := getIntFromQuery(ctx, "fromDate")
		if err != nil {
			ctx.SetStatusCode(400);
			return
		}

		hasToDate, toDateValue, err := getIntFromQuery(ctx, "toDate")
		if err != nil {
			ctx.SetStatusCode(400);
			return
		}

		hasFromAge, fromAgeValue, err := getIntFromQuery(ctx, "fromAge")
		if err != nil {
			ctx.SetStatusCode(400);
			return
		}

		hasToAge, toAgeValue, err := getIntFromQuery(ctx, "toAge")
		if err != nil {
			ctx.SetStatusCode(400);
			return
		}

		gender := string(ctx.URI().QueryArgs().Peek("gender"))
		if gender != "" && gender != "f" && gender != "m" {
			ctx.SetStatusCode(400)
			return
		}

		now := time.Now().UTC()

		avgCount := 0
		avgSum := 0
		for _, v := range visits {
			if v.Location != id {
				continue
			}
			if hasFromDate && v.VisitedAt <= fromDateValue {
				continue
			}
			if hasToDate && v.VisitedAt >= toDateValue {
				continue
			}
			u := users[v.User]
			if gender != "" && u.Gender != gender {
				continue
			}
			age := diff(time.Unix(int64(u.BirthDate), 0).UTC(), now)
			if hasFromAge && age < fromAgeValue {
				continue
			}
			if hasToAge && age >= toAgeValue {
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
		json.NewEncoder(ctx).Encode(DataAvg{Avg: avg})
	})

	fmt.Println("Good luck ^-^")

	server := fasthttp.Server{
		Handler: router.Handler,
		//WriteTimeout: 2*time.Second,
	}
	err := server.ListenAndServe(":80")
	if err != nil {
		log.Fatal(err)
	}
}
