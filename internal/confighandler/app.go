package confighandler

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/supportingfunctions"
	"github.com/av-belyakov/enricher_zabbix_information/internal/wrappers"
)

func New(rootDir string) (*ConfigApp, error) {
	conf := &ConfigApp{}

	var (
		validate *validator.Validate
		envList  map[string]string = map[string]string{
			"GO_" + constants.Application_Name + "_MAIN": "",

			// Получение авторизационных данных
			"GO_" + constants.Application_Name + "_TOKEN":        "",
			"GO_" + constants.Application_Name + "_PASSWD":       "",
			"GO_" + constants.Application_Name + "_DBWLOGPASSWD": "",

			// Подключение к некоторому сервису Service
			"GO_" + constants.Application_Name + "_SHOST": "",
			"GO_" + constants.Application_Name + "_SPORT": "",
			"GO_" + constants.Application_Name + "_SUSER": "",

			// Настройки доступа к БД в которую будут записыватся логи
			"GO_" + constants.Application_Name + "_DBWLOGHOST":        "",
			"GO_" + constants.Application_Name + "_DBWLOGPORT":        "",
			"GO_" + constants.Application_Name + "_DBWLOGNAME":        "",
			"GO_" + constants.Application_Name + "_DBWLOGUSER":        "",
			"GO_" + constants.Application_Name + "_DBWLOGSTORAGENAME": "",
		}
	)

	getFileName := func(sf, confPath string, lfs []fs.DirEntry) (string, error) {
		for _, v := range lfs {
			if v.Name() == sf && !v.IsDir() {
				return filepath.Join(confPath, v.Name()), nil
			}
		}

		return "", fmt.Errorf("file '%s' is not found", sf)
	}

	setCommonSettings := func(fn string) error {
		viper.SetConfigFile(fn)
		viper.SetConfigType("yml")
		if err := viper.ReadInConfig(); err != nil {
			return err
		}

		ls := Logs{}
		if ok := viper.IsSet("LOGGING"); ok {
			if err := viper.GetViper().Unmarshal(&ls); err != nil {
				return err
			}

			conf.Common.Logs = ls.Logging
		}

		return nil
	}

	setSpecial := func(fn string) error {
		viper.SetConfigFile(fn)
		viper.SetConfigType("yml")
		if err := viper.ReadInConfig(); err != nil {
			return err
		}

		// Настройки для модуля подключения к Zabbix
		if viper.IsSet("Zabbix.host") {
			conf.Zabbix.Host = viper.GetString("Zabbix.host")
		}
		if viper.IsSet("Zabbix.port") {
			conf.Zabbix.Port = viper.GetInt("Zabbix.port")
		}
		if viper.IsSet("Zabbix.user") {
			conf.Zabbix.User = viper.GetString("Zabbix.user")
		}

		// Настройки для модуля подключения к NetBox
		if viper.IsSet("NetBox.host") {
			conf.NetBox.Host = viper.GetString("NetBox.host")
		}
		if viper.IsSet("NetBox.port") {
			conf.NetBox.Port = viper.GetInt("NetBox.port")
		}
		if viper.IsSet("NetBox.user") {
			conf.NetBox.User = viper.GetString("NetBox.user")
		}

		// Настройки доступа к БД в которую будут записыватся логи
		if viper.IsSet("WriteLogDataBase.host") {
			conf.LogDB.Host = viper.GetString("WriteLogDataBase.host")
		}
		if viper.IsSet("WriteLogDataBase.port") {
			conf.LogDB.Port = viper.GetInt("WriteLogDataBase.port")
		}
		if viper.IsSet("WriteLogDataBase.user") {
			conf.LogDB.User = viper.GetString("WriteLogDataBase.user")
		}
		if viper.IsSet("WriteLogDataBase.namedb") {
			conf.LogDB.NameDB = viper.GetString("WriteLogDataBase.namedb")
		}
		if viper.IsSet("WriteLogDataBase.storage_name_db") {
			conf.LogDB.StorageNameDB = viper.GetString("WriteLogDataBase.storage_name_db")
		}

		return nil
	}

	validate = validator.New(validator.WithRequiredStructEnabled())

	for v := range envList {
		if env, ok := os.LookupEnv(v); ok {
			envList[v] = env
		}
	}

	rootPath, err := supportingfunctions.GetRootPath(rootDir)
	if err != nil {
		return conf, wrappers.WrapperError(err)
	}

	confPath := filepath.Join(rootPath, "config")
	list, err := os.ReadDir(confPath)
	if err != nil {
		return conf, wrappers.WrapperError(err)
	}

	fileNameCommon, err := getFileName("config.yml", confPath, list)
	if err != nil {
		return conf, wrappers.WrapperError(err)
	}

	//чтение общего конфигурационного файла
	if err := setCommonSettings(fileNameCommon); err != nil {
		return conf, wrappers.WrapperError(err)
	}

	var fn string
	switch envList["GO_"+constants.Application_Name+"_MAIN"] {
	case "development":
		fn, err = getFileName("config_dev.yml", confPath, list)
		if err != nil {
			return conf, wrappers.WrapperError(err)
		}

	case "test":
		fn, err = getFileName("config_test.yml", confPath, list)
		if err != nil {
			return conf, wrappers.WrapperError(err)
		}

	default:
		fn, err = getFileName("config_prod.yml", confPath, list)
		if err != nil {
			return conf, wrappers.WrapperError(err)
		}

	}

	if err := setSpecial(fn); err != nil {
		return conf, wrappers.WrapperError(err)
	}

	// Настройки получения авторизационной информации
	if envList["GO_"+constants.Application_Name+"_TOKEN"] != "" {
		conf.AuthenticationData.SomeToken = envList["GO_"+constants.Application_Name+"_TOKEN"]
	}
	if envList["GO_"+constants.Application_Name+"_PASSWD"] != "" {
		conf.AuthenticationData.ServicePasswd = envList["GO_"+constants.Application_Name+"_PASSWD"]
	}
	if envList["GO_"+constants.Application_Name+"_DBWLOGPASSWD"] != "" {
		conf.AuthenticationData.WriteLogBDPasswd = envList["GO_"+constants.Application_Name+"_DBWLOGPASSWD"]
	}

	// Настройки для модуля подключения к некоторому сервису Service
	if envList["GO_"+constants.Application_Name+"_SHOST"] != "" {
		conf.Service.Host = envList["GO_"+constants.Application_Name+"_SHOST"]
	}
	if envList["GO_"+constants.Application_Name+"_SPORT"] != "" {
		if p, err := strconv.Atoi(envList["GO_"+constants.Application_Name+"_SPORT"]); err == nil {
			conf.Service.Port = p
		}
	}
	if envList["GO_"+constants.Application_Name+"_SUSER"] != "" {
		conf.Service.User = envList["GO_"+constants.Application_Name+"_SUSER"]
	}

	// Настройки доступа к БД в которую будут записыватся логи
	if envList["GO_"+constants.Application_Name+"_DBWLOGHOST"] != "" {
		conf.LogDB.Host = envList["GO_"+constants.Application_Name+"_DBWLOGHOST"]
	}
	if envList["GO_"+constants.Application_Name+"_DBWLOGPORT"] != "" {
		if p, err := strconv.Atoi(envList["GO_"+constants.Application_Name+"_DBWLOGPORT"]); err == nil {
			conf.LogDB.Port = p
		}
	}
	if envList["GO_"+constants.Application_Name+"_DBWLOGNAME"] != "" {
		conf.LogDB.NameDB = envList["GO_"+constants.Application_Name+"_DBWLOGNAME"]
	}
	if envList["GO_"+constants.Application_Name+"_DBWLOGUSER"] != "" {
		conf.LogDB.User = envList["GO_"+constants.Application_Name+"_DBWLOGUSER"]
	}
	if envList["GO_"+constants.Application_Name+"_DBWLOGSTORAGENAME"] != "" {
		conf.LogDB.StorageNameDB = envList["GO_"+constants.Application_Name+"_DBWLOGSTORAGENAME"]
	}

	//выполнение проверки заполненой структуры
	if err = validate.Struct(conf); err != nil {
		return conf, wrappers.WrapperError(err)
	}

	return conf, nil
}
