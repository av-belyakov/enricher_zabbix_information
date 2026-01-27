package netboxapi

import (
	"net/http"
	"net/netip"
)

// Settings настройки для подключения к Netbox
type Settings struct {
	token string
	host  string
	port  int
}

// Client клиент для работы с Netbox
type Client struct {
	client   *http.Client
	settings Settings
}

type ShortPrefixList []ShortPrefixInfo

//type ShortPrefixList struct {
//	mutex    sync.RWMutex
//	Prefixes []ShortPrefixInfo
//	Count    int
//}

type ShortPrefixInfo struct {
	Prefix   netip.Prefix
	Status   string
	SensorId string
	Id       int
}

type ListPrefixes struct {
	Results  []Prefixes `json:"results"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Count    int        `json:"count"`
}

// Prefixes структура для хранения информации о префиксах
// здесь не все поля, а только те, которые нужны
type Prefixes struct {
	Status struct {
		Value string `json:"value"`
	} `json:"status"`
	Tags []struct {
		Id      int    `json:"id"`
		Url     string `json:"url"`
		Display string `json:"display"`
		Name    string `json:"name"`
		Slug    string `json:"slug"`
		Color   string `json:"color"`
	} `json:"tags"`
	Url     string `json:"url"`
	Address string `json:"address"`
	Prefix  string `json:"prefix"`
	Display string `json:"display"`
	Vrf     struct {
		URL         string `json:"url"`
		Display     string `json:"display"`
		Name        string `json:"name"`
		Rd          string `json:"rd"`
		Description string `json:"description"`
		Id          int    `json:"id"`
	} `json:"vrf"`
	Tenant struct {
		URL         string `json:"url"`
		Display     string `json:"display"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		Id          int    `json:"id"`
	} `json:"tenant"`
	Family struct {
		Label string `json:"label"`
		Value int    `json:"value"`
	} `json:"family"`
	CustomFields struct {
		Sensors []struct {
			Url         string `json:"url"`
			Name        string `json:"name"`
			Display     string `json:"display"`
			Description string `json:"description"`
			Id          int    `json:"id"`
		} `json:"sensors"`
	} `json:"custom_fields"`
	Id int `json:"id"`
}
