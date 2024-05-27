package basket_item

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type BasketItemTypeSuite struct {
	suite.Suite
}

func (s *BasketItemTypeSuite) TestType_IsAddedByServer() {
	tests := []struct {
		name string
		typ  Type
		want func() (isAddedByServer bool)
	}{
		{
			"type product",
			TypeProduct,
			func() bool {
				return false
			},
		},
		{
			"subcontract service for product",
			TypeSubcontractServiceForProduct,
			func() bool {
				return false
			},
		},
		{
			"insurance service for product",
			TypeInsuranceServiceForProduct,
			func() bool {
				return false
			},
		},
		{
			"digital service",
			TypeDigitalService,
			func() bool {
				return false
			},
		},
		{
			"present",
			TypePresent,
			func() bool {
				return true
			},
		},
		{
			"property insurance",
			TypePropertyInsurance,
			func() bool {
				return false
			},
		},
		{
			"configuration",
			TypeConfiguration,
			func() bool {
				return false
			},
		},
		{
			"configuration product",
			TypeConfigurationProduct,
			func() bool {
				return false
			},
		},
		{
			"configuration product service",
			TypeConfigurationProductService,
			func() bool {
				return false
			},
		},
		{
			"configuration assembly service",
			TypeConfigurationAssemblyService,
			func() bool {
				return false
			},
		},
		{
			"delivery service",
			TypeDeliveryService,
			func() bool {
				return true
			},
		},
		{
			"lifting service",
			TypeLiftingService,
			func() bool {
				return true
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			s.Equal(tt.want(), tt.typ.IsAddedByServer())
		})
	}
}

func (s *BasketItemTypeSuite) TestType_IsService() {
	tests := []struct {
		name string
		typ  Type
		want func() (isService bool)
	}{
		{
			"type product",
			TypeProduct,
			func() bool {
				return false
			},
		},
		{
			"subcontract service for product",
			TypeSubcontractServiceForProduct,
			func() bool {
				return true
			},
		},
		{
			"insurance service for product",
			TypeInsuranceServiceForProduct,
			func() bool {
				return true
			},
		},
		{
			"digital service",
			TypeDigitalService,
			func() bool {
				return true
			},
		},
		{
			"present",
			TypePresent,
			func() bool {
				return false
			},
		},
		{
			"property insurance",
			TypePropertyInsurance,
			func() bool {
				return true
			},
		},
		{
			"configuration",
			TypeConfiguration,
			func() bool {
				return false
			},
		},
		{
			"configuration product",
			TypeConfigurationProduct,
			func() bool {
				return false
			},
		},
		{
			"configuration product service",
			TypeConfigurationProductService,
			func() bool {
				return true
			},
		},
		{
			"configuration assembly service",
			TypeConfigurationAssemblyService,
			func() bool {
				return true
			},
		},
		{
			"delivery service",
			TypeDeliveryService,
			func() bool {
				return true
			},
		},
		{
			"lifting service",
			TypeLiftingService,
			func() bool {
				return true
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			s.Equal(tt.want(), tt.typ.IsService())
		})
	}
}

func (s *BasketItemTypeSuite) TestType_IsConfiguration() {
	tests := []struct {
		name string
		typ  Type
		want func() bool
	}{
		{
			"type product",
			TypeProduct,
			func() bool {
				return false
			},
		},
		{
			"subcontract service for product",
			TypeSubcontractServiceForProduct,
			func() bool {
				return false
			},
		},
		{
			"insurance service for product",
			TypeInsuranceServiceForProduct,
			func() bool {
				return false
			},
		},
		{
			"digital service",
			TypeDigitalService,
			func() bool {
				return false
			},
		},
		{
			"present",
			TypePresent,
			func() bool {
				return false
			},
		},
		{
			"property insurance",
			TypePropertyInsurance,
			func() bool {
				return false
			},
		},
		{
			"configuration",
			TypeConfiguration,
			func() bool {
				return true
			},
		},
		{
			"configuration product",
			TypeConfigurationProduct,
			func() bool {
				return false
			},
		},
		{
			"configuration product service",
			TypeConfigurationProductService,
			func() bool {
				return false
			},
		},
		{
			"configuration assembly service",
			TypeConfigurationAssemblyService,
			func() bool {
				return false
			},
		},
		{
			"delivery service",
			TypeDeliveryService,
			func() bool {
				return false
			},
		},
		{
			"lifting service",
			TypeLiftingService,
			func() bool {
				return false
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			s.Equal(tt.want(), tt.typ.IsConfiguration())
		})
	}
}

func (s *BasketItemTypeSuite) TestType_IsPartOfConfiguration() {
	tests := []struct {
		name string
		typ  Type
		want func() bool
	}{
		{
			"type product",
			TypeProduct,
			func() bool {
				return false
			},
		},
		{
			"subcontract service for product",
			TypeSubcontractServiceForProduct,
			func() bool {
				return false
			},
		},
		{
			"insurance service for product",
			TypeInsuranceServiceForProduct,
			func() bool {
				return false
			},
		},
		{
			"digital service",
			TypeDigitalService,
			func() bool {
				return false
			},
		},
		{
			"present",
			TypePresent,
			func() bool {
				return false
			},
		},
		{
			"property insurance",
			TypePropertyInsurance,
			func() bool {
				return false
			},
		},
		{
			"configuration",
			TypeConfiguration,
			func() bool {
				return false
			},
		},
		{
			"configuration product",
			TypeConfigurationProduct,
			func() bool {
				return true
			},
		},
		{
			"configuration product service",
			TypeConfigurationProductService,
			func() bool {
				return true
			},
		},
		{
			"configuration assembly service",
			TypeConfigurationAssemblyService,
			func() bool {
				return true
			},
		},
		{
			"delivery service",
			TypeDeliveryService,
			func() bool {
				return false
			},
		},
		{
			"lifting service",
			TypeLiftingService,
			func() bool {
				return false
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			s.Equal(tt.want(), tt.typ.IsPartOfConfiguration())
		})
	}
}

func (s *BasketItemTypeSuite) TestType_ServiceType() {
	tests := []struct {
		name string
		typ  Type
		want func() ServiceType
	}{
		{
			"type product",
			TypeProduct,
			func() ServiceType {
				return ServiceType(0)
			},
		},
		{
			"subcontract service for product",
			TypeSubcontractServiceForProduct,
			func() ServiceType {
				return ServiceTypeSubcontract
			},
		},
		{
			"insurance service for product",
			TypeInsuranceServiceForProduct,
			func() ServiceType {
				return ServiceTypeProductInsurance
			},
		},
		{
			"digital service",
			TypeDigitalService,
			func() ServiceType {
				return ServiceTypeDigital
			},
		},
		{
			"present",
			TypePresent,
			func() ServiceType {
				return ServiceType(0)
			},
		},
		{
			"property insurance",
			TypePropertyInsurance,
			func() ServiceType {
				return ServiceTypeInsuranceOfProperty
			},
		},
		{
			"configuration",
			TypeConfiguration,
			func() ServiceType {
				return ServiceType(0)
			},
		},
		{
			"configuration product",
			TypeConfigurationProduct,
			func() ServiceType {
				return ServiceType(0)
			},
		},
		{
			"configuration product service",
			TypeConfigurationProductService,
			func() ServiceType {
				return ServiceTypeForConfigurationFeature
			},
		},
		{
			"configuration assembly service",
			TypeConfigurationAssemblyService,
			func() ServiceType {
				return ServiceType(0)
			},
		},
		{
			"delivery service",
			TypeDeliveryService,
			func() ServiceType {
				return ServiceTypeDelivery
			},
		},
		{
			"lifting service",
			TypeLiftingService,
			func() ServiceType {
				return ServiceType(0)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			s.Equal(tt.want(), tt.typ.ServiceType())
		})
	}
}

func (s *BasketItemTypeSuite) TestType_Name() {
	tests := []struct {
		name string
		typ  Type
		want func() string
	}{
		{
			name: "product",
			typ:  TypeProduct,
			want: func() string {
				return "товар"
			},
		},
		{
			name: "subcontract service for product",
			typ:  TypeSubcontractServiceForProduct,
			want: func() string {
				return "услуга установки"
			},
		},
		{
			name: "insurance service for product",
			typ:  TypeInsuranceServiceForProduct,
			want: func() string {
				return "защита покупки"
			},
		},
		{
			name: "digital service",
			typ:  TypeDigitalService,
			want: func() string {
				return "цифровая услуга"
			},
		},
		{
			name: "present",
			typ:  TypePresent,
			want: func() string {
				return "подарок"
			},
		},
		{
			name: "property insurance",
			typ:  TypePropertyInsurance,
			want: func() string {
				return "услуга страхования имущества"
			},
		},
		{
			name: "configuration",
			typ:  TypeConfiguration,
			want: func() string {
				return "сборка компьютера"
			},
		},
		{
			name: "configuration product",
			typ:  TypeConfigurationProduct,
			want: func() string {
				return "товар"
			},
		},
		{
			name: "configuration product service",
			typ:  TypeConfigurationProductService,
			want: func() string {
				return "услуга для комплектующих сборки"
			},
		},
		{
			name: "configuration assembly service",
			typ:  TypeConfigurationAssemblyService,
			want: func() string {
				return "услуга сборки компьютера"
			},
		},
		{
			name: "delivery service",
			typ:  TypeDeliveryService,
			want: func() string {
				return "услуга доставки"
			},
		},
		{
			name: "lifting service",
			typ:  TypeLiftingService,
			want: func() string {
				return "услуга подъема на этаж"
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			s.Equal(tt.want(), tt.typ.Name())
		})
	}
}

func TestType_NavType(t *testing.T) {
	tests := []struct {
		name string
		typ  Type
		want func() NavType
	}{
		{
			name: "type product",
			typ:  TypeProduct,
			want: func() NavType {
				return NavTypeProduct
			},
		},
		{
			name: "subcontract service for product",
			typ:  TypeSubcontractServiceForProduct,
			want: func() NavType {
				return NavTypeService
			},
		},
		{
			name: "insurance service for product",
			typ:  TypeInsuranceServiceForProduct,
			want: func() NavType {
				return NavTypeService
			},
		},
		{
			name: "digital service",
			typ:  TypeDigitalService,
			want: func() NavType {
				return NavTypeService
			},
		},
		{
			name: "present",
			typ:  TypePresent,
			want: func() NavType {
				return NavTypeProduct
			},
		},
		{
			name: "property insurance",
			typ:  TypePropertyInsurance,
			want: func() NavType {
				return NavTypeService
			},
		},
		{
			name: "configuration",
			typ:  TypeConfiguration,
			want: func() NavType {
				return NavTypeProduct
			},
		},
		{
			name: "configuration product",
			typ:  TypeConfigurationProduct,
			want: func() NavType {
				return NavTypeProduct
			},
		},
		{
			name: "configuration product service",
			typ:  TypeConfigurationProductService,
			want: func() NavType {
				return NavTypeService
			},
		},
		{
			name: "configuration assembly service",
			typ:  TypeConfigurationAssemblyService,
			want: func() NavType {
				return NavTypeProduct
			},
		},
		{
			name: "delivery service",
			typ:  TypeDeliveryService,
			want: func() NavType {
				return NavTypeService
			},
		},
		{
			name: "lifting service",
			typ:  TypeLiftingService,
			want: func() NavType {
				return NavTypeService
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want(), tt.typ.NavType())
		})
	}
}

func TestBasketItemTypeSuite(t *testing.T) {
	suite.Run(t, new(BasketItemTypeSuite))
}
