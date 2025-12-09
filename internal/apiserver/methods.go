package apiserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
)

func (is *InformationServer) Start(ctx context.Context) error {
	routers := map[string]func(http.ResponseWriter, *http.Request){
		"/":                       is.RouteIndex,
		"/api":                    is.RouteApi,
		"/task_information":       is.RouteTaskInformation,
		"/memory_statistics":      is.RouteMemoryStatistics,
		"/manually_task_starting": is.RouteManuallyTaskStarting,
		"/logs":                   is.RouteLogs,
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
