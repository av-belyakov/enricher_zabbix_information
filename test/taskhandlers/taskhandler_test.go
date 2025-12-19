package applicationtamplateexample_test

import (
	"flag"
	"testing"

	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	"github.com/joho/godotenv"
)

// при запуске теста можно указать флаг status=prod
// тогда тест будет запускатся с реальными данными авторизации
func TestTaskHandler(t *testing.T) {
	var appStatus string
	flag.StringVar(&appStatus, "status", "", "test")

	if appStatus == "prod" {
		if err := godotenv.Load("../../.env"); err != nil {
			t.Fatal(err)
		}
	}
	if err := godotenv.Load("../filesfortest/.env"); err != nil {
	}

	storageTemp := storage.NewShortTermStorage()
}
