package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hailocab/go-geoindex"
	"github.com/harlow/go-micro-services/geo/data"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	maxSearchRadius  = 10
	maxSearchResults = 1000000000
)

type Hotel struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	PhoneNumber string  `json:"phoneNumber"`
	Description string  `json:"description"`
	Address     Address `json:"address"`
	Images      []Image `json:"images"`
}

type Address struct {
	StreetNumber string  `json:"streetNumber"`
	StreetName   string  `json:"streetName"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	Country      string  `json:"country"`
	PostalCode   string  `json:"postalCode"`
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
}

type Image struct {
	URL     string `json:"url"`
	Default bool   `json:"default"`
}

var (
	rateTable map[stay]*RatePlan
	geoIndex  *geoindex.ClusteringIndex
	profiles  map[string]*Hotel
)

func main() {
	var port = flag.Int("port", 8080, "The service port")
	flag.Parse()

	loadAllData()

	http.HandleFunc("/hotels", hotelsHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func hotelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	latParam, lonParam := r.URL.Query().Get("lat"), r.URL.Query().Get("lon")
	if latParam == "" || lonParam == "" {
		http.Error(w, "Please specify lat/lon params", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(latParam), 64)
	if err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(strings.TrimSpace(lonParam), 64)
	if err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	points := getNearbyPoints(lat, lon)

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

	var hotels []*Hotel
	for _, rate := range ratePlans {
		if hotel, ok := profiles[rate.HotelId]; ok {
			hotels = append(hotels, hotel)
		}
	}

	err = json.NewEncoder(w).Encode(geoJSONResponse(hotels))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func loadAllData() {
	geoIndex = newGeoIndex()
	rateTable = loadRateTable()
	profiles = loadProfiles()
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

type RatePlan struct {
	HotelId         string   `json:"hotelId"`
	Code            string   `json:"code"`
	InDate          string   `json:"inDate"`
	OutDate         string   `json:"outDate"`
	RoomTypeDetails RoomType `json:"roomType"`
}

type RoomType struct {
	BookableRate       float64 `json:"bookableRate"`
	TotalRate          float64 `json:"totalRate"`
	TotalRateInclusive float64 `json:"totalRateInclusive"`
	Code               string  `json:"code"`
	Currency           string  `json:"currency"`
	RoomDescription    string  `json:"roomDescription"`
}

type stay struct {
	HotelID string
	InDate  string
	OutDate string
}

type point struct {
	Pid  string  `json:"hotelId"`
	Plat float64 `json:"lat"`
	Plon float64 `json:"lon"`
}

func (p *point) Lat() float64 { return p.Plat }
func (p *point) Lon() float64 { return p.Plon }
func (p *point) Id() string   { return p.Pid }

func newGeoIndex() *geoindex.ClusteringIndex {
	var (
		file   = data.MustAsset("data/geo.json")
		points []*point
	)

	if err := json.Unmarshal(file, &points); err != nil {
		log.Fatalf("Failed to load hotels: %v", err)
	}

	index := geoindex.NewClusteringIndex()
	for _, point := range points {
		index.Add(point)
	}

	return index
}

func loadProfiles() map[string]*Hotel {
	var (
		file   = data.MustAsset("data/hotels.json")
		hotels []*Hotel
	)

	if err := json.Unmarshal(file, &hotels); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	profiles := make(map[string]*Hotel)
	for _, hotel := range hotels {
		profiles[hotel.Id] = hotel
	}
	return profiles
}

func loadRateTable() map[stay]*RatePlan {
	file := data.MustAsset("data/inventory.json")

	var rates []*RatePlan
	if err := json.Unmarshal(file, &rates); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	rateTable := make(map[stay]*RatePlan)
	for _, ratePlan := range rates {
		stay := stay{
			HotelID: ratePlan.HotelId,
			InDate:  ratePlan.InDate,
			OutDate: ratePlan.OutDate,
		}
		rateTable[stay] = ratePlan
	}

	return rateTable
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
