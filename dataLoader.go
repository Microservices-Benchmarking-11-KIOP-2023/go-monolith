package main

import (
	"encoding/json"
	"github.com/hailocab/go-geoindex"
	"github.com/harlow/go-micro-services/geo/data"
	"log"
)

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
