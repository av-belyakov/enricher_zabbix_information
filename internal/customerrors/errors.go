package customerrors

import "fmt"

// ------ не корректный URL --------
// ErrorNoValidUrl ошибка при невалидном URL
type ErrorNoValidUrl struct {
	Message string
}

func NewErrorNoValidUrl(urlName string, err error) *ErrorNoValidUrl {
	return &ErrorNoValidUrl{
		Message: fmt.Sprintf("url '%s' contains invalid values, learn more (%+v)", urlName, err),
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

func NewErrorUrlNotFound(urlName string, err error) *ErrorUrlNotFound {
	return &ErrorUrlNotFound{
		Message: fmt.Sprintf("the url '%s' was not found, learn more (%+v)", urlName, err),
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

func NewErrorIpInvalid(ip string, err error) *ErrorIpInvalid {
	return &ErrorIpInvalid{
		Message: fmt.Sprintf("invalid ip address '%s' received, learn more (%+v)", ip, err),
	}
}

func (e *ErrorIpInvalid) Error() string {
	return e.Message
}
