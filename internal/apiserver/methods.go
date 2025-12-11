package apiserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"

	"github.com/a-h/templ"
	"golang.org/x/sync/errgroup"

	"github.com/av-belyakov/enricher_zabbix_information/components"
	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
	"github.com/av-belyakov/enricher_zabbix_information/internal/websocketserver"
)

func (is *InformationServer) Start(ctx context.Context) error {
	wsServer := websocketserver.New()
	routers := map[string]func(http.ResponseWriter, *http.Request){
		"/":                       is.RouteIndex,
		"/task_information":       is.RouteTaskInformation,
		"/memory_statistics":      is.RouteMemoryStatistics,
		"/manually_task_starting": is.RouteManuallyTaskStarting,
		"/logs":                   is.RouteLogs,
		"/ws": func(w http.ResponseWriter, r *http.Request) {
			websocketserver.ServeWs(is.logger, wsServer, w, r)
		},
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

	//запускаем ws сервер и получаем канал с входящими данными
	chIncomingData := wsServer.Run(ctx)

	//обработчик входящих и исходящих данных
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case msg := <-is.chInput:
				wsServer.SendBroadcast(msg)

			case msg := <-chIncomingData:
				is.chOutput <- msg

			}
		}
	}()

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

// SendData отправка данных в APIServer
func (is *InformationServer) SendData(b []byte) {
	is.chInput <- b
}

// GetChannelOutgoingData канал с исходящими от APIServer данными
func (is *InformationServer) GetChannelOutgoingData() <-chan []byte {
	return is.chOutput
}

func (is *InformationServer) getBasePage(tmpComponent templ.Component, componentScript templ.ComponentScript) *templ.ComponentHandler {
	links := []struct {
		Name string
		Link string
		Icon string
	}{
		{
			Name: "главная страница",
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
			Name: "логи приложения",
			Link: "logs",
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
