package apiserver

import (
	"strings"

	"github.com/a-h/templ"

	"github.com/av-belyakov/enricher_zabbix_information/components"
	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
)

func (is *InformationServer) getBasePage(tmpComponent templ.Component, componentScript templ.ComponentScript) *templ.ComponentHandler {
	links := []struct {
		Name string
		Link string
		Icon string
	}{
		{
			Name: "начало",
			Link: "/",
		},
		{
			Name: "информация о выполненной задаче",
			Link: "task_information",
		},
		{
			Name: "общая статистика расходования памяти",
			Link: "memory_statistics",
		},
		{
			Name: "ручной запуск задачи",
			Link: "manually_task_starting",
		},
	}

	return templ.Handler(
		components.BasePage(datamodels.TemplBasePage{
			Title:      appname.GetName(),
			AppName:    strings.ToUpper(appname.GetName()),
			AppVersion: is.getAppVersion(),
			//AppShortInfo: hellowMsg,
			MenuLinks: links,
		},
			tmpComponent,
			componentScript,
		))
}

func (is *InformationServer) getAppVersion() string {
	version := "v0.0.1"
	if is.version != "" {
		version = "v" + is.version
	}

	return version
}
