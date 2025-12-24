package netboxapi

import (
	"net/http"
)

// Settings настройки для подключения к Netbox
type Settings struct {
	Token string
	Host  string
	Port  int
}

// Client клиент для работы с Netbox
type Client struct {
	client   *http.Client
	Settings Settings
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
	Url          string `json:"url"`
	Prefix       string `json:"prefix"`
	Display      string `json:"display"`
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
