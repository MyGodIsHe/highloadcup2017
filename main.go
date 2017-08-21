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
	"github.com/patrickmn/go-cache"
)

var cs *cache.Cache

func main() {
	cs = cache.New(cache.NoExpiration, cache.NoExpiration)

	loadData("/tmp/data/data.zip")

	router := fasthttprouter.New()

	router.GET("/users/:id", func(ctx *fasthttp.RequestCtx) {
		id, err := strconv.Atoi(ctx.UserValue("id").(string))
		if err != nil {
			ctx.SetStatusCode(404)
			return
		}
		rec, ok := users.Load(id)
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
			r, ok := users.Load(id)
			if !ok {
				ctx.SetStatusCode(404)
				return
			}
			rec = r.(User)
		}

		old_email := rec.Email

		if !updateUser(ctx, &rec, is_insert) {
			ctx.SetStatusCode(400)
			return
		}

		if is_insert {
			_, ok := users_emails.Load(rec.Email)
			if ok {
				ctx.SetStatusCode(400)
				return
			}
			users_emails.Store(rec.Email, true)
		} else {
			if old_email != rec.Email {
				_, ok := users_emails.Load(rec.Email)
				if ok {
					ctx.SetStatusCode(400)
					return
				}

				users_emails.Delete(old_email)
				users_emails.Store(rec.Email, true)
			}
		}

		users.Store(rec.Id, rec)

		ctx.Write(OK)
		ctx.SetConnectionClose()
	})

	router.GET("/locations/:id", func(ctx *fasthttp.RequestCtx) {
		id, err := strconv.Atoi(ctx.UserValue("id").(string))
		if err != nil {
			ctx.SetStatusCode(404)
			return
		}
		rec, ok := locations.Load(id)
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
			r, ok := locations.Load(id)
			if !ok {
				ctx.SetStatusCode(404)
				return
			}
			rec = r.(Location)
		}

		if !updateLocation(ctx, &rec, is_insert) {
			ctx.SetStatusCode(400)
			return
		}

		locations.Store(rec.Id, rec)

		ctx.Write(OK)
		ctx.SetConnectionClose()
	})

	router.GET("/visits/:id", func(ctx *fasthttp.RequestCtx) {
		id, err := strconv.Atoi(ctx.UserValue("id").(string))
		if err != nil {
			ctx.SetStatusCode(404)
			return
		}
		rec, ok := visits.Load(id)
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
			r, ok := visits.Load(id)
			if !ok {
				ctx.SetStatusCode(404)
				return
			}
			rec = r.(Visit)
		}

		if !updateVisit(ctx, &rec, is_insert) {
			ctx.SetStatusCode(400)
			return
		}

		visitSetEvent(rec)

		ctx.Write(OK)
		ctx.SetConnectionClose()
	})

	router.GET("/users/:id/visits", func(ctx *fasthttp.RequestCtx) {
		var id int
		var err interface{}

		id, err = strconv.Atoi(ctx.UserValue("id").(string))
		if err != nil {
			ctx.SetStatusCode(404)
			return
		}

		_, ok := users.Load(id)
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

		ctx.URI().QueryArgs().GetUintOrZero("toDistance")
		hasToDistance, toDistanceValue, err := getIntFromQuery(ctx, "toDistance")
		if err != nil {
			ctx.SetStatusCode(400);
			return
		}

		var l Location
		var v Visit
		result := ShortVisits{}
		items, ok := visits_by_user.Load(id)
		if ok {
			items.(*Map).Range(func(_, value interface{}) bool {
				v = value.(Visit)
				if hasFromDate && v.VisitedAt <= fromDateValue {
					return true
				}
				if hasToDate && v.VisitedAt >= toDateValue {
					return true
				}
				l_raw, _ := locations.Load(v.Location)
				l = l_raw.(Location)
				if country != "" && l.Country != country {
					return true
				}
				if hasToDistance && l.Distance >= toDistanceValue {
					return true
				}
				result = append(result, ShortVisit{Mark: v.Mark, Place: l.Place, VisitedAt: v.VisitedAt})

				return true
			})
			sort.Sort(result)
		}
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

		_, ok := locations.Load(id)
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

		var v Visit
		avgCount := 0
		avgSum := 0
		visits.Range(func(_, value interface{}) bool {
			v = value.(Visit)
			if v.Location != id {
				return true
			}
			if hasFromDate && v.VisitedAt <= fromDateValue {
				return true
			}
			if hasToDate && v.VisitedAt >= toDateValue {
				return true
			}
			u_raw, _ := users.Load(v.User)
			u := u_raw.(User)
			if gender != "" && u.Gender != gender {
				return true
			}
			age := diff(time.Unix(int64(u.BirthDate), 0).UTC(), now)
			if hasFromAge && age < fromAgeValue {
				return true
			}
			if hasToAge && age >= toAgeValue {
				return true
			}
			avgCount++
			avgSum += v.Mark
			return true
		})
		var avg float64
		if avgCount != 0 {
			avg = float64(avgSum) / float64(avgCount)
		}
		avg, _ = strconv.ParseFloat(fmt.Sprintf("%.5f", avg), 64)
		json.NewEncoder(ctx).Encode(DataAvg{Avg: avg})
	})

	fmt.Println("Good luck ^-^")

	server := fasthttp.Server{
		Handler: CacheHandlerFunc(router.Handler),
		//Handler: router.Handler,
	}
	err := server.ListenAndServe(":80")
	if err != nil {
		log.Fatal(err)
	}
}
