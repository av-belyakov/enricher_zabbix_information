package storage

import (
	"fmt"
	"log"
	"testing"

	"github.com/av-belyakov/enricher_zabbix_information/internal/appstorage"
	"github.com/stretchr/testify/assert"
)

func TestTagComparison(t *testing.T) {
	var (
		hostId int = 1122
		as     *appstorage.SharedAppStorage
		err    error
	)

	someTags := []appstorage.Tag{
		{
			Tag:   "COA",
			Value: "45632, 665545, 4763, 18996",
		},
		{
			Tag:   "HomeNet: yes",
			Value: "",
		},
		{
			Tag:   "IP",
			Value: "96.123.36.36, 78.163.155.65",
		},
	}

	as, err = appstorage.New()
	if err != nil {
		log.Fatalln(err)
	}

	as.AddElement(appstorage.HostDetailedInformation{
		HostId:       hostId,
		OriginalHost: "any.example.name",
		Tags:         someTags,
	})

	t.Run("Тест 1. Теги должны совпадать", func(t *testing.T) {
		isComparison, err := as.IsTagComparison(hostId, someTags)
		assert.NoError(t, err)
		assert.True(t, isComparison)
	})

	t.Run("Тест 2.1. Теги НЕ должны совпадать. Добавил новый тег.", func(t *testing.T) {
		newTags := make([]appstorage.Tag, len(someTags))
		copy(newTags, someTags)

		newTags = append(newTags, appstorage.Tag{
			Tag:   "Any_Tag",
			Value: "Any_Value",
		})

		fmt.Println("2.1 newTags:", newTags)

		isComparison, err := as.IsTagComparison(hostId, newTags)
		assert.NoError(t, err)
		assert.False(t, isComparison)
	})

	t.Run("Тест 2.2. Теги НЕ должны совпадать. Изменил существующий тег.", func(t *testing.T) {
		newTags := make([]appstorage.Tag, len(someTags))
		copy(newTags, someTags)

		newTags[1].Tag = "Any_Tag"
		newTags[1].Value = "Any_Value"

		fmt.Println("2.2-1 newTags:", newTags)

		isComparison, err := as.IsTagComparison(hostId, newTags)
		assert.NoError(t, err)
		assert.False(t, isComparison)

		newTags = make([]appstorage.Tag, len(someTags))
		copy(newTags, someTags)

		newTags[0].Value = "11111, 22222, 33333"

		fmt.Println("2.2-2 newTags:", newTags)

		isComparison, err = as.IsTagComparison(hostId, newTags)
		assert.NoError(t, err)
		assert.False(t, isComparison)
	})
}
