package confighandler

// ConfigApp конфигурационные настройки приложения
type ConfigApp struct {
	Common             CfgCommon
	LogDB              CfgWriteLogDB
	AuthenticationData CfgAuthenticationData
	Schedule           CfgSchedule
	NetBox             CfgNetBox
	Zabbix             CfgZabbix
}

// CfgCommon общие настройки
type CfgCommon struct {
	Logs []*LogSet
}

// Logs настройки логирования
type Logs struct {
	Logging []*LogSet
}

type LogSet struct {
	MsgTypeName   string `validate:"oneof=error info warning" yaml:"msgTypeName"`
	PathDirectory string `validate:"required" yaml:"pathDirectory"`
	MaxFileSize   int    `validate:"min=1000" yaml:"maxFileSize"`
	WritingStdout bool   `validate:"required" yaml:"writingStdout"`
	WritingFile   bool   `validate:"required" yaml:"writingFile"`
	WritingDB     bool   `validate:"required" yaml:"writingDB"`
}

// CfgSchedule настройки планирования запуска сервиса
type CfgSchedule struct {
	DailyJob DailyJobOptions `validate:"validateFn" yaml:"dailyJob"` //расписание в формате HH:MM
	TimerJob int             `validate:"lte=1439" yaml:"timerJob"`   //таймер в формате минут
}

type DailyJobOptions []string

// CfgWriteLogDB настройки записи данных в БД
type CfgWriteLogDB struct {
	Host          string `yaml:"host"`
	User          string `yaml:"user"`
	NameDB        string `yaml:"namedb"`
	StorageNameDB string `yaml:"storage_name_db"`
	Port          int    `validate:"gt=0,lte=65535" yaml:"port"`
}

// CfgNetBox настройки доступа к некоторому сервису
type CfgNetBox struct {
	Host string `validate:"required" yaml:"host"`
	User string `validate:"required" yaml:"user"`
	Port int    `validate:"gt=0,lte=65535" yaml:"port"`
}

// CfgZabbix настройки доступа к некоторому сервису
type CfgZabbix struct {
	Host    string `validate:"required" yaml:"host"`
	User    string `validate:"required" yaml:"user"`
	Port    int    `validate:"gt=0,lte=65535" yaml:"port"`
	Timeout int    `validate:"gte=0,lte=6000" yaml:"timeout"`
}

type CfgAuthenticationData struct {
	NetBoxPasswd     string `validate:"required"`
	ZabbixPasswd     string `validate:"required"`
	WriteLogBDPasswd string `yaml:"passwd"`
}
