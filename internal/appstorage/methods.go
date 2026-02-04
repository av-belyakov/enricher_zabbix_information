package appstorage

import (
	"cmp"
	"fmt"
	"net/netip"
	"reflect"
	"slices"
	"time"

	"github.com/av-belyakov/enricher_zabbix_information/internal/dnsresolver"
)

//********************* статистическая информация ***********************

// GetStatusProcessRunning получить статус выполнение процесса
func (as *SharedAppStorage) GetStatusProcessRunning() bool {
	return as.statistics.isExecution.Load()
}

// SetProcessRunning установить статус 'процесс выполняется'
func (as *SharedAppStorage) SetProcessRunning() {
	as.statistics.isExecution.Store(true)
	as.SetStartDateExecution()
}

// SetProcessNotRunning установить статус 'процесс не выполняется'
func (as *SharedAppStorage) SetProcessNotRunning() {
	as.statistics.isExecution.Store(false)
	as.SetEndDateExecution()
}

// GetDateExecution получить дату начала и окончания выполнения процесса
func (as *SharedAppStorage) GetDateExecution() (start, end time.Time) {
	as.statistics.mutex.RLock()
	defer as.statistics.mutex.RUnlock()

	return as.statistics.startDateExecution, as.statistics.endDateExecution
}

// SetDateExecution установить дату начала выполнения процесса
func (as *SharedAppStorage) SetStartDateExecution() {
	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	as.statistics.startDateExecution = time.Now()
}

// SetDateExecution установить дату окончания выполнения процесса
func (as *SharedAppStorage) SetEndDateExecution() {
	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	as.statistics.endDateExecution = time.Now()
}

// GetList список подробной информации о хостах
func (as *SharedAppStorage) GetList() []HostDetailedInformation {
	as.statistics.mutex.RLock()
	defer as.statistics.mutex.RUnlock()

	list := make([]HostDetailedInformation, len(as.statistics.data))
	copy(list, as.statistics.data)

	return list
}

// GetHosts список всех хостов
func (as *SharedAppStorage) GetHosts() []dnsresolver.ShortInformationAboutHost {
	as.statistics.mutex.RLock()
	defer as.statistics.mutex.RUnlock()

	newList := make([]dnsresolver.ShortInformationAboutHost, 0, len(as.statistics.data))
	for _, v := range as.statistics.data {
		newList = append(newList, &v)
	}

	return newList
}

// GetProcessedHosts список обработанных хостов
func (as *SharedAppStorage) GetProcessedHosts() []HostDetailedInformation {
	as.statistics.mutex.RLock()
	defer as.statistics.mutex.RUnlock()

	newList := []HostDetailedInformation{}

	for _, v := range as.statistics.data {
		if !v.IsProcessed {
			continue
		}

		newList = append(newList, v)
	}

	return newList
}

// GetForHostId данные по id хоста (быстрый поиск)
func (as *SharedAppStorage) GetForHostId(hostId int) (int, HostDetailedInformation, bool) {
	as.statistics.mutex.RLock()
	defer as.statistics.mutex.RUnlock()

	as.statistics.sort()
	index := as.statistics.binarySearch(hostId)

	if index == -1 {
		return index, HostDetailedInformation{}, false
	}

	return index, as.statistics.data[index], true
}

// GetForDomainName данные по домену (медленный поиск)
func (as *SharedAppStorage) GetForDomainName(domaniName string) (int, HostDetailedInformation, bool) {
	as.statistics.mutex.RLock()
	defer as.statistics.mutex.RUnlock()

	index := as.statistics.search("DomainName", domaniName)
	if index == -1 {
		return index, HostDetailedInformation{}, false
	}

	return index, as.statistics.data[index], true
}

// GetForOriginalHost данные по неверифицированному названию хоста (медленный поиск)
func (as *SharedAppStorage) GetForOriginalHost(originalHost string) (int, HostDetailedInformation, bool) {
	as.statistics.mutex.RLock()
	defer as.statistics.mutex.RUnlock()

	index := as.statistics.search("OriginalHost", originalHost)
	if index == -1 {
		return index, HostDetailedInformation{}, false
	}

	return index, as.statistics.data[index], true
}

// GetHostsWithSensorId список хостов с id обслуживающего сенсора
func (as *SharedAppStorage) GetHostsWithSensorId() []HostDetailedInformation {
	as.statistics.mutex.RLock()
	defer as.statistics.mutex.RUnlock()

	var hostList []HostDetailedInformation
	for _, v := range as.statistics.data {
		if len(v.SensorsId) == 0 {
			continue
		}

		hostList = append(hostList, v)
	}

	return hostList
}

// AddElement добавляет элемент в хранилище
func (as *SharedAppStorage) AddElement(event HostDetailedInformation) {
	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	as.statistics.data = append(as.statistics.data, event)
}

// SetDomainName устанавливает доменное имя для заданного id хоста
func (as *SharedAppStorage) SetDomainName(hostId int, domainName string) error {
	index, elem, ok := as.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	elem.DomainName = domainName
	as.statistics.data[index] = elem

	return nil
}

// SetIps устанавливает ip адреса для заданного id хоста
func (as *SharedAppStorage) SetIps(hostId int, ips ...netip.Addr) error {
	index, elem, ok := as.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	for _, ip := range ips {
		elem.Ips = append(elem.Ips, ip)
	}
	as.statistics.data[index] = elem

	return nil
}

// SetIsProcessed устанавливает статус 'обработано' для заданного id хоста
func (as *SharedAppStorage) SetIsProcessed(hostId int) error {
	index, _, ok := as.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	as.statistics.data[index].IsProcessed = true

	return nil
}

// SetError устанавливает описание ошибки для заданного id хоста
func (as *SharedAppStorage) SetError(hostId int, err error) error {
	index, elem, ok := as.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	elem.Error = err
	as.statistics.data[index] = elem

	return nil
}

// SetSensorId устанавливает id обслуживающего сенсора для заданного id хоста
func (as *SharedAppStorage) SetSensorId(hostId int, sensorsId ...string) error {
	index, elem, ok := as.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	for _, sensorId := range sensorsId {
		if sensorId == "" {
			continue
		}

		if !slices.Contains(elem.SensorsId, sensorId) {
			elem.SensorsId = append(elem.SensorsId, sensorId)
		}
	}

	as.statistics.data[index] = elem

	return nil
}

// SetIsActive устанавливает флаг активности для заданного id хоста, то есть хост
// в Netbox отмечен как активный
func (as *SharedAppStorage) SetIsActive(hostId int) error {
	index, elem, ok := as.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	elem.IsActive = true
	as.statistics.data[index] = elem

	return nil
}

// SetNetboxHostId устанавливает id хоста в Netbox для заданного id хоста
func (as *SharedAppStorage) SetNetboxHostId(hostId int, netboxHostsId ...int) error {
	index, elem, ok := as.GetForHostId(hostId)
	if !ok {
		return fmt.Errorf("the element with hostId '%d' was not found", hostId)
	}

	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	for _, netboxHostId := range netboxHostsId {
		if netboxHostId == 0 {
			continue
		}

		if !slices.Contains(elem.NetboxHostsId, netboxHostId) {
			elem.NetboxHostsId = append(elem.NetboxHostsId, netboxHostId)
		}
	}

	as.statistics.data[index] = elem

	return nil
}

// SetCountZabbixHostsGroup количество групп хостов в Zabbix
func (as *SharedAppStorage) SetCountZabbixHostsGroup(v int) {
	as.statistics.countZabbixHostsGroup.Store(int32(v))
}

// GetCountZabbixHostsGroup количество групп хостов в Zabbix
func (as *SharedAppStorage) GetCountZabbixHostsGroup() int32 {
	return as.statistics.countZabbixHostsGroup.Load()
}

// SetCountZabbixHosts общее количество хостов в Zabbix
func (as *SharedAppStorage) SetCountZabbixHosts(v int) {
	as.statistics.countZabbixHosts.Store(int32(v))
}

// GetCountZabbixHosts общее количество хостов в Zabbix
func (as *SharedAppStorage) GetCountZabbixHosts() int32 {
	return as.statistics.countZabbixHosts.Load()
}

// SetCountMonitoringHostsGroup количество групп хостов по которым осуществляется мониторинг
func (as *SharedAppStorage) SetCountMonitoringHostsGroup(v int) {
	as.statistics.countMonitoringHostsGroup.Store(int32(v))
}

// GetCountMonitoringHostsGroup количество групп хостов по которым осуществляется мониторинг
func (as *SharedAppStorage) GetCountMonitoringHostsGroup() int32 {
	return as.statistics.countMonitoringHostsGroup.Load()
}

// SetCountMonitoringHosts количество хостов по которым осуществляется мониторинг
func (as *SharedAppStorage) SetCountMonitoringHosts(v int) {
	as.statistics.countMonitoringHosts.Store(int32(v))
}

// GetCountMonitoringHosts количество хостов по которым осуществляется мониторинг
func (as *SharedAppStorage) GetCountMonitoringHosts() int32 {
	return as.statistics.countMonitoringHosts.Load()
}

// SetCountNetboxPrefixes количество найденных префиксов в Netbox
func (as *SharedAppStorage) SetCountNetboxPrefixes(v int) {
	as.statistics.countNetboxPrefixes.Store(int32(v))
}

// GetCountNetboxPrefixes количество найденных префиксов в Netbox
func (as *SharedAppStorage) GetCountNetboxPrefixes() int32 {
	return as.statistics.countNetboxPrefixes.Load()
}

// SetCountNetboxPrefixesReceived количество полученных из Netbox префиксов
func (as *SharedAppStorage) SetCountNetboxPrefixesReceived(v int) {
	as.statistics.countNetboxPrefixesReceived.Store(int32(v))
}

// GetCountNetboxPrefixesReceived количество полученных из Netbox префиксов
func (as *SharedAppStorage) GetCountNetboxPrefixesReceived() int32 {
	return as.statistics.countNetboxPrefixesReceived.Load()
}

// SetCountNetboxPrefixesMatches количество префиксов Netbox с совпадениями ip адресов
func (as *SharedAppStorage) SetCountNetboxPrefixesMatches(v int) {
	as.statistics.countNetboxPrefixesMatches.Store(int32(v))
}

// GetCountNetboxPrefixesMatches количество префиксов Netbox с совпадениями ip адресов
func (as *SharedAppStorage) GetCountNetboxPrefixesMatches() int32 {
	return as.statistics.countNetboxPrefixesMatches.Load()
}

// SetCountUpdatedZabbixHosts количество обновленных хостов в Zabbix
func (as *SharedAppStorage) SetCountUpdatedZabbixHosts(v int) {
	as.statistics.countUpdatedZabbixHosts.Store(int32(v))
}

// GetCountUpdatedZabbixHosts количество обновленных хостов в Zabbix
func (as *SharedAppStorage) GetCountUpdatedZabbixHosts() int32 {
	return as.statistics.countUpdatedZabbixHosts.Load()
}

// DeleteElement удаляет заданный элемент по hostId
func (as *SharedAppStorage) DeleteElement(hostId int) {
	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	as.statistics.sort()
	if index := as.statistics.binarySearch(hostId); index != -1 {
		if index < 0 || index >= len(as.statistics.data) {
			return
		}

		as.statistics.data = append(as.statistics.data[:index], as.statistics.data[index+1:]...)
	}
}

// DeleteAll удаляет все элементы из хранилища, выставляет время начала и конца
// выполнения в дату начала эпохи Unix, выставляет статус выполнения в
// 'не выполняется' (false)
func (as *SharedAppStorage) DeleteAll() {
	as.statistics.countZabbixHostsGroup.Store(0)
	as.statistics.countZabbixHosts.Store(0)
	as.statistics.countMonitoringHostsGroup.Store(0)
	as.statistics.countMonitoringHosts.Store(0)
	as.statistics.countNetboxPrefixes.Store(0)
	as.statistics.countNetboxPrefixesReceived.Store(0)
	as.statistics.countUpdatedZabbixHosts.Store(0)

	as.statistics.mutex.Lock()
	defer as.statistics.mutex.Unlock()

	as.statistics.startDateExecution = time.Time{}
	as.statistics.endDateExecution = time.Time{}
	as.statistics.isExecution.Store(false)

	as.statistics.data = []HostDetailedInformation{}
}

// GetListErrors список объектов с ошибками
func (as *SharedAppStorage) GetListErrors() []HostDetailedInformation {
	as.statistics.mutex.RLock()
	defer as.statistics.mutex.RUnlock()

	list := make([]HostDetailedInformation, 0)
	for _, v := range as.statistics.data {
		if v.Error != nil {
			list = append(list, v)
		}
	}

	return list
}

/*
	Ips           []netip.Addr `json:"ips"`             // список ip адресов
	SensorsId     []string     `json:"sensor_id"`       // id обслуживающего сенсора
	NetboxHostsId []int        `json:"netbox_hosts_id"` // id хоста в netbox
	OriginalHost  string       `json:"original_host"`   // исходное наименование хоста
	DomainName    string       `json:"domain_name"`     // доменное имя
	Error         error        `json:"error"`           // ошибка
	HostId        int          `json:"host_id"`         // id хоста
	IsActive      bool         `json:"is_active"`       // флаг активный ли хост
	IsProcessed   bool         `json:"is_processed"`    // флаг обработан ли хост
*/

func (hdi *HostDetailedInformation) GetIps() []netip.Addr {
	return hdi.Ips
}

func (hdi *HostDetailedInformation) GetSensorsId() []string {
	return hdi.SensorsId
}

func (hdi *HostDetailedInformation) GetNetboxHostsId() []int {
	return hdi.NetboxHostsId
}

func (hdi *HostDetailedInformation) GetOriginalHost() string {
	return hdi.OriginalHost
}

func (hdi *HostDetailedInformation) GetDomainName() string {
	return hdi.DomainName
}

func (hdi *HostDetailedInformation) GetError() error {
	return hdi.Error
}

func (hdi *HostDetailedInformation) GetHostId() int {
	return hdi.HostId
}

func (hdi *HostDetailedInformation) GetIsActive() bool {
	return hdi.IsActive
}

func (hdi *HostDetailedInformation) GetIsProcessed() bool {
	return hdi.IsProcessed
}

// sort сортировка
func (sa *StatisticsApp) sort() {
	slices.SortFunc(sa.data, func(a, b HostDetailedInformation) int {
		return cmp.Compare(a.HostId, b.HostId)
	})
}

// binarySearch выполняет стандартный двоичный поиск.
// Возвращает индекс целевого объекта, если он найден, или -1, если не найден.
func (sa *StatisticsApp) binarySearch(id int) int {
	left, right := 0, len(sa.data)

	for left < right {
		mid := left + (right-left)/2
		if sa.data[mid].HostId == id {
			return mid
		} else if sa.data[mid].HostId < id {
			left = mid + 1
		} else {
			right = mid
		}
	}

	return -1
}

func (sa *StatisticsApp) search(nameElem, valueElem string) int {
	for k, v := range sa.data {
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

//********************* логи приложения ***********************

// AddLog добавить информацию по логам
func (as *SharedAppStorage) AddLog(v LogInformation) {
	as.logs.mutex.Lock()
	defer as.logs.mutex.Unlock()

	if len(as.logs.story) == as.logs.size {
		as.logs.story = append(as.logs.story[1:], v)

		return
	}

	as.logs.story = append(as.logs.story, v)
}

// GetLog получить информицию по логам
func (as *SharedAppStorage) GetLogs() []LogInformation {
	as.logs.mutex.RLock()
	defer as.logs.mutex.RUnlock()

	newList := make([]LogInformation, len(as.logs.story))
	copy(newList, as.logs.story)

	slices.Reverse(newList)

	return newList
}

// LogMaxSize максимальный размер логов
func (as *SharedAppStorage) LogMaxSize() int {
	return as.logs.size
}

//********************* настройки приложения ***********************

// GetTaskSchedulerDailyJobs рассписание времени выполнения задач
func (as *SharedAppStorage) GetTaskSchedulerDailyJobs() []string {
	return as.configuration.taskSchedulerDailyJobs
}

// SetTaskSchedulerDailyJobs рассписание времени выполнения задач
func (as *SharedAppStorage) SetTaskSchedulerDailyJobs(v []string) {
	as.configuration.taskSchedulerDailyJobs = v
}

// GetTaskSchedulerTimeJob интервал времени для выполнения задач
func (as *SharedAppStorage) GetTaskSchedulerTimeJob() int {
	return as.configuration.taskSchedulerTimerJob
}

// SetTaskSchedulerTimeJob интервал времени для выполнения задач
func (as *SharedAppStorage) SetTaskSchedulerTimeJob(v int) {
	as.configuration.taskSchedulerTimerJob = v
}

// GetNetbox настройки Netbox
func (as *SharedAppStorage) GetNetbox() ShortParameters {
	return as.configuration.netbox
}

// SetNetbox настройки Netbox
func (as *SharedAppStorage) SetNetbox(v ShortParameters) {
	as.configuration.netbox = v
}

// GetZabbix настройки Zabbix
func (as *SharedAppStorage) GetZabbix() ShortParameters {
	return as.configuration.zabbix
}

// SetZabbix настройки Zabbix
func (as *SharedAppStorage) SetZabbix(v ShortParameters) {
	as.configuration.zabbix = v
}

// GetDatabaseLogging настройки доступа к БД для логирования событий и ошибок
func (as *SharedAppStorage) GetDatabaseLogging() ShortParameters {
	return as.configuration.databaseLogging
}

// SetDatabaseLogging настройки доступа к БД для логирования событий и ошибок
func (as *SharedAppStorage) SetDatabaseLogging(v ShortParameters) {
	as.configuration.databaseLogging = v
}
