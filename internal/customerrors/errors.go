package customerrors

import "fmt"

// ------ не корректный URL --------
// ErrorNoValidUrl ошибка при невалидном URL
type ErrorNoValidUrl struct {
	Message string
}

func NewErrorNoValidUrl(urlName, description string) *ErrorNoValidUrl {
	return &ErrorNoValidUrl{
		Message: fmt.Sprintf("url '%s' contains invalid values, learn more (%s)", urlName, description),
	}
}

func (e *ErrorNoValidUrl) Error() string {
	return e.Message
}

// ------ URL не найден --------
// ErrorUrlNotFound ошибка, когда URL не найден в DNS
type ErrorUrlNotFound struct {
	Message string
}

func NewErrorUrlNotFound(urlName, description string) *ErrorUrlNotFound {
	return &ErrorUrlNotFound{
		Message: fmt.Sprintf("the url '%s' was not found, learn more (%s)", urlName, description),
	}
}

func (e *ErrorUrlNotFound) Error() string {
	return e.Message
}

// ------ не корректный ip адрес --------
// ErrorIpInvalid ошибка, получен некорректный ip адрес
type ErrorIpInvalid struct {
	Message string
}

func NewErrorIpInvalid(ip, description string) *ErrorIpInvalid {
	return &ErrorIpInvalid{
		Message: fmt.Sprintf("invalid ip address '%s' received, learn more (%s)", ip, description),
	}
}

func (e *ErrorIpInvalid) Error() string {
	return e.Message
}
