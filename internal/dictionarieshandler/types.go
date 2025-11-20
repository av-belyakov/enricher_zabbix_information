package dictionarieshandler

// ListDictionaries список словарей
type ListDictionaries struct {
	Dictionaries Dictionaries `yaml:"dictionaries"`
}

// Dictionaries словари
type Dictionaries struct {
	WebSiteGroupMonitoring []struct {
		URLs        []string `yaml:"urls"`
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
	} `mapstructure:"websites_group_monitoring"`
	Version string `yaml:"version"`
}
