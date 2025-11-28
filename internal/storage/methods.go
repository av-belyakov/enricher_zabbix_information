package storage

import (
	"cmp"
	"reflect"
	"slices"
)

// GetList весь список
func (sts *ShortTermStorage) GetList() []HostDetailedInformation {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	list := make([]HostDetailedInformation, len(sts.data))
	copy(list, sts.data)

	return list
}

// GetForHostId данные по id хоста (быстрый поиск)
func (sts *ShortTermStorage) GetForHostId(hostId int) (HostDetailedInformation, bool) {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	sts.sort()
	index := sts.binarySearch(hostId)

	if index == -1 {
		return HostDetailedInformation{}, false
	}

	return sts.data[index], true
}

// GetForDomainName данные по домену (медленный поиск)
func (sts *ShortTermStorage) GetForDomainName(domaniName string) (int, HostDetailedInformation, bool) {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	index := sts.search("DomainName", domaniName)
	if index == -1 {
		return index, HostDetailedInformation{}, false
	}

	return index, sts.data[index], true
}

// GetForOriginalHost данные по неверифицированному названию хоста (медленный поиск)
func (sts *ShortTermStorage) GetForOriginalHost(originalHost string) (int, HostDetailedInformation, bool) {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	index := sts.search("OriginalHost", originalHost)
	if index == -1 {
		return index, HostDetailedInformation{}, false
	}

	return index, sts.data[index], true
}

// Add добавляет элемент в хранилище
func (sts *ShortTermStorage) Add(event HostDetailedInformation) {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	sts.data = append(sts.data, event)
}

// DeleteElement удаляет заданный элемент по hostId
func (sts *ShortTermStorage) DeleteElement(hostId int) {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	sts.sort()
	if index := sts.binarySearch(hostId); index != -1 {
		if index < 0 || index >= len(sts.data) {
			return
		}

		sts.data = append(sts.data[:index], sts.data[index+1:]...)
	}
}

// DeleteAll удаляет все элементы из хранилища
func (sts *ShortTermStorage) DeleteAll() {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	sts.data = []HostDetailedInformation{}
}

// GetListErrors список объектов с ошибками
func (sts *ShortTermStorage) GetListErrors() []HostDetailedInformation {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	list := make([]HostDetailedInformation, 0)
	for _, v := range sts.data {
		if v.Errors != nil {
			list = append(list, v)
		}
	}

	return list
}

// sort сортировка
func (sts *ShortTermStorage) sort() {
	slices.SortFunc(sts.data, func(a, b HostDetailedInformation) int {
		return cmp.Compare(a.HostId, b.HostId)
	})
}

// binarySearch выполняет стандартный двоичный поиск.
// Возвращает индекс целевого объекта, если он найден, или -1, если не найден.
func (sts *ShortTermStorage) binarySearch(id int) int {
	left, right := 0, len(sts.data)

	for left < right {
		mid := left + (right-left)/2
		if sts.data[mid].HostId == id {
			return mid
		} else if sts.data[mid].HostId < id {
			left = mid + 1
		} else {
			right = mid
		}
	}

	return -1
}

func (sts *ShortTermStorage) search(nameElem, valueElem string) int {
	for k, v := range sts.data {
		fields := reflect.TypeOf(v)
		values := reflect.ValueOf(v)

		for i := 0; i < fields.NumField(); i++ {
			if fields.Field(i).Name != nameElem {
				continue
			}

			if fields.Field(i).Type.Kind() == reflect.String {
				if values.Field(i).String() == valueElem {
					return k
				}
			}
		}
	}

	return -1
}
