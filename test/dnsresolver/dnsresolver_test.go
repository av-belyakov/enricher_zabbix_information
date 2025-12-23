package dnsresolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/enricher_zabbix_information/internal/dnsresolver"
	"github.com/av-belyakov/enricher_zabbix_information/internal/storage"
	"github.com/av-belyakov/enricher_zabbix_information/test/helpers"
	"github.com/av-belyakov/enricher_zabbix_information/test/helpersfile"
)

func TestDnsResolver(t *testing.T) {
	b, err := os.ReadFile("../filesfortest/exampledata.json")
	if err != nil {
		log.Fatalln(err)
	}

	examleData := helpersfile.TypeExampleData{}
	if err := json.Unmarshal(b, &examleData); err != nil {
		log.Fatalln(err)
	}
	if len(examleData.Hosts) == 0 {
		log.Fatalln(errors.New("the structure 'TypeExampleData' should not be empty"))
	}

	sts := storage.NewShortTermStorage()

	//наполняем хранилище
	for _, v := range examleData.Hosts {
		hostId, err := strconv.Atoi(v.HostId)
		assert.NoError(t, err)

		sts.Add(storage.HostDetailedInformation{
			HostId:       hostId,
			OriginalHost: v.Host,
		})
	}

	listElement := sts.GetList()
	if len(listElement) == 0 {
		log.Fatalln(errors.New("the storage should not be empty"))
	}

	logging := helpers.NewLoggingForTest()

	ctx, ctxCancel := context.WithCancel(t.Context())
	go func() {
		for {
			select {
			case <-ctx.Done():
				ctxCancel()

				return

			case msg := <-logging.GetChan():
				fmt.Printf("Log message: type:'%s', message:'%s'\n", msg.GetType(), msg.GetMessage())
			}
		}
	}()

	dnsRes, err := dnsresolver.New(
		sts,
		logging,
		dnsresolver.WithTimeout(10),
	)
	if err != nil {
		log.Fatalln(err)
	}

	t.Run("Тест 1. Выполняем верификацию доменных имён.", func(t *testing.T) {
		//изменить статус выполнения процесса
		sts.SetProcessRunning()
		assert.True(t, sts.GetStatusProcessRunning())

		chOutput, err := dnsRes.Run(ctx)
		assert.NoError(t, err)

		for msg := range chOutput {
			fmt.Printf("%s, ips:%v (error: %v)\n", msg.DomainName, msg.Ips, msg.Error)
		}

		//изменить статус выполнения процесса
		sts.SetProcessNotRunning()
		assert.True(t, sts.GetStatusProcessRunning())

		errList := sts.GetListErrors()
		fmt.Println("\nCount element with errors:", len(errList))
		for k, v := range errList {
			fmt.Printf("%d.\n\tOriginalHost:'%s'\n\tList ip:'%v'\n\tError:'%s'\n", k+1, v.OriginalHost, v.Ips, v.Error.Error())
		}

		/*
			_, data, ok := sts.GetForHostId(11665)
			assert.True(t, ok)
			fmt.Printf("DATA:'%+v'\n", data)
		*/

		assert.Len(t, errList, 0)
	})

	t.Cleanup(func() {
		ctxCancel()
	})
}

/*
type LoggingForTest struct {
	chMessage chan interfaces.Messager
}

func NewLoggingForTest() *LoggingForTest {
	return &LoggingForTest{
		chMessage: make(chan interfaces.Messager),
	}
}

func (l *LoggingForTest) GetChan() <-chan interfaces.Messager {
	return l.chMessage
}

func (l *LoggingForTest) Send(msgType, msgData string) {
	msg := &MessageForTest{}
	msg.SetType(msgType)
	msg.SetMessage(msgData)

	l.chMessage <- msg
}

type MessageForTest struct {
	msgType, msgData string
}

func (m *MessageForTest) GetType() string {
	return m.msgType
}

func (m *MessageForTest) SetType(v string) {
	m.msgType = v
}

func (m *MessageForTest) GetMessage() string {
	return m.msgData
}

func (m *MessageForTest) SetMessage(v string) {
	m.msgData = v
}
*/
