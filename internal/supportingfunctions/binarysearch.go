package supportingfunctions

import "slices"

// BinarySearch ыполняет стандартный двоичный поиск, чтобы найти целевой объект в отсортированном массиве.
// Возвращает индекс целевого объекта, если он найден, или -1, если не найден.
func BinarySearch(arr []int, target int) int {
	slices.Sort(arr)
	left, right := 0, len(arr)

	for left < right {
		mid := left + (right-left)/2
		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			left = mid + 1
		} else {
			right = mid
		}
	}

	return -1
}
