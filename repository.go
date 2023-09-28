package main

import (
	"encoding/json"
	"github.com/hailocab/go-geoindex"
	"github.com/harlow/go-micro-services/app/data"
	"log"
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
