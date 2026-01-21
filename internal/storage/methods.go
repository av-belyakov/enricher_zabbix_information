package storage

import (
	"cmp"
	"fmt"
	"net/netip"
	"reflect"
	"slices"
	"time"
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
func (sts *ShortTermStorage) GetList() []HostDetailedInformation {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	list := make([]HostDetailedInformation, len(sts.data))
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
func (sts *ShortTermStorage) GetForHostId(hostId int) (int, HostDetailedInformation, bool) {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	sts.sort()
	index := sts.binarySearch(hostId)

	if index == -1 {
		return index, HostDetailedInformation{}, false
	}

	return index, sts.data[index], true
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

// GetHostsWithSensorId список хостов с id обслуживающего сенсора
func (sts *ShortTermStorage) GetHostsWithSensorId() []HostDetailedInformation {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	var hostList []HostDetailedInformation
	for _, v := range sts.data {
		if len(v.SensorsId) == 0 {
			continue
		}

		hostList = append(hostList, v)
	}

	return hostList
}

// Add добавляет элемент в хранилище
func (sts *ShortTermStorage) Add(event HostDetailedInformation) {
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

	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	elem.DomainName = domainName
	sts.data[index] = elem

	return nil
}

// SetIps устанавливает ip адреса для заданного id хоста
func (sts *ShortTermStorage) SetIps(hostId int, ips ...netip.Addr) error {
	index, elem, ok := sts.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	for _, ip := range ips {
		elem.Ips = append(elem.Ips, ip)
	}
	sts.data[index] = elem

	return nil
}

// SetIsProcessed устанавливает статус 'обработано' для заданного id хоста
func (sts *ShortTermStorage) SetIsProcessed(hostId int) error {
	index, elem, ok := sts.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	elem.IsProcessed = true
	sts.data[index] = elem

	return nil
}

// SetError устанавливает описание ошибки для заданного id хоста
func (sts *ShortTermStorage) SetError(hostId int, err error) error {
	index, elem, ok := sts.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	elem.Error = err
	sts.data[index] = elem

	return nil
}

// SetSensorId устанавливает id обслуживающего сенсора для заданного id хоста
func (sts *ShortTermStorage) SetSensorId(hostId int, sensorsId ...string) error {
	index, elem, ok := sts.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	for _, sensorId := range sensorsId {
		if !slices.Contains(elem.SensorsId, sensorId) {
			elem.SensorsId = append(elem.SensorsId, sensorId)
		}
	}

	sts.data[index] = elem

	return nil
}

// SetIsActive устанавливает флаг активности для заданного id хоста, то есть хост
// в Netbox отмечен как активный
func (sts *ShortTermStorage) SetIsActive(hostId int) error {
	index, elem, ok := sts.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	elem.IsActive = true
	sts.data[index] = elem

	return nil
}

// SetNetboxHostId устанавливает id хоста в Netbox для заданного id хоста
func (sts *ShortTermStorage) SetNetboxHostId(hostId int, netboxHostsId ...int) error {
	index, elem, ok := sts.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	for _, netboxHostId := range netboxHostsId {
		if !slices.Contains(elem.NetboxHostsId, netboxHostId) {
			elem.NetboxHostsId = append(elem.NetboxHostsId, netboxHostId)
		}
	}

	sts.data[index] = elem

	return nil
}

// SetCountZabbixHostsGroup количество групп хостов в Zabbix
func (sts *ShortTermStorage) SetCountZabbixHostsGroup(v int) {
	sts.countZabbixHostsGroup.Store(int32(v))
}

// GetCountZabbixHostsGroup количество групп хостов в Zabbix
func (sts *ShortTermStorage) GetCountZabbixHostsGroup() int32 {
	return sts.countZabbixHostsGroup.Load()
}

// SetCountZabbixHosts общее количество хостов в Zabbix
func (sts *ShortTermStorage) SetCountZabbixHosts(v int) {
	sts.countZabbixHosts.Store(int32(v))
}

// GetCountZabbixHosts общее количество хостов в Zabbix
func (sts *ShortTermStorage) GetCountZabbixHosts() int32 {
	return sts.countZabbixHosts.Load()
}

// SetCountMonitoringHostsGroup количество групп хостов по которым осуществляется мониторинг
func (sts *ShortTermStorage) SetCountMonitoringHostsGroup(v int) {
	sts.countMonitoringHostsGroup.Store(int32(v))
}

// GetCountMonitoringHostsGroup количество групп хостов по которым осуществляется мониторинг
func (sts *ShortTermStorage) GetCountMonitoringHostsGroup() int32 {
	return sts.countMonitoringHostsGroup.Load()
}

// SetCountMonitoringHosts количество хостов по которым осуществляется мониторинг
func (sts *ShortTermStorage) SetCountMonitoringHosts(v int) {
	sts.countMonitoringHosts.Store(int32(v))
}

// GetCountMonitoringHosts количество хостов по которым осуществляется мониторинг
func (sts *ShortTermStorage) GetCountMonitoringHosts() int32 {
	return sts.countMonitoringHosts.Load()
}

// SetCountNetboxPrefixes количество найденных префиксов в Netbox
func (sts *ShortTermStorage) SetCountNetboxPrefixes(v int) {
	sts.countNetboxPrefixes.Store(int32(v))
}

// GetCountNetboxPrefixes количество найденных префиксов в Netbox
func (sts *ShortTermStorage) GetCountNetboxPrefixes() int32 {
	return sts.countNetboxPrefixes.Load()
}

// SetCountUpdatedZabbixHosts количество обновленных хостов в Zabbix
func (sts *ShortTermStorage) SetCountUpdatedZabbixHosts(v int) {
	sts.countUpdatedZabbixHosts.Store(int32(v))
}

// GetCountUpdatedZabbixHosts количество обновленных хостов в Zabbix
func (sts *ShortTermStorage) GetCountUpdatedZabbixHosts() int32 {
	return sts.countUpdatedZabbixHosts.Load()
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

	sts.data = []HostDetailedInformation{}
}

// GetListErrors список объектов с ошибками
func (sts *ShortTermStorage) GetListErrors() []HostDetailedInformation {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	list := make([]HostDetailedInformation, 0)
	for _, v := range sts.data {
		if v.Error != nil {
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
