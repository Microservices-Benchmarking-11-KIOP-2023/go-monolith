package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/hailocab/go-geoindex"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	maxSearchRadius  = 10
	maxSearchResults = 1000000000
)

var (
	rateTable map[stay]*RatePlan
	geoIndex  *geoindex.ClusteringIndex
	profiles  map[string]*Hotel
)

func loadAllData() {
	geoIndex = newGeoIndex()
	rateTable = loadRateTable()
	profiles = loadProfiles()
}

func geoJSONResponse(hotels []*Hotel) map[string]interface{} {
	var fs []interface{}

	for _, h := range hotels {
		fs = append(fs, map[string]interface{}{
			"type": "Feature",
			"id":   h.Id,
			"properties": map[string]string{
				"name":         h.Name,
				"phone_number": h.PhoneNumber,
			},
			"geometry": map[string]interface{}{
				"type": "Point",
				"coordinates": []float64{
					h.Address.Lon,
					h.Address.Lat,
				},
			},
		})
	}

	return map[string]interface{}{
		"type":     "FeatureCollection",
		"features": fs,
	}
}

func main() {
	var port = flag.Int("port", 8080, "The service port")
	flag.Parse()

	loadAllData()

	http.HandleFunc("/hotels", hotelsHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func hotelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	inDate, outDate, lon, lat, err := getParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	points := getNearbyPoints(lat, lon)
	ratePlans := getRatePlans(points, inDate, outDate)
	hotels := getHotels(ratePlans)

	err = json.NewEncoder(w).Encode(geoJSONResponse(hotels))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getParams(r *http.Request) (string, string, float64, float64, error) {
	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		return "_", "_", 0, 0, errors.New("inDate/outDate params not specified")
	}

	latParam, lonParam := r.URL.Query().Get("lat"), r.URL.Query().Get("lon")
	if latParam == "" || lonParam == "" {
		return "_", "_", 0, 0, errors.New("lon/lat params not specified")
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(latParam), 64)
	if err != nil {
		return "_", "_", 0, 0, errors.New("invalid latitude")
	}

	lon, err := strconv.ParseFloat(strings.TrimSpace(lonParam), 64)
	if err != nil {
		return "_", "_", 0, 0, errors.New("invalid longitude")
	}

	return inDate, outDate, lon, lat, nil
}

func getNearbyPoints(lat, lon float64) []geoindex.Point {
	center := &geoindex.GeoPoint{
		Pid:  "",
		Plat: lat,
		Plon: lon,
	}

	return geoIndex.KNearest(
		center,
		maxSearchResults,
		geoindex.Km(maxSearchRadius),
		func(p geoindex.Point) bool {
			return true
		},
	)
}

func getRatePlans(points []geoindex.Point, inDate string, outDate string) []*RatePlan {
	var ratePlans []*RatePlan

	for _, p := range points {
		s := stay{
			HotelID: p.Id(),
			InDate:  inDate,
			OutDate: outDate,
		}
		if rate, ok := rateTable[s]; ok {
			ratePlans = append(ratePlans, rate)
		}
	}

	return ratePlans
}

func getHotels(ratePlans []*RatePlan) []*Hotel {
	var hotels []*Hotel
	for _, rate := range ratePlans {
		if hotel, ok := profiles[rate.HotelId]; ok {
			hotels = append(hotels, hotel)
		}
	}
	return hotels
}
