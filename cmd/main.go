package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-ctx.Done()

		fmt.Printf("Module '%s' is stop", appname.GetName())

		stop()
	}()

	app(ctx)
}
