package supportingfunctions

import "fmt"

// NumberType общий вспомогательный тип содержащий числовые значения
type NumberType interface {
	int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

// GetPointer вернет стрелку вверх, если b > a, иначе стрелку вниз
func GetPointerUpOrDown[T NumberType](a, b T) string {
	//fmt.Printf("Стрелка вверх: %q\nСтрелка вниз: %q\n", rune(0x2191), rune(0x2193))

	if b > a {
		return fmt.Sprintf("%q", rune(0x2191))
	} else if a > b {
		return fmt.Sprintf("%q", rune(0x2193))
	}

	return "=="
}
