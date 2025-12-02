package dnsresolver

import "errors"

// WithTimeout время ожидания выполнения запроса
func WithTimeout(timeout int) Options {
	return func(s *Settings) error {
		if timeout == 0 {
			return errors.New("timeout cannot be zero")
		}

		return nil
	}
}
