package zabbixinteraction

// New создание нового экземпляра модуля
func New(opts ...zabbixConnectionOptions) (*ZabbixApiModule, error) {
	api := &ZabbixApiModule{
		conn: 3,
	}

	for _, opt := range opts {
		if err := opt(api); err != nil {
			return api, err
		}
	}

	return nil, nil
}
