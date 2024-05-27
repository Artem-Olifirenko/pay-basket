package basket_item

import (
	servicev1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/service/v1"
)

// CatalogServicesToDomainMapper Маппер response mic catalog к доменным моделям
type CatalogServicesToDomainMapper struct {
}

func NewCatalogServicesToDomainMapper() *CatalogServicesToDomainMapper {
	return &CatalogServicesToDomainMapper{}
}

func (m *CatalogServicesToDomainMapper) MapCatalogServicesTypeToBasketItemType(services []*servicev1.ServiceInfo) Type {
	itemType := TypeUnknown
	if len(services) == 0 {
		return itemType
	}

	serviceType := services[0].GetServiceType()
	serviceSubType := services[0].GetSubtype()

	switch serviceType {
	case servicev1.ServiceType_SERVICE_TYPE_SUBCONTRACT: // Услуга "установки"
		itemType = TypeSubcontractServiceForProduct
	case servicev1.ServiceType_SERVICE_TYPE_INSURANCE_POLICE: // Услуга страхования товара
		itemType = TypeInsuranceServiceForProduct
	case servicev1.ServiceType_SERVICE_TYPE_INSURANCE_PRODUCT: // Страхование имущества
		itemType = TypePropertyInsurance
	}

	switch serviceSubType {
	case servicev1.Subtype_SUBTYPE_SOFTWARE_INSTALL: // Услуга цифровая
		itemType = TypeDigitalService
	case servicev1.Subtype_SUBTYPE_ASSEMBLY: // Услуга сборки для конфигурации
		itemType = TypeConfigurationAssemblyService
	case servicev1.Subtype_SUBTYPE_DELIVERY: // Услуга доставки
		itemType = TypeDeliveryService
	case servicev1.Subtype_SUBTYPE_RISE: // Услуга подъема на этаж
		itemType = TypeLiftingService
	}

	return itemType
}
