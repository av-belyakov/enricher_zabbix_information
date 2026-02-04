package dnsresolver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/idna"

	"github.com/stretchr/testify/assert"
)

func TestGetIp(t *testing.T) {
	var (
		b          []byte
		err        error
		domanNames map[string]string = map[string]string{
			"https://epp.genproc.gov.ru:443/web/gprf":   "",
			"https://тверскаяобласть.рф/":               "",
			"https://оак-здоровье.рф/":                  "",
			"https://липецкаяобласть.рф/":               "",
			"http://президент.рф":                       "",
			"http://premier.gov.ru/":                    "",
			"https://www.yarregion.ru/default.aspx":     "",
			"https://git18.rostrud.gov.ru/":             "",
			"https://disk.roscosmos.ru/index.php/login": "",
		}
	)

	dnsResolver := func() *net.Resolver {
		return &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, address)
			},
		}
	}

	resolver := dnsResolver()

	b, err = os.ReadFile("../filesfortest/domainnames.json")
	if err != nil {
		log.Fatalln(err)
	}

	t.Run("Тест 1. Парсим доменные имена", func(t *testing.T) {
		//fmt.Println("")
		//fmt.Println("Parse request:")

		var num int
		for k := range domanNames {
			num++

			//fmt.Printf("%d. %s\n", num, k)

			urlHost, err := url.ParseRequestURI(k)
			assert.NoError(t, err)

			asciiDomain, err := idna.ToASCII(urlHost.Hostname())
			assert.NoError(t, err)

			//fmt.Printf("Result url parse:'%+v', host name:'%s'\n", urlHost, urlHost.Hostname())
			//fmt.Println("Internationalized Domain Names:", asciiDomain)

			domanNames[k] = asciiDomain
		}
	})

	t.Run("Тест 2. Делаем DNS запрос", func(t *testing.T) {
		fmt.Println("")
		fmt.Println("DNS lookup reguests:")

		var num int
		domainNameError := []string(nil)
		for k, v := range domanNames {
			num++

			fmt.Printf("%d. Origin host:'%s', parsed host:'%s'\n", num, k, v)

			addrs, err := resolver.LookupHost(t.Context(), v)
			if err != nil {
				domainNameError = append(domainNameError, v)
				assert.NoError(t, err)
			}

			fmt.Printf("addrs:'%v'\n", addrs)
		}

		//fmt.Printf("\nОтчёт:\nвсего домменных имён:'%d', преобразование с ошибкой:'%d'\n\n", len(domanNames), len(domainNameError))
		//for k, v := range domainNameError {
		//	fmt.Printf("%d. %s\n", k+1, v)
		//}
	})

	t.Run("Тест 3. Обработка доменных имён из файла", func(t *testing.T) {
		type ListDomainName struct {
			Hosts []struct {
				HostId string `json:"host_id"`
				Host   string `json:"host"`
				Name   string `json:"name"`
			} `json:"hosts"`
		}

		type Result []struct {
			Ips                []string
			Error              error
			OriginalDomainName string
			ParseDomainName    string
		}

		listDomainName := new(ListDomainName)
		result := Result{}

		err := json.Unmarshal(b, listDomainName)
		assert.NoError(t, err)

		for _, v := range listDomainName.Hosts {
			newStruct := struct {
				Ips                []string
				Error              error
				OriginalDomainName string
				ParseDomainName    string
			}{
				OriginalDomainName: v.Host,
			}

			if !strings.HasPrefix(v.Host, "http") {
				v.Host = "http://" + v.Host
			}

			urlHost, err := url.ParseRequestURI(v.Host)
			assert.NoError(t, err)

			asciiDomain, err := idna.ToASCII(urlHost.Hostname())
			assert.NoError(t, err)

			newStruct.ParseDomainName = asciiDomain

			addrs, err := resolver.LookupHost(t.Context(), asciiDomain)
			if err != nil {
				newStruct.Error = err
			}

			newStruct.Ips = addrs
			result = append(result, newStruct)
		}

		var domainNameWithError int
		fmt.Println("Отчёт Тест 3.")
		fmt.Printf("\n\tВсего домменных имён:'%d'\n\tСледующие доменные имена были преобразовани с ошибкой:\n", len(result))

		for _, v := range result {
			if v.Error == nil {
				continue
			}

			domainNameWithError++

			fmt.Printf("%d.\n\tOriginal:'%s'\n\tParse:'%s'\n\tError:'%s'\n", domainNameWithError, v.OriginalDomainName, v.ParseDomainName, v.Error.Error())
		}

		fmt.Printf("\n\tПреобразование с ошибкой:'%d'\n\n", domainNameWithError)

		assert.Equal(t, domainNameWithError, 1)
	})
}
