package dnsresolver

// ******** краткая информация о хосте ************
type ShortInformationAboutHost interface {
	GetHostId() int
	GetOriginalHost() string
}
