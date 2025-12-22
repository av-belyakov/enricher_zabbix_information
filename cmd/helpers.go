package main

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appversion"
	"github.com/av-belyakov/enricher_zabbix_information/internal/confighandler"
)

func getInformationMessage(conf *confighandler.ConfigApp, zabbixVersion string) string {
	version, err := appversion.GetVersion()
	if err != nil {
		log.Println(err)
	}

	appStatus := fmt.Sprintf("%vproduction%v", constants.Ansi_Bright_Blue, constants.Ansi_Reset)
	envValue, ok := os.LookupEnv("GO_" + constants.App_Environment_Name + "_MAIN")
	if ok && (envValue == "development" || envValue == "test") {
		appStatus = fmt.Sprintf("%v%s%v", constants.Ansi_Bright_Red, strings.ToUpper(envValue), constants.Ansi_Reset)
	}

	msg := fmt.Sprintf("Application '%s' v%s was successfully launched", appname.GetName(), strings.ReplaceAll(version, "\n", ""))

	fmt.Printf("\n%v%v%v%s%v\n", constants.Underlining, constants.Bold_Font, constants.Ansi_Bright_Green, msg, constants.Ansi_Reset)
	fmt.Printf(
		"%vApplication status is '%s%v'%v\n",
		constants.Ansi_Dark_Gray,
		appStatus,
		constants.Ansi_Dark_Gray,
		constants.Ansi_Reset,
	)
	fmt.Printf(
		"%vConnect to NetBox with address %v%s:%d%v%v user '%v%s%v'%v\n",
		constants.Ansi_Dark_Gray,
		constants.Ansi_Bright_Magenta,
		conf.GetNetBox().Host,
		conf.GetNetBox().Port,
		constants.Ansi_Reset,
		constants.Ansi_Dark_Gray,
		constants.Ansi_Bright_Magenta,
		conf.GetNetBox().User,
		constants.Ansi_Dark_Gray,
		constants.Ansi_Reset,
	)
	fmt.Printf(
		"%vConnect to Zabbix v%s with address %v%s:%d%v user '%v%s%v'%v\n",
		constants.Ansi_Dark_Gray,
		cmp.Or(zabbixVersion, "0.0.0"),
		constants.Ansi_Bright_Magenta,
		conf.GetZabbix().Host,
		conf.GetZabbix().Port,
		constants.Ansi_Dark_Gray,
		constants.Ansi_Bright_Magenta,
		conf.GetZabbix().User,
		constants.Ansi_Dark_Gray,
		constants.Ansi_Reset,
	)
	fmt.Printf(
		"%vConnect to write log data base with address %v%s:%d%v user '%v%s%v'%v\n",
		constants.Ansi_Dark_Gray,
		constants.Ansi_Bright_Magenta,
		conf.GetLogDB().Host,
		conf.GetLogDB().Port,
		constants.Ansi_Dark_Gray,
		constants.Ansi_Bright_Magenta,
		conf.GetLogDB().User,
		constants.Ansi_Dark_Gray,
		constants.Ansi_Reset,
	)
	fmt.Printf(
		"%vTask completion schedule. %v%s",
		constants.Ansi_Dark_Gray,
		constants.Ansi_Reset,
		getSchedule(conf.GetSchedule()),
	)

	return msg
}

func getSchedule(conf *confighandler.CfgSchedule) string {
	if len(conf.DailyJob) > 0 {
		s := strings.Builder{}
		s.WriteString(
			fmt.Sprintf(
				"%vTask start time list:%v\n",
				constants.Ansi_Dark_Gray,
				constants.Ansi_Reset,
			))
		for k, v := range conf.DailyJob {
			if v != "" {
				s.WriteString(
					fmt.Sprintf(
						"  %v%d. %v%s%v\n",
						constants.Ansi_Dark_Gray,
						k+1,
						constants.Ansi_Bright_Magenta,
						v,
						constants.Ansi_Reset,
					))
			}
		}

		return s.String()
	}

	return fmt.Sprintf(
		"%vTask launch frequency: %v%d%v min.%v\n",
		constants.Ansi_Dark_Gray,
		constants.Ansi_Bright_Magenta,
		conf.TimerJob,
		constants.Ansi_Dark_Gray,
		constants.Ansi_Reset,
	)
}
