package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type countries struct {
	Code        string    `json:"alpha2Code"`
	Countryname string    `json:"name"`
	Countryflag string    `json:"flag"`
	Results     []results `json:"results"`
}

type results struct {
	Species    string `json:"species"`
	SpeciesKey int    `json:"speciesKey"`
}

type speciesDetail struct {
	Key            int    `json:"speciesKey"`
	Kingdom        string `json:"kingdom"`
	Phylum         string `json:"phylum"`
	Order          string `json:"order"`
	Family         string `json:"family"`
	Genus          string `json:"genus"`
	ScientificName string `json:"scientificName"`
	CanonicalName  string `json:"canonicalName"`
	Year           string `json:"bracketYear"`
}

type diagnostics struct {
	Gbif          string        `json:"gbif"`
	Restcountries string        `json:"restcountries"`
	Version       string        `json:"version"`
	Uptime        time.Duration `json:"uptime"`
}

var version = "1"
var startTime time.Time

func uptime() time.Duration {
	return time.Since(startTime)
}

func init() {
	startTime = time.Now()
}

func countryHandler(w http.ResponseWriter, r *http.Request) {
	var country countries
	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)

	var req = strings.Split(r.URL.RequestURI(), "/")
	var temp = strings.Split(req[2], "?")
	var temp2 = strings.Split(temp[1], "=")
	var code = temp[0]
	var limit = temp2[1]

	resp, err := http.Get("https://restcountries.eu/rest/v2/alpha/" + code)

	if err != nil {
		panic(err)
	}

	if resp != nil {
		err := json.NewDecoder(resp.Body).Decode(&country)
		if err != nil {
			panic(err)
		}
	}

	resp, err = http.Get("http://api.gbif.org/v1/occurrence/search?country=" + code + "&limit=" + limit)

	if err != nil {
		panic(err)
	}

	if resp != nil {
		err := json.NewDecoder(resp.Body).Decode(&country)
		if err != nil {
			panic(err)
		}
	}

	var limitNr, er = strconv.Atoi(limit)
	if er != nil {
		panic(er)
	}

	for i := 0; i < limitNr; i++ {
		var nameCheck = country.Results[i].Species
		for j := i + 1; j < limitNr; j++ {
			if country.Results[j].Species == nameCheck {
				country.Results[j].Species = ""
				country.Results[j].SpeciesKey = 0
			}
		}

	}

	json.NewEncoder(w).Encode(country)
}

func speciesHandler(w http.ResponseWriter, r *http.Request) {
	var spDetail speciesDetail
	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)

	var req = strings.Split(r.URL.RequestURI(), "/")
	var key = req[2]

	resp, err := http.Get("http://api.gbif.org/v1/species/" + key)

	if err != nil {
		panic(err)
	}

	if resp != nil {
		err := json.NewDecoder(resp.Body).Decode(&spDetail)
		if err != nil {
			panic(err)
		}
	}

	resp, err = http.Get("http://api.gbif.org/v1/species/" + key + "/name")

	if err != nil {
		panic(err)
	}

	if resp != nil {
		err := json.NewDecoder(resp.Body).Decode(&spDetail)
		if err != nil {
			panic(err)
		}
	}

	json.NewEncoder(w).Encode(spDetail)
}

func diagnosticsHandler(w http.ResponseWriter, r *http.Request) {
	var dia diagnostics
	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)

	resp, err := http.Get("http://api.gbif.org/v1/")

	if err != nil {
		panic(err)
	}

	if resp != nil {
		dia.Gbif = "Found gbif"
	} else {
		dia.Gbif = "N/A"
	}

	resp, err = http.Get("https://restcountries.eu/rest/v2/")

	if err != nil {
		panic(err)
	}

	if resp != nil {
		dia.Restcountries = "Found restcountries"
	} else {
		dia.Restcountries = "N/A"
	}

	dia.Version = version

	dia.Uptime = uptime() / 1000000000

	json.NewEncoder(w).Encode(dia)
}

func main() {
	port := os.Getenv("PORT")

	http.HandleFunc("/diag/", diagnosticsHandler)
	http.HandleFunc("/species/", speciesHandler)
	http.HandleFunc("/countries/", countryHandler)

	http.ListenAndServe(":"+port, nil)
}
