package dictionarieshandler

// Dictionaries словари
type Dictionaries struct {
	WebSiteGroupMonitoring WebSiteGroupMonitoring `yaml:"websites_group_monitoring"`
}

// WebSiteGroupMonitoring информацию о группах сайтов предназначенных для мониторинга
type WebSiteGroupMonitoring struct {
	GroupMonitoring []struct {
		URLs        []string `yaml:"urls"`
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
	} `yaml:"group_monitoring"`
}
