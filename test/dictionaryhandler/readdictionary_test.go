package dictionaryhandler

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	dictionaries "github.com/av-belyakov/enricher_zabbix_information/internal/dictionarieshandler"
)

func TestReadDictionary(t *testing.T) {
	dict, err := dictionaries.Read("config/dictionary.yml")
	assert.NoError(t, err)

	fmt.Printf("Result: '%+v'\n", dict)
	assert.Greater(t, len(dict.Dictionaries.WebSiteGroupMonitoring), 0)
}
