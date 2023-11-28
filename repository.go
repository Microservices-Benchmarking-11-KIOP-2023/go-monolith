package main

import (
	"encoding/json"
	"github.com/harlow/go-micro-services/app/data"
	"log"
	"math"
)

const (
	maxSearchRadius = 10
	earthRadiusKm   = 6371
)

var (
	rateTable map[stay]*RatePlan
	points    []*point
	profiles  map[string]*Hotel
)

func loadAllData() {
	points = loadPoints("data/geo.json")
	rateTable = loadRateTable()
	profiles = loadProfiles()
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad, lon1Rad := lat1*math.Pi/180, lon1*math.Pi/180
	lat2Rad, lon2Rad := lat2*math.Pi/180, lon2*math.Pi/180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func getNearbyPoints(lat, lon float64) []string {
	var nearbyHotels []string
	for _, point := range points {
		distance := haversineDistance(lat, lon, point.Plat, point.Plon)
		if distance <= maxSearchRadius {
			nearbyHotels = append(nearbyHotels, point.Pid)
		}
	}

	return nearbyHotels
}

func loadPoints(path string) []*point {
	var (
		file   = data.MustAsset(path)
		points []*point
	)

	if err := json.Unmarshal(file, &points); err != nil {
		log.Fatalf("Failed to load hotels: %v", err)
	}

	return points
}

func getRatePlans(hotelIds []string, inDate string, outDate string) []*RatePlan {
	var ratePlans []*RatePlan

	for _, id := range hotelIds {
		s := stay{
			HotelID: id,
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
