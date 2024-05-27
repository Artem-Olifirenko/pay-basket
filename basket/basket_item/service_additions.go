package basket_item

import (
	"sync"
)

type Service struct {
	IsCreditAvail              bool         // 1
	IsAvailableForInstallments bool         // 2
	mx                         sync.RWMutex `msgpack:"-"`
}

func NewService(
	isCreditAvail bool,
	isAvailableForInstallments bool,
) *Service {
	return &Service{IsCreditAvail: isCreditAvail, IsAvailableForInstallments: isAvailableForInstallments, mx: sync.RWMutex{}}
}

func (s *Service) GetIsCreditAvail() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.IsCreditAvail
}

func (s *Service) GetIsAvailableForInstallments() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.IsAvailableForInstallments
}

func (s *Service) SetIsCreditAvail(v bool) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.IsCreditAvail = v
}

func (s *Service) SetIsAvailableForInstallments(v bool) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.IsAvailableForInstallments = v
}
