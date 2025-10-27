package zabbixinteraction

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

// NewZabbixConnectionJsonRPC создает объект соединения с Zabbix API
func NewZabbixConnectionJsonRPC(settings SettingsZabbixConnectionJsonRPC) (*ZabbixConnectionJsonRPC, error) {
	var zc *ZabbixConnectionJsonRPC

	connTimeout := 30 * time.Second
	if settings.ConnectionTimeout > (1 * time.Second) {
		connTimeout = settings.ConnectionTimeout
	}

	if settings.Host == "" {
		return zc, wrappers.WrapperError(errors.New("the value 'host' should not be empty"))
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     connTimeout,
			MaxIdleConnsPerHost: 10,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				RootCAs:            x509.NewCertPool(),
			},
		},
		Timeout: 15 * time.Second,
	}

	return &ZabbixConnectionJsonRPC{
		url:             fmt.Sprintf("https://%s/api_jsonrpc.php", settings.Host),
		host:            settings.Host,
		login:           settings.Login,
		passwd:          settings.Passwd,
		applicationType: "application/json-rpc",
		connClient:      client,
	}, nil
}

// Authorization запрос к Zabbix с целью получения хеша авторизации необходимого для
// дальнейшей работы с API
func (zc *ZabbixConnectionJsonRPC) Authorization(ctx context.Context) error {
	data := strings.NewReader(fmt.Sprintf(`{
	  "jsonrpc":"2.0",
	  "method":"user.login",
	  "params": {
	    "username":"%s",
		"password":"%s"
	  },
	  "id":1
	}`, zc.login, zc.passwd))

	result := ZabbixAuthorizationData{}
	res, err := zc.PostRequest(ctx, data)
	if err != nil {
		return wrappers.WrapperError(err)
	}

	if err := json.Unmarshal(res, &result); err != nil {
		return wrappers.WrapperError(err)
	}

	if len(result.Error) > 0 {
		var shortMsg, fullMsg string
		for k, v := range result.Error {
			if k == "message" {
				shortMsg = fmt.Sprint(v)
			}
			if k == "data" {
				fullMsg = fmt.Sprint(v)
			}
		}

		return wrappers.WrapperError(fmt.Errorf("error authorization, (%s %s)", shortMsg, fullMsg))
	}

	zc.authorizationHash = result.Result

	return nil
}

// GetAuthorizationData хеш авторизации
func (zc *ZabbixConnectionJsonRPC) GetAuthorizationData() string {
	return zc.authorizationHash
}

// PostRequest HTTP запрос типа POST
func (zc *ZabbixConnectionJsonRPC) PostRequest(ctx context.Context, data *strings.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", zc.url, data)
	if err != nil {
		return []byte{}, wrappers.WrapperError(err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", zc.authorizationHash))
	req.Header.Set("Content-Type", "application/json-rpc")

	res, err := zc.connClient.Do(req)
	if err != nil {
		return []byte{}, wrappers.WrapperError(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return []byte{}, wrappers.WrapperError(fmt.Errorf("error sending the request, response status is %s", res.Status))
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, wrappers.WrapperError(err)
	}

	return resBody, nil
}
