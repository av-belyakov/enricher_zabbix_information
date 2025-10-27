package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appversion"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
)

func getInformationMessage(conf *confighandler.ConfigApp) string {
	version, err := appversion.GetVersion()
	if err != nil {
		log.Println(err)
	}

	appStatus := fmt.Sprintf("%vproduction%v", constants.Ansi_Bright_Blue, constants.Ansi_Reset)
	envValue, ok := os.LookupEnv("GO_ENRICHERZIN_MAIN")
	if ok && (envValue == "development" || envValue == "test") {
		appStatus = fmt.Sprintf("%v%s%v", constants.Ansi_Bright_Red, envValue, constants.Ansi_Reset)
	}

	msg := fmt.Sprintf("Application '%s' v%s was successfully launched", appname.GetName(), strings.Replace(version, "\n", "", -1))

	fmt.Printf("\n%v%v%s%v\n", constants.Bold_Font, constants.Ansi_Bright_Green, msg, constants.Ansi_Reset)
	fmt.Printf(
		"%v%vApplication status is '%s'%v\n",
		constants.Underlining,
		constants.Ansi_Bright_Green,
		appStatus,
		constants.Ansi_Reset,
	)
	fmt.Printf(
		"%vConnect to NetBox with address %v%s:%d%v%v, user %v'%s'%v\n",
		constants.Ansi_Bright_Green,
		constants.Ansi_Dark_Gray,
		conf.NetBox.Host,
		conf.NetBox.Port,
		constants.Ansi_Reset,
		constants.Ansi_Bright_Green,
		constants.Ansi_Dark_Gray,
		conf.NetBox.User,
		constants.Ansi_Reset,
	)
	fmt.Printf(
		"%vConnect to Zabbix with address %v%s:%d%v%v, user %v'%s'%v\n",
		constants.Ansi_Bright_Green,
		constants.Ansi_Dark_Gray,
		conf.Zabbix.Host,
		conf.Zabbix.Port,
		constants.Ansi_Reset,
		constants.Ansi_Bright_Green,
		constants.Ansi_Dark_Gray,
		conf.Zabbix.User,
		constants.Ansi_Reset,
	)

	return msg
}
