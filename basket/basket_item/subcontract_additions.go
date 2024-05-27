package basket_item

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"go.citilink.cloud/store_types"
	"sync"
	"time"
)

type SubcontractItemAdditions struct {
	// Доп. данные для оказания услуги
	//
	// задается самим пользователем, так что всегда необходимо проверять на nil
	ApplyServiceInfo *SubcontractApplyServiceInfo // 1
	mx               sync.RWMutex                 `msgpack:"-"`
}

func NewSubcontractItemAdditions(
	applyServiceInfo *SubcontractApplyServiceInfo,
) *SubcontractItemAdditions {
	return &SubcontractItemAdditions{
		ApplyServiceInfo: applyServiceInfo,
		mx:               sync.RWMutex{},
	}
}

func (s *SubcontractItemAdditions) GetApplyServiceInfo() *SubcontractApplyServiceInfo {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.ApplyServiceInfo
}

type SubcontractApplyServiceInfo struct {
	Date        time.Time           // 1
	Address     string              // 2
	CityKladrId store_types.KladrId // 3
	CityName    string              // 4
	mx          sync.RWMutex        `msgpack:"-"`
}

func NewSubcontractApplyServiceInfo(
	date time.Time,
	address string,
	cityKladrId store_types.KladrId,
	cityName string,
) *SubcontractApplyServiceInfo {
	return &SubcontractApplyServiceInfo{
		Date:        date,
		Address:     address,
		CityKladrId: cityKladrId,
		CityName:    cityName,
		mx:          sync.RWMutex{},
	}
}

func (s *SubcontractApplyServiceInfo) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Address, validation.Required),
		validation.Field(&s.CityName, validation.Required),
		validation.Field(&s.CityKladrId, validation.Required),
		validation.Field(&s.Date, validation.Required),
	)
}

func (s *SubcontractApplyServiceInfo) GetDate() time.Time {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.Date
}

func (s *SubcontractApplyServiceInfo) SetDate(date time.Time) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.Date = date
}

func (s *SubcontractApplyServiceInfo) GetAddress() string {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.Address
}

func (s *SubcontractApplyServiceInfo) SetAddress(address string) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.Address = address
}

func (s *SubcontractApplyServiceInfo) GetCityKladrId() store_types.KladrId {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.CityKladrId
}

func (s *SubcontractApplyServiceInfo) SetCityKladrId(kladrId store_types.KladrId) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.CityKladrId = kladrId
}

func (s *SubcontractApplyServiceInfo) GetCityName() string {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.CityName
}

func (s *SubcontractApplyServiceInfo) SetCityName(city string) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.CityName = city
}
