package netboxhandler

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/enricher_zabbix_information/internal/netboxapi"
	"github.com/av-belyakov/enricher_zabbix_information/internal/taskhandlers"
	"github.com/av-belyakov/enricher_zabbix_information/test/helpers"
)

const (
	host = "netbox.cloud.gcm"
	port = 8005
)

func TestGetPrefixes(t *testing.T) {
	//var (
	//	sizePrefixes int
	//)

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(err)
	}

	client, err := netboxapi.New(
		os.Getenv("GO_ENRICHERZI_NBTOKEN"),
		netboxapi.WithHost(host),
		netboxapi.WithPort(port),
		netboxapi.WithTimeout(10),
	)
	if err != nil {
		log.Fatal(err)
	}

	logging := helpers.NewLoggingForTest()
	ctx, ctxCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case msg := <-logging.GetChan():
				fmt.Printf("Log message: type:'%s', message:'%s'\n", msg.GetType(), msg.GetMessage())
			}
		}
	}()

	t.Run("Тест 1. Получить префиксы из Netbox", func(t *testing.T) {
		shortPrefixList, count, err := taskhandlers.NetboxPrefixes(ctx, client, logging)
		assert.NoError(t, err)
		assert.Greater(t, count, 0)

		var i int
		fmt.Println("Read first 30 elements:")
		for msg := range shortPrefixList {
			i++
			fmt.Println("a piece was accepted:", i)

			if i < 2 {
				for j, v := range msg {
					fmt.Println(
						"Index:", j+1,
						"Prefix:", v.Prefix,
						"Status:", v.Status,
						"Id:", v.Id,
						"SensorId:", v.SensorId,
					)
				}
			}
		}
	})

	/*t.Run("Тест 1. Получить общее количество префиксов", func(t *testing.T) {
		res, statusCode, err := client.Get(t.Context(), "/api/ipam/prefixes/?limit=1")
		assert.NoError(t, err)
		assert.Equal(t, statusCode, http.StatusOK)

		nbPrefixes := netboxapi.ListPrefixes{}
		err = json.Unmarshal(res, &nbPrefixes)
		assert.NoError(t, err)

		sizePrefixes = nbPrefixes.Count
	})

	t.Run("Тест 2. Получить полный список префиксов имеющихся в NetBox", func(t *testing.T) {
		chunkCount := math.Ceil(float64(sizePrefixes) / float64(chunkSize))

		pl := PrefixList{
			Count: sizePrefixes,
		}

		for i := 0; i < int(chunkCount); i++ {
			offset := i * chunkSize
			res, statusCode, err := client.Get(t.Context(), fmt.Sprintf("/api/ipam/prefixes/?limit=%d&offset=%d", chunkSize, offset))
			assert.NoError(t, err)
			assert.Equal(t, statusCode, http.StatusOK)

			nbPrefixes := netboxapi.ListPrefixes{}
			err = json.Unmarshal(res, &nbPrefixes)
			assert.NoError(t, err)

			pl.Prefixes = append(pl.Prefixes, nbPrefixes.Results...)
		}

		assert.NotEmpty(t, pl.Prefixes)
		assert.Equal(t, pl.Count, len(pl.Prefixes))

		fmt.Printf("GET PREFIXES RESULT:\nCount:'%d',\nSize Prefixes:'%d'\n", pl.Count, len(pl.Prefixes))

		//заполняем структуру с краткой информацией о префиксах
		shortPrefixList := netboxapi.ShortPrefixList{}
		for _, prefix := range pl.Prefixes {
			netPrefix, err := netip.ParsePrefix(prefix.Prefix)
			if err != nil {
				fmt.Println("Error:", err)

				continue
			}

			var sensorId string
			if len(prefix.CustomFields.Sensors) > 0 {
				sensorId = prefix.CustomFields.Sensors[0].Name
			}

			shortPrefixList.Prefixes = append(
				shortPrefixList.Prefixes,
				netboxapi.ShortPrefixInfo{
					Status:   prefix.Status.Value,
					Prefix:   netPrefix,
					Id:       prefix.Id,
					SensorId: sensorId,
				})
		}

		fmt.Println("Size shortPrefixList:", len(shortPrefixList.Prefixes))
		fmt.Println("Read first 30 elements:")
		for i := range 30 {
			fmt.Println(
				"Index:", i,
				"Prefix:", shortPrefixList.Prefixes[i].Prefix,
				"Status:", shortPrefixList.Prefixes[i].Status,
				"Id:", shortPrefixList.Prefixes[i].Id,
				"SensorId:", shortPrefixList.Prefixes[i].SensorId,
			)
		}


		//	prefix, err := netip.ParsePrefix()
		//	ok := prefix.Contains()


		//положить в файл (опционально через -- -tofile)
		var flagIsExist bool
		for _, arg := range os.Args {
			if arg == "-tofile" {
				flagIsExist = true
			}
		}

		if !flagIsExist {
			return
		}

		// пишем в файл //

		_, err := os.Stat(fileName)
		if !os.IsNotExist(err) {
			t.Log("удаляем файл", fileName)

			assert.NoError(t, os.RemoveAll(fileName))
		}

		b, err := json.Marshal(pl)
		assert.NoError(t, err)

		f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)
		assert.NoError(t, err)

		_, err = f.WriteString(string(b))
		assert.NoError(t, err)

		f.Close()
	})*/

	//t.Run("", func(t *testing.T) {})

	t.Cleanup(func() {
		ctxCancel()

		os.Unsetenv("GO_ENRICHERZI_MAIN")

		os.Unsetenv("GO_ENRICHERZI_ZPASSWD")
		os.Unsetenv("GO_ENRICHERZI_NBTOKEN")
		os.Unsetenv("GO_ENRICHERZI_DBWLOGPASSWD")
		os.Unsetenv("GO_ENRICHERZI_APISERVERTOKEN")
	})
}

type PrefixList struct {
	Prefixes []netboxapi.Prefixes
	Count    int
}
