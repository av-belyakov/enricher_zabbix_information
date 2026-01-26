package appstorage

import "errors"

func New(opts ...Options) (*SharedAppStorage, error) {
	as := &SharedAppStorage{
		statistics: StatisticsApp{
			data: make([]HostDetailedInformation, 0),
		},
		logs: LogsApp{
			story: make([]LogInformation, 0),
			size:  10,
		},
	}

	for _, opt := range opts {
		if err := opt(as); err != nil {
			return nil, err
		}
	}

	return as, nil
}

// WithSizeStatistics установить размер хранилища логов
func WithSizeLogs(size int) Options {
	return func(as *SharedAppStorage) error {
		if size < 10 {
			return errors.New("the size of the log storage cannot be less than 10")
		}

		as.logs.size = size

		return nil
	}
}
