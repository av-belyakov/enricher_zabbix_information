package dictionarieshandler

// ListDictionaries список словарей
type ListDictionaries struct {
	Dictionaries Dictionaries `yaml:"dictionaries"`
}

// Dictionaries словари
type Dictionaries struct {
	WebSiteGroupMonitoring []WebSiteMonitoring `mapstructure:"websites_group_monitoring"`
	Version                string              `yaml:"version"`
}

// WebSiteMonitoring информация о сайте для мониторинга
type WebSiteMonitoring struct {
	URLs        []string `yaml:"urls"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
}
