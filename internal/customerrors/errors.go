package customerrors

import "fmt"

// ErrorNoValidUrl ошибка при невалидном URL
type ErrorNoValidUrl struct {
	Message string
}

func NewErrorNoValidUrl(urlName string) *ErrorNoValidUrl {
	return &ErrorNoValidUrl{
		Message: fmt.Sprintf("url '%s' contains invalid values", urlName),
	}
}

func (e *ErrorNoValidUrl) Error() string {
	return e.Message
}
