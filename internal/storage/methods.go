package storage

import (
	"cmp"
	"fmt"
	"net/netip"
	"reflect"
	"slices"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/datamodels"
)

// GetStatusProcessRunning получить статус выполнение процесса
func (sts *ShortTermStorage) GetStatusProcessRunning() bool {
	return sts.isExecution.Load()
}

// SetProcessRunning установить статус 'процесс выполняется'
func (sts *ShortTermStorage) SetProcessRunning() {
	sts.isExecution.Store(true)
	sts.SetStartDateExecution()
}

// SetProcessNotRunning установить статус 'процесс не выполняется'
func (sts *ShortTermStorage) SetProcessNotRunning() {
	sts.isExecution.Store(false)
	sts.SetEndDateExecution()
}

// GetDateExecution получить дату начала и окончания выполнения процесса
func (sts *ShortTermStorage) GetDateExecution() (start, end time.Time) {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	return sts.startDateExecution, sts.endDateExecution
}

// SetDateExecution установить дату начала выполнения процесса
func (sts *ShortTermStorage) SetStartDateExecution() {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	sts.startDateExecution = time.Now()
}

// SetDateExecution установить дату окончания выполнения процесса
func (sts *ShortTermStorage) SetEndDateExecution() {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	sts.endDateExecution = time.Now()
}

// GetList список подробной информации о хостах
func (sts *ShortTermStorage) GetList() []datamodels.HostDetailedInformation {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	list := make([]datamodels.HostDetailedInformation, len(sts.data))
	copy(list, sts.data)

	return list
}

// GetHosts список всех хостов
func (sts *ShortTermStorage) GetHosts() map[int]string {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	newList := make(map[int]string, len(sts.data))

	for _, v := range sts.data {
		newList[v.HostId] = v.OriginalHost
	}

	return newList
}

// GetForHostId данные по id хоста (быстрый поиск)
func (sts *ShortTermStorage) GetForHostId(hostId int) (int, datamodels.HostDetailedInformation, bool) {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	sts.sort()
	index := sts.binarySearch(hostId)

	if index == -1 {
		return index, datamodels.HostDetailedInformation{}, false
	}

	return index, sts.data[index], true
}

// GetForDomainName данные по домену (медленный поиск)
func (sts *ShortTermStorage) GetForDomainName(domaniName string) (int, datamodels.HostDetailedInformation, bool) {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	index := sts.search("DomainName", domaniName)
	if index == -1 {
		return index, datamodels.HostDetailedInformation{}, false
	}

	return index, sts.data[index], true
}

// GetForOriginalHost данные по неверифицированному названию хоста (медленный поиск)
func (sts *ShortTermStorage) GetForOriginalHost(originalHost string) (int, datamodels.HostDetailedInformation, bool) {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	index := sts.search("OriginalHost", originalHost)
	if index == -1 {
		return index, datamodels.HostDetailedInformation{}, false
	}

	return index, sts.data[index], true
}

// Add добавляет элемент в хранилище
func (sts *ShortTermStorage) Add(event datamodels.HostDetailedInformation) {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	sts.data = append(sts.data, event)
}

// SetDomainName устанавливает доменное имя для заданного id хоста
func (sts *ShortTermStorage) SetDomainName(hostId int, domainName string) error {
	index, elem, ok := sts.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	elem.DomainName = domainName
	sts.data[index] = elem

	return nil
}

// SetIps устанавливает ip адреса для заданного id хоста
func (sts *ShortTermStorage) SetIps(hostId int, ip netip.Addr, ips ...netip.Addr) error {
	index, elem, ok := sts.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	ips = append(ips, ip)
	elem.Ips = append(elem.Ips, ips...)
	sts.data[index] = elem

	return nil
}

// SetError устанавливает описание ошибки для заданного id хоста
func (sts *ShortTermStorage) SetError(hostId int, err error) error {
	index, elem, ok := sts.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	elem.Error = err
	sts.data[index] = elem

	return nil
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

// DeleteAll удаляет все элементы из хранилища, выставляет время начала и конца
// выполнения в дату начала эпохи Unix, выставляет статус выполнения в
// 'не выполняется' (false)
func (sts *ShortTermStorage) DeleteAll() {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	sts.startDateExecution = time.Time{}
	sts.endDateExecution = time.Time{}
	sts.isExecution.Store(false)

	sts.data = []datamodels.HostDetailedInformation{}
}

// GetListErrors список объектов с ошибками
func (sts *ShortTermStorage) GetListErrors() []datamodels.HostDetailedInformation {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	list := make([]datamodels.HostDetailedInformation, 0)
	for _, v := range sts.data {
		if v.Error != nil {
			list = append(list, v)
		}
	}

	return list
}

// sort сортировка
func (sts *ShortTermStorage) sort() {
	slices.SortFunc(sts.data, func(a, b datamodels.HostDetailedInformation) int {
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
