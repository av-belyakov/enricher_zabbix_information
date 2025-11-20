package zabbixinteraction

import (
	"errors"
)

// WithHost имя или ip адрес хоста API
func WithHost(v string) zabbixConnectionOptions {
	return func(api *ZabbixApiModule) error {
		if v == "" {
			return errors.New("the value of 'host' cannot be empty")
		}

		api.settings.host = v

		return nil
	}
}

// WithPort сетевой порт API
func WithPort(v int) zabbixConnectionOptions {
	return func(api *ZabbixApiModule) error {
		if v <= 0 || v > 65535 {
			return errors.New("an incorrect network port value was received")
		}

		api.settings.port = v

		return nil
	}
}

// WithLogin имя пользователя
func WithLogin(v string) zabbixConnectionOptions {
	return func(api *ZabbixApiModule) error {
		if v == "" {
			return errors.New("the value of 'login' cannot be empty")
		}

		api.settings.user = v

		return nil
	}
}

// WithPasswd пароль пользователя
func WithPasswd(v string) zabbixConnectionOptions {
	return func(api *ZabbixApiModule) error {
		if v == "" {
			return errors.New("the value of 'password' cannot be empty")
		}

		api.settings.passwd = v

		return nil
	}
}

// WithConnectionTimeout временной интервал соединения в секундах
func WithConnectionTimeout(v int) zabbixConnectionOptions {
	return func(api *ZabbixApiModule) error {
		if v <= 1 || v > 180 {
			return errors.New("an incorrect value, the value should be in the range from 1 to 1800")
		}

		api.settings.connectionTimeout = v

		return nil
	}
}

/*
// WithTLS использовать TLS
func WithTLS(v bool) zabbixConnectionOptions {
	return func(api *ZabbixConnectionJsonRPC) error {
		api.isTls = v

		return nil
	}
}

func WithFileRootCA(v ...string) zabbixConnectionOptions {
	return func(api *ZabbixConnectionJsonRPC) error {
		for _, cert := range v {
			if cert != "" {
				api.rootCAs = append(api.rootCAs, cert)
			}
		}

		return nil
	}
}
*/
