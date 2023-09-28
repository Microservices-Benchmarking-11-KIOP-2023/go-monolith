package main

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
