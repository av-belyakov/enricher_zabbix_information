package zabbixinteraction

import (
	zc "github.com/av-belyakov/zabbixapicommunicator/v2"
)

type ZabbixApiModule struct {
	conn     *zc.ZabbixConnectionJsonRPC
	settings SettingsZabbixConnection
}

type SettingsZabbixConnection struct {
	host              string
	user              string
	passwd            string
	connectionTimeout int
	port              int
}

type zabbixConnectionOptions func(*ZabbixApiModule) error
