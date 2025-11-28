package helpersfile

// TypeExampleData тип декодирования json файла filesfortest/exampledata.json
type TypeExampleData struct {
	Hosts []struct {
		Name   string `json:"name"`
		Host   string `json:"host"`
		HostId string `json:"host_id"`
	}
}
