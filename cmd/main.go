package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/av-belyakov/enricher_zabbix_information/constants"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appname"
	"github.com/av-belyakov/enricher_zabbix_information/internal/appversion"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-ctx.Done()

		version, _ := appversion.GetVersion()
		fmt.Printf(
			"%vThe module '%s' %v was successfully stopped.%v\n",
			constants.Ansi_Bright_Red,
			appname.GetName(),
			version,
			constants.Ansi_Reset,
		)

		stop()
	}()

	app(ctx)
}
