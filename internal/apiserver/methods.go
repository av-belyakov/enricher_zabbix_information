package apiserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/a-h/templ"
	"github.com/av-belyakov/enricher_zabbix_information/components"
	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
)

func (is *InformationServer) Start(ctx context.Context) error {
	routers := map[string]func(http.ResponseWriter, *http.Request){
		"/":                       is.RouteIndex,
		"/api":                    is.RouteApi,
		"/task_information":       is.RouteTaskInformation,
		"/memory_statistics":      is.RouteMemoryStatistics,
		"/manually_task_starting": is.RouteManuallyTaskStarting,
		"/sse":                    is.RouteSSE,
	}

	//отладка через pprof (только для тестов)
	//http://InformationServerApi.Host:InformationServerApi.Port/debug/pprof/
	//go tool pprof http://InformationServerApi.Host:InformationServerApi.Port/debug/pprof/heap
	//go tool pprof http://InformationServerApi.Host:InformationServerApi.Port/debug/pprof/allocs
	//go tool pprof http://InformationServerApi.Host:InformationServerApi.Port/debug/pprof/goroutine
	if os.Getenv("GO_"+constants.App_Environment_Name+"_MAIN") == "test" ||
		os.Getenv("GO_"+constants.App_Environment_Name+"_MAIN") == "development" {
		routers["/debug/pprof/"] = pprof.Index
	}

	//регистрируем обработчики маршрутов
	mux := http.NewServeMux()
	for k, v := range routers {
		mux.HandleFunc(k, v)
	}

	//инициализируем api сервер
	is.server = &http.Server{
		Addr:    net.JoinHostPort(is.host, fmt.Sprint(is.port)),
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	//инициализируем sse сервер

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return is.server.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()

		return is.server.Shutdown(context.Background())
	})

	return g.Wait()
}

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
		{
			Name: "интерактивные сообщения",
			Link: "sse",
		},
	}

	return templ.Handler(
		components.TemplateBasePage(datamodels.TemplBasePage{
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
