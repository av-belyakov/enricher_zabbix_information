package apiserver

import (
	"net/http"

	"github.com/av-belyakov/enricher_zabbix_information/components"
	"github.com/av-belyakov/enricher_zabbix_information/internal/memorystatistics"
)

// RouteMemoryStatistics статистика использования памяти
func (is *InformationServer) RouteMemoryStatistics(w http.ResponseWriter, r *http.Request) {
	is.getBasePage(
		components.MemoryStats(memorystatistics.GetMemoryStats()),
		components.BaseComponentScripts(),
	).Component.Render(r.Context(), w)
}
