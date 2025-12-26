package netboxhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/enricher_zabbix_information/internal/netboxapi"
)

func TestGetIPAddresses(t *testing.T) {
	var (
		host    string = "netbox.cloud.gcm"
		port    int    = 8005
		request string = "/api/ipam/ip-addresses/?limit=30&offset=302"
	)

	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(err)
	}

	client := http.Client{
		Timeout: 15 * time.Second,
	}

	url := fmt.Sprintf("http://%s:%d%s", host, port, request)

	req, err := http.NewRequestWithContext(t.Context(), "GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Token:", os.Getenv("GO_ENRICHERZI_NBTOKEN"))

	req.Header.Add("Authorization", "Token "+os.Getenv("GO_ENRICHERZI_NBTOKEN"))
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	resBody, err := io.ReadAll(res.Body)
	assert.NoError(t, err)

	fmt.Printf("RESPONSE:\n'%s'\n", string(resBody))

	listPrefixes := netboxapi.ListPrefixes{}
	err = json.Unmarshal(resBody, &listPrefixes)
	assert.NoError(t, err)

	fmt.Println("Response count:", listPrefixes.Count)
	for k, v := range listPrefixes.Results {
		fmt.Printf("\t%d.\n%+v\n", k+1, v)
	}

	t.Cleanup(func() {
		res.Body.Close()

		os.Unsetenv("GO_ENRICHERZI_MAIN")

		os.Unsetenv("GO_ENRICHERZI_ZPASSWD")
		os.Unsetenv("GO_ENRICHERZI_NBTOKEN")
		os.Unsetenv("GO_ENRICHERZI_DBWLOGPASSWD")
		os.Unsetenv("GO_ENRICHERZI_APISERVERTOKEN")
	})
}
