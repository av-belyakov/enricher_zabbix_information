package confighandler_test

import (
	"fmt"
	"log"
	"os"
	"slices"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
)

var (
	conf *confighandler.ConfigApp

	err error
)

func TestMain(m *testing.M) {
	unSetEnviroment()

	//загружаем ключи и пароли
	if err := godotenv.Load("../filesfortest/.env"); err != nil {
		log.Fatalln(err)
	}

	os.Setenv("GO_"+constants.App_Environment_Name+"_MAIN", "test")

	conf, err = confighandler.New(constants.Root_Dir)
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(m.Run())
}

func TestReadConfigHandler(t *testing.T) {
	t.Run("Тест чтения конфигурационного файла", func(t *testing.T) {
		t.Run("Тест 1. Проверка аутентификационных данных", func(t *testing.T) {
			assert.NotEmpty(t, conf.GetAuthenticationData().NetBoxPasswd)
			assert.NotEmpty(t, conf.GetAuthenticationData().ZabbixPasswd)
			assert.NotEmpty(t, conf.GetAuthenticationData().APIServerToken)
			assert.NotEmpty(t, conf.GetAuthenticationData().WriteLogBDPasswd)
		})

		t.Run("Тест 2. Проверка настроек расписания работы сервиса", func(t *testing.T) {
			assert.Equal(t, conf.GetSchedule().TimerJob, 1)
			assert.Equal(t, len(conf.GetSchedule().DailyJob), 6)

			listTime := []string{
				"00:00:00",
				"06:45:00",
				"12:01:23",
				"19:59:03",
				"21:13:13",
				"22:47:03",
			}
			for _, v := range conf.GetSchedule().DailyJob {
				assert.True(t, slices.Contains(listTime, v))
			}
		})

		t.Run("Тест 3. Проверка настройки NetBox", func(t *testing.T) {
			assert.Equal(t, conf.GetNetBox().Host, "localhost")
			assert.Equal(t, conf.GetNetBox().Port, 4455)
			assert.Equal(t, conf.GetNetBox().User, "someuser")
		})

		t.Run("Тест 4. Проверка настройки Zabbix", func(t *testing.T) {
			assert.Equal(t, conf.GetZabbix().Host, "192.168.9.45")
			assert.Equal(t, conf.GetZabbix().Port, 443)
			assert.Equal(t, conf.GetZabbix().Timeout, 30)
			assert.Equal(t, conf.GetZabbix().User, "803.p.vishnitsky@avz-center.ru")
		})

		t.Run("Тест 5. Проверка настройки API Information Server", func(t *testing.T) {
			assert.Equal(t, conf.GetInformationServerApi().Host, "localhost")
			assert.Equal(t, conf.GetInformationServerApi().Port, 8989)
		})

		t.Run("Тест 6. Проверка настройки WriteLogDataBase", func(t *testing.T) {
			assert.Equal(t, conf.GetLogDB().Host, "database.cloud.example")
			assert.Equal(t, conf.GetLogDB().Port, 9200)
			assert.Equal(t, conf.GetLogDB().User, "log_writer")
			assert.Equal(t, conf.GetLogDB().NameDB, "")
			assert.Equal(t, conf.GetLogDB().StorageNameDB, "application_template_db")
		})
	})

	t.Run("Тест чтения переменных окружения", func(t *testing.T) {
		t.Run("Тест 0. Проверка аутентификационных данных", func(t *testing.T) {
			passwdNetBox := "c8wyfihi8fdy9r8feguf82ry2r23"
			passwdZabbix := "superStrongPassWd"
			apiServerToken := "superStrongTokenForApiServer"
			passwdForLogDb := "superStrongPassWdForDatabAse"

			os.Setenv("GO_"+constants.App_Environment_Name+"_ZPASSWD", passwdZabbix)
			os.Setenv("GO_"+constants.App_Environment_Name+"_NBPASSWD", passwdNetBox)
			os.Setenv("GO_"+constants.App_Environment_Name+"_APISERVERTOKEN", apiServerToken)
			os.Setenv("GO_"+constants.App_Environment_Name+"_DBWLOGPASSWD", passwdForLogDb)

			conf, err := confighandler.New(constants.Root_Dir)
			assert.NoError(t, err)

			assert.Equal(t, conf.GetAuthenticationData().NetBoxPasswd, passwdNetBox)
			assert.Equal(t, conf.GetAuthenticationData().ZabbixPasswd, passwdZabbix)
			assert.Equal(t, conf.GetAuthenticationData().APIServerToken, apiServerToken)
			assert.Equal(t, conf.GetAuthenticationData().WriteLogBDPasswd, passwdForLogDb)
		})

		t.Run("Тест 2. Проверка настройки сервиса NetBox", func(t *testing.T) {
			os.Setenv("GO_"+constants.App_Environment_Name+"_NBHOST", "127.0.0.1")
			os.Setenv("GO_"+constants.App_Environment_Name+"_NBPORT", "4242")
			os.Setenv("GO_"+constants.App_Environment_Name+"_NBUSER", "some_user_service")

			conf, err := confighandler.New(constants.Root_Dir)
			assert.NoError(t, err)

			assert.Equal(t, conf.GetNetBox().Host, "127.0.0.1")
			assert.Equal(t, conf.GetNetBox().Port, 4242)
			assert.Equal(t, conf.GetNetBox().User, "some_user_service")
		})

		t.Run("Тест 3. Проверка настройки сервиса Zabbix", func(t *testing.T) {
			os.Setenv("GO_"+constants.App_Environment_Name+"_ZHOST", "127.0.0.1")
			os.Setenv("GO_"+constants.App_Environment_Name+"_ZPORT", "4242")
			os.Setenv("GO_"+constants.App_Environment_Name+"_ZUSER", "some_user_service")

			conf, err := confighandler.New(constants.Root_Dir)
			assert.NoError(t, err)

			assert.Equal(t, conf.GetZabbix().Host, "127.0.0.1")
			assert.Equal(t, conf.GetZabbix().Port, 4242)
			assert.Equal(t, conf.GetZabbix().User, "some_user_service")
		})

		t.Run("Тест 4. Проверка настройки сервиса InformationServerApi", func(t *testing.T) {
			const (
				host = "127.0.0.1"
				port = 4242
			)

			os.Setenv("GO_"+constants.App_Environment_Name+"_APIISHOST", host)
			os.Setenv("GO_"+constants.App_Environment_Name+"_APIISPORT", fmt.Sprint(port))

			conf, err := confighandler.New(constants.Root_Dir)
			assert.NoError(t, err)
			assert.Equal(t, conf.GetInformationServerApi().Host, host)
			assert.Equal(t, conf.GetInformationServerApi().Port, port)
		})

		t.Run("Тест 5. Проверка настройки WriteLogDataBase", func(t *testing.T) {
			os.Setenv("GO_"+constants.App_Environment_Name+"_DBWLOGHOST", "domaniname.database.cm")
			os.Setenv("GO_"+constants.App_Environment_Name+"_DBWLOGPORT", "8989")
			os.Setenv("GO_"+constants.App_Environment_Name+"_DBWLOGUSER", "somebody_user")
			os.Setenv("GO_"+constants.App_Environment_Name+"_DBWLOGNAME", "any_name_db")
			os.Setenv("GO_"+constants.App_Environment_Name+"_DBWLOGSTORAGENAME", "log_storage")

			conf, err := confighandler.New(constants.Root_Dir)
			assert.NoError(t, err)

			assert.Equal(t, conf.GetLogDB().Host, "domaniname.database.cm")
			assert.Equal(t, conf.GetLogDB().Port, 8989)
			assert.Equal(t, conf.GetLogDB().User, "somebody_user")
			assert.Equal(t, conf.GetLogDB().NameDB, "any_name_db")
			assert.Equal(t, conf.GetLogDB().StorageNameDB, "log_storage")
		})
	})

	t.Cleanup(func() {
		unSetEnviroment()
	})
}

func unSetEnviroment() {
	// Тип запуска приложения
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_MAIN")

	// Подключение к NetBox
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_NBHOST")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_NBPORT")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_NBUSER")

	// Подключение к Zabbix
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_ZHOST")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_ZPORT")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_ZUSER")

	// Подключение к API сервера информации
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_APIISHOST")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_APIISPORT")

	// Настройки доступа к БД в которую будут записыватся логи
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_DBWLOGHOST")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_DBWLOGPORT")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_DBWLOGNAME")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_DBWLOGUSER")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_DBWLOGSTORAGENAME")

	// Авторизационные данные
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_ZPASSWD")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_NBPASSWD")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_APISERVERTOKEN")
	os.Unsetenv("GO_" + constants.App_Environment_Name + "_DBWLOGPASSWD")
}
