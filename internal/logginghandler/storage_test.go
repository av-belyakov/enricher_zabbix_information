package logginghandler_test

import (
	"fmt"
	"testing"

	"github.com/av-belyakov/enricher_zabbix_information/internal/logginghandler"
	"github.com/stretchr/testify/assert"
)

func TestStorage(t *testing.T) {
	storageLog := logginghandler.NewShortLogStory(10)
	for num := range 11 {
		storageLog.Add(logginghandler.LogInformation{
			Type:        fmt.Sprintf("info:%d", num),
			Description: fmt.Sprintf("description for log №%d", num),
		})
	}

	list := storageLog.Get()
	assert.Equal(t, len(list), 10)
	assert.Equal(t, list[0].Type, "info:1")

	//fmt.Println("1 LIST:", list)

	storageLog.Add(logginghandler.LogInformation{
		Type:        "info:12",
		Description: "description for log №11",
	})
	list = storageLog.Get()
	assert.Equal(t, list[0].Type, "info:2")

	//fmt.Println("2 LIST:", list)
}
