package shortlogstory_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/enricher_zabbix_information/internal/shortlogstory"
)

func TestStorage(t *testing.T) {
	storageLog := shortlogstory.NewShortLogStory(10)
	for num := range 11 {
		storageLog.Add(shortlogstory.LogInformation{
			Type:        fmt.Sprintf("info:%d", num),
			Description: fmt.Sprintf("description for log №%d", num),
		})
	}

	list := storageLog.Get()
	assert.Equal(t, len(list), 10)
	assert.Equal(t, list[0].Type, "info:1")

	//fmt.Println("1 LIST:", list)

	storageLog.Add(shortlogstory.LogInformation{
		Type:        "info:12",
		Description: "description for log №11",
	})
	list = storageLog.Get()
	assert.Equal(t, list[0].Type, "info:2")

	//fmt.Println("2 LIST:", list)
}
