package confighandler_test

import (
	"log"
	"os"
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
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatalln(err)
	}

	os.Setenv("GO_"+constants.Application_Name+"_MAIN", "test")

	conf, err = confighandler.New(constants.Root_Dir)
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(m.Run())
}

func TestReadConfigHandler(t *testing.T) {
	t.Run("Тест чтения конфигурационного файла", func(t *testing.T) {
		t.Run("Тест 1. Проверка аутентификационных данных", func(t *testing.T) {
			assert.Equal(t, conf.GetAuthenticationData().SomeToken, "yoursometoken")
			assert.Equal(t, conf.GetAuthenticationData().ServicePasswd, "yoursomepassword")
			assert.Equal(t, conf.GetAuthenticationData().WriteLogBDPasswd, "yoursomepasswordfordatabase")
		})

		t.Run("Тест 2. Проверка настройки Service из файла config_test.yml", func(t *testing.T) {
			assert.Equal(t, conf.GetService().Host, "localhost")
			assert.Equal(t, conf.GetService().Port, 80)
			assert.Equal(t, conf.GetService().User, "user-name")
		})

		t.Run("Тест 3. Проверка настройки WriteLogDataBase из файла config_dev.yml", func(t *testing.T) {
			assert.Equal(t, conf.GetLogDB().Host, "database.cloud.example")
			assert.Equal(t, conf.GetLogDB().Port, 9200)
			assert.Equal(t, conf.GetLogDB().User, "log_writer")
			assert.Equal(t, conf.GetLogDB().NameDB, "")
			assert.Equal(t, conf.GetLogDB().StorageNameDB, "application_template_db")
		})
	})

	t.Run("Тест чтения переменных окружения", func(t *testing.T) {
		t.Run("Тест 0. Проверка аутентификационных данных", func(t *testing.T) {
			token := "c8wyfihi8fdy9r8feguf82ry2r23"
			password := "superStrongPassWd"
			passwdForDb := "superStrongPassWdForDatabAse"

			os.Setenv("GO_"+constants.Application_Name+"_TOKEN", token)
			os.Setenv("GO_"+constants.Application_Name+"_PASSWD", password)
			os.Setenv("GO_"+constants.Application_Name+"_DBWLOGPASSWD", passwdForDb)

			conf, err := confighandler.New(constants.Root_Dir)
			assert.NoError(t, err)

			assert.Equal(t, conf.GetAuthenticationData().SomeToken, token)
			assert.Equal(t, conf.GetAuthenticationData().ServicePasswd, password)
			assert.Equal(t, conf.GetAuthenticationData().WriteLogBDPasswd, passwdForDb)
		})

		t.Run("Тест 2. Проверка настройки некоторого сервиса", func(t *testing.T) {
			os.Setenv("GO_"+constants.Application_Name+"_SHOST", "127.0.0.1")
			os.Setenv("GO_"+constants.Application_Name+"_SPORT", "4242")
			os.Setenv("GO_"+constants.Application_Name+"_SUSER", "some_user_service")

			conf, err := confighandler.New(constants.Root_Dir)
			assert.NoError(t, err)

			assert.Equal(t, conf.GetService().Host, "127.0.0.1")
			assert.Equal(t, conf.GetService().Port, 4242)
			assert.Equal(t, conf.GetService().User, "some_user_service")
		})

		t.Run("Тест 3. Проверка настройки WriteLogDataBase", func(t *testing.T) {
			os.Setenv("GO_"+constants.Application_Name+"_DBWLOGHOST", "domaniname.database.cm")
			os.Setenv("GO_"+constants.Application_Name+"_DBWLOGPORT", "8989")
			os.Setenv("GO_"+constants.Application_Name+"_DBWLOGUSER", "somebody_user")
			os.Setenv("GO_"+constants.Application_Name+"_DBWLOGNAME", "any_name_db")
			os.Setenv("GO_"+constants.Application_Name+"_DBWLOGSTORAGENAME", "log_storage")

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
	os.Unsetenv("GO_" + constants.Application_Name + "_MAIN")

	// Подключение к некоторому сервису Service
	os.Unsetenv("GO_" + constants.Application_Name + "_SHOST")
	os.Unsetenv("GO_" + constants.Application_Name + "_SPORT")
	os.Unsetenv("GO_" + constants.Application_Name + "_SUSER")

	// Настройки доступа к БД в которую будут записыватся логи
	os.Unsetenv("GO_" + constants.Application_Name + "_DBWLOGHOST")
	os.Unsetenv("GO_" + constants.Application_Name + "_DBWLOGPORT")
	os.Unsetenv("GO_" + constants.Application_Name + "_DBWLOGNAME")
	os.Unsetenv("GO_" + constants.Application_Name + "_DBWLOGUSER")
	os.Unsetenv("GO_" + constants.Application_Name + "_DBWLOGSTORAGENAME")

	// Авторизационные данные
	os.Unsetenv("GO_" + constants.Application_Name + "_TOKEN")
	os.Unsetenv("GO_" + constants.Application_Name + "_PASSWD")
	os.Unsetenv("GO_" + constants.Application_Name + "_DBWLOGPASSWD")
}
