package apiserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/interfaces"
)

func New(logger interfaces.Logger, storage interfaces.StorageInformation, opts ...informationServerOptions) (*InformationServer, error) {
	is := &InformationServer{
		version:   "0.0.1",
		timeStart: time.Now(),
		host:      "127.0.0.1",
		port:      7575,
		timeout:   time.Second * 10,
		logger:    logger,
		storage:   storage,
	}

	for _, opt := range opts {
		if err := opt(is); err != nil {
			return is, err
		}
	}

	return is, nil
}

func (is *InformationServer) Start(ctx context.Context) error {
	routers := map[string]func(http.ResponseWriter, *http.Request){
		"/":                       is.RouteIndex,
		"/api":                    is.RouteApi,
		"/task_information":       is.RouteTaskInformation,
		"/memory_statistics":      is.RouteMemoryStatistics,
		"/manually_task_starting": is.RouteManuallyTaskStarting,
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

	mux := http.NewServeMux()
	for k, v := range routers {
		mux.HandleFunc(k, v)
	}

	is.server = &http.Server{
		Addr:    net.JoinHostPort(is.host, fmt.Sprint(is.port)),
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

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
