package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/enricher_zabbix_information/internal/appstorage"
)

func TestAppStorageLogs(t *testing.T) {
	storageLog, err := appstorage.New(appstorage.WithSizeLogs(10))
	if err != nil {
		t.Fatal(err)
	}

	for num := range 11 {
		storageLog.AddLog(appstorage.LogInformation{
			Type:        fmt.Sprintf("info:%d", num),
			Description: fmt.Sprintf("description for log №%d", num),
		})
	}

	list := storageLog.GetLog()
	assert.Equal(t, len(list), 10)
	assert.Equal(t, list[0].Type, "info:10") // так как есть slice reverce

	//fmt.Println("1 LIST:", list)

	storageLog.AddLog(appstorage.LogInformation{
		Type:        "info:11",
		Description: "description for log №11",
	})
	list = storageLog.GetLog()
	assert.Equal(t, list[0].Type, "info:11") // так как есть slice reverce

	//fmt.Println("2 LIST:", list)
}
