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
	"strconv"
	"fmt"
	"log"
	"bytes"

	"github.com/valyala/fasthttp"
	"sort"
)

func users_get(ctx *fasthttp.RequestCtx, id int) {
	rec := users[id]
	if rec.Id == 0 {
		ctx.SetStatusCode(404)
		return
	}
	rec.Write(ctx)
}

func users_create(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	if len(body) == 0 || bytes.Contains(body, NULL) {
		ctx.SetStatusCode(400)
		ctx.SetConnectionClose()
		return
	}
	ctx.Write(OK)
	ctx.SetConnectionClose()

	var rec User

	if !rec.Update(body, true) {
		ctx.SetStatusCode(400)
		return
	}

	_, ok := users_emails[rec.Email]
	if ok {
		ctx.SetStatusCode(400)
		return
	}
	users_emails[rec.Email] = true
	users[rec.Id] = rec
}

func users_update(ctx *fasthttp.RequestCtx, id int) {
	body := ctx.PostBody()
	if len(body) == 0 || bytes.Contains(body, NULL) {
		ctx.SetStatusCode(400)
		ctx.SetConnectionClose()
		return
	}
	ctx.Write(OK)
	ctx.SetConnectionClose()

	rec := users[id]
	if rec.Id == 0 {
		ctx.SetStatusCode(404)
		return
	}

	old_email := rec.Email

	if !rec.Update(body, false) {
		ctx.SetStatusCode(400)
		return
	}

	if old_email != rec.Email {
		_, ok := users_emails[rec.Email]
		if ok {
			ctx.SetStatusCode(400)
			return
		}

		delete(users_emails, old_email)
		users_emails[rec.Email] = true
	}
	users[rec.Id] = rec
}

func locations_get(ctx *fasthttp.RequestCtx, id int) {
	rec := locations[id]
	if rec.Id == 0 {
		ctx.SetStatusCode(404)
		return
	}
	rec.Write(ctx)
}

func locations_create(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	if len(body) == 0 || bytes.Contains(body, NULL) {
		ctx.SetStatusCode(400)
		ctx.SetConnectionClose()
		return
	}

	ctx.Write(OK)
	ctx.SetConnectionClose()

	var rec Location
	if !rec.Update(body, true) {
		ctx.SetStatusCode(400)
		return
	}

	locations[rec.Id] = rec
}

func locations_update(ctx *fasthttp.RequestCtx, id int) {
	body := ctx.PostBody()
	if len(body) == 0 || bytes.Contains(body, NULL) {
		ctx.SetStatusCode(400)
		ctx.SetConnectionClose()
		return
	}

	ctx.Write(OK)
	ctx.SetConnectionClose()

	rec := locations[id]
	if rec.Id == 0 {
		ctx.SetStatusCode(404)
		return
	}

	if !rec.Update(body, false) {
		ctx.SetStatusCode(400)
		return
	}

	locations[rec.Id] = rec
}

func visits_get(ctx *fasthttp.RequestCtx, id int) {
	rec := visits[id]
	if rec.Id == 0 {
		ctx.SetStatusCode(404)
		return
	}
	rec.Write(ctx)
}

func visits_create(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	if len(body) == 0 || bytes.Contains(body, NULL) {
		ctx.SetStatusCode(400)
		ctx.SetConnectionClose()
		return
	}

	ctx.Write(OK)
	ctx.SetConnectionClose()

	var rec Visit
	if !rec.Update(body, true) {
		ctx.SetStatusCode(400)
		return
	}

	go visitSetEvent(rec)
}

func visits_update(ctx *fasthttp.RequestCtx, id int) {
	body := ctx.PostBody()
	if len(body) == 0 || bytes.Contains(body, NULL) {
		ctx.SetStatusCode(400)
		ctx.SetConnectionClose()
		return
	}

	ctx.Write(OK)
	ctx.SetConnectionClose()

	rec := visits[id]
	if rec.Id == 0 {
		ctx.SetStatusCode(404)
		return
	}

	if !rec.Update(body, false) {
		ctx.SetStatusCode(400)
		return
	}

	go visitSetEvent(rec)
}

func users_visits(ctx *fasthttp.RequestCtx, id int) {
	var err interface{}

	rec := users[id]
	if rec.Id == 0 {
		ctx.SetStatusCode(404)
		return
	}

	hasFromDate, fromDateValue, err := getIntFromQuery(ctx,"fromDate")
	if err != nil {
		ctx.SetStatusCode(400)
		return
	}

	hasToDate, toDateValue, err := getIntFromQuery(ctx, "toDate")
	if err != nil {
		ctx.SetStatusCode(400)
		return
	}

	country := string(ctx.URI().QueryArgs().Peek("country"))
	var l Location

	ctx.URI().QueryArgs().GetUintOrZero("toDistance")
	hasToDistance, toDistanceValue, err := getIntFromQuery(ctx, "toDistance")
	if err != nil {
		ctx.SetStatusCode(400)
		return
	}

	result := ShortVisits{}
	for _, vid := range visits_by_user[id] {
		v := visits[vid]
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
		if hasToDistance && l.Distance >= toDistanceValue {
			continue
		}
		result = append(result, ShortVisit{Mark: v.Mark, Place: l.Place, VisitedAt: v.VisitedAt})
	}
	sort.Sort(result)
	WriteShortVisits(ctx, result)
	//json.NewEncoder(ctx).Encode(DataShortVisit{Visits: result})
}

func locations_avg(ctx *fasthttp.RequestCtx, id int) {
	var err interface{}
	rec := locations[id]
	if rec.Id == 0 {
		ctx.SetStatusCode(404)
		return
	}


	hasFromDate, fromDateValue, err := getIntFromQuery(ctx, "fromDate")
	if err != nil {
		ctx.SetStatusCode(400)
		return
	}

	hasToDate, toDateValue, err := getIntFromQuery(ctx, "toDate")
	if err != nil {
		ctx.SetStatusCode(400)
		return
	}

	hasFromAge, fromAgeValue, err := getIntFromQuery(ctx, "fromAge")
	if err != nil {
		ctx.SetStatusCode(400)
		return
	}

	hasToAge, toAgeValue, err := getIntFromQuery(ctx, "toAge")
	if err != nil {
		ctx.SetStatusCode(400)
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
	for _, vid := range visits_by_location[id] {
		v := visits[vid]
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
	WriteAvg(ctx, avg)
}

func RouterHandler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	pl := len(path) - 1
	method := ctx.Method()

	if method[0] == 'G' {
		if path[1] == 'u' {
			if path[pl] == 's' {
				id, _ := strconv.Atoi(path[7:pl-6])
				users_visits(ctx, id)
				return
			}
			id, _ := strconv.Atoi(path[7:])
			users_get(ctx, id)
			return
		}
		if path[1] == 'l' {
			if path[pl] == 'g' {
				id, _ := strconv.Atoi(path[11:pl-3])
				locations_avg(ctx, id)
				return
			}
			id, _ := strconv.Atoi(path[11:])
			locations_get(ctx, id)
			return
		}
		if path[1] == 'v' {
			id, _ := strconv.Atoi(path[8:])
			visits_get(ctx, id)
			return
		}
	} else {
		if path[1] == 'u' {
			if path[7] == 'n' {
				users_create(ctx)
				return
			}
			id, _ := strconv.Atoi(path[7:])
			users_update(ctx, id)
			return
		}
		if path[1] == 'l' {
			if path[11] == 'n' {
				locations_create(ctx)
				return
			}
			id, _ := strconv.Atoi(path[11:])
			locations_update(ctx, id)
			return
		}
		if path[1] == 'v' {
			if path[8] == 'n' {
				visits_create(ctx)
				return
			}
			id, _ := strconv.Atoi(path[8:])
			visits_update(ctx, id)
			return
		}
	}

	ctx.Error(fasthttp.StatusMessage(fasthttp.StatusNotFound),
		fasthttp.StatusNotFound)
}


func main() {
	loadData("/tmp/data/data.zip")
	fmt.Println("Good luck ^-^")

	server := fasthttp.Server{
		Handler: RouterHandler,
	}
	err := server.ListenAndServe(":80")
	if err != nil {
		log.Fatal(err)
	}
}
