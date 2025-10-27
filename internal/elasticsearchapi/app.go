package elasticsearchapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

// NewElasticsearchConnect конструктор соединения с БД Elasticsearch
func NewElasticsearchConnect(settings Settings) (*ElasticsearchDB, error) {
	edb := &ElasticsearchDB{settings: settings}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%d", settings.Host, settings.Port)},
		Username:  settings.User,
		Password:  settings.Passwd,
		Transport: &http.Transport{
			MaxIdleConns:          10,              //число открытых TCP-соединений, которые в данный момент не используются
			IdleConnTimeout:       1 * time.Second, //время, через которое закрываются такие неактивные соединения
			MaxIdleConnsPerHost:   10,              //число неактивных TCP-соединений, которые допускается устанавливать на один хост
			ResponseHeaderTimeout: 2 * time.Second, //время в течении которого сервер ожидает получение ответа после записи заголовка запроса
			DialContext: (&net.Dialer{
				Timeout: 3 * time.Second,
				//KeepAlive: 1 * time.Second,
			}).DialContext,
		},
		//RetryOnError: ,
		//RetryOnStatus: ,
	})
	if err != nil {
		return edb, err
	}

	edb.client = es

	return edb, err
}

// Write запись в БД
func (edb *ElasticsearchDB) Write(msgType, msg string) error {
	if edb.client == nil {
		return errors.New("the client parameters for connecting to the Elasticsearch database are not set correctly")
	}

	msg = supportingfunctions.ReplaceCommaCharacter(msg)

	tn := time.Now()
	buf := bytes.NewReader(fmt.Appendf(nil, `{
		  "datetime": "%s",
		  "type": "%s",
		  "nameRegionalObject": "%s",
		  "message": "%s"
		}`,
		tn.Format(time.RFC3339),
		msgType,
		edb.settings.NameRegionalObject,
		msg,
	))

	res, err := edb.client.Index(fmt.Sprintf("logs.%s_%s_%d", edb.settings.IndexDB, strings.ToLower(tn.Month().String()), tn.Year()), buf)
	defer responseClose(res)
	if err != nil {
		return wrappers.WrapperError(err)
	}

	if res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusOK {
		return nil
	}

	r := map[string]any{}
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		return wrappers.WrapperError(err)
	}

	if e, ok := r["error"]; ok {
		return wrappers.WrapperError(fmt.Errorf("%s received from module Elasticsearch: %s", res.Status(), e))
	}

	return nil
}

func responseClose(res *esapi.Response) {
	if res == nil || res.Body == nil {
		return
	}

	res.Body.Close()
}
