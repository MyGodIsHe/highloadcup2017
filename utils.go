package main

import (
	"time"
	"log"
	"fmt"
	"strings"
	"encoding/json"
	"archive/zip"
	"github.com/valyala/fasthttp"
	"github.com/buger/jsonparser"
	"sort"
)

func diff(a, b time.Time) int {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year := int(y2 - y1)
	month := int(M2 - M1)
	day := int(d2 - d1)
	hour := int(h2 - h1)
	min := int(m2 - m1)
	sec := int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return year
}

func parseId(body []byte, value *int, required bool) bool {
	if !required {
		return true
	}
	v , err := jsonparser.GetInt(body, "id")
	if err != nil {
		return false
	}
	*value = int(v)
	return true
}

func parseString(body []byte, value *string, name string, required bool) bool {
	v , err := jsonparser.GetString(body, name)
	if err != nil {
		return !required
	}
	*value = v
	return true
}

func parseInt(body []byte, value *int, name string, required bool) bool {
	v , err := jsonparser.GetInt(body, name)
	if err != nil {
		return !required
	}
	*value = int(v)
	return true
}


func getIntFromQuery(ctx *fasthttp.RequestCtx, sv string) (bool, int, interface{}) {
	args := ctx.URI().QueryArgs()
	if args.Has(sv) {
		v, err := args.GetUint(sv)
		return true, v, err
	}
	return false, 0, nil
}


func OrderedInsert(a []int, j int) []int {
	n := len(a)
	if n == 0 {
		return append(a, j)
	}

	i := sort.Search(n, func(i int) bool { return a[i] >= j})
	return append(a[:i], append([]int{j}, a[i:]...)...)
}

func OrderedSearch(a []int, j int) (int, bool) {
	n := len(a)
	i := sort.Search(n, func(i int) bool { return a[i] == j})
	if i == n {  // not found
		return 0, false
	}
	return i, true
}

func OrderedHas(a []int, j int) bool {
	n := len(a)
	i := sort.Search(n, func(i int) bool { return a[i] == j})
	if i == n {  // not found
		return false
	}
	return true
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
		defer rc.Close()
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
			}
		}
		if strings.HasPrefix(f.Name, "visits") {
			var recs DataVisit
			err = json.NewDecoder(rc).Decode(&recs)
			if err != nil {
				log.Fatal(err)
			}
			for _, rec := range recs.Visits {
				visitSetEvent(rec)
			}
		}
		fmt.Println("done")
	}
	fmt.Println("Users: ", len(users))
	fmt.Println("Locations: ", len(locations))
	fmt.Println("Visits: ", len(visits))
}
