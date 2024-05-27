package basket_item

import "fmt"

// Type Тип позиции
type Type string

const (
	TypeUnknown                      Type = "unknown"
	TypeProduct                      Type = "product"
	TypeSubcontractServiceForProduct Type = "subcontract_service_for_product"
	TypeInsuranceServiceForProduct   Type = "insurance_service_for_product"
	TypeDigitalService               Type = "digital_service"
	TypePresent                      Type = "present"
	TypePropertyInsurance            Type = "property_insurance"
	TypeConfiguration                Type = "configuration"
	TypeConfigurationProduct         Type = "product_in_configuration"
	TypeConfigurationProductService  Type = "service_for_product_in_configuration"
	TypeConfigurationAssemblyService Type = "assembly_service_in_configurations"
	TypeDeliveryService              Type = "DeliveryService"
	TypeLiftingService               Type = "lifting_service"
)

var typeOptions = map[Type]struct {
	serviceType ServiceType
	name        string
	navType     NavType
	// Данный тип позиций добавляется сервером. Эта опция нужна для того, чтобы определять "пустоту корзины", так как
	// обычно такие позиции нельзя удалять и ничего с ними сделать и может оказаться так, что пользователь удалил все
	// свои позиции из заказа, но у него осталась, к примеру, услуга доставки...
	isAddedByServer bool
	isProduct       bool
	isService       bool
	isConfiguration bool
	// Являются ли частью конфигурации или самое конфигурацией
	isPartOfConfiguration bool
	isPresent             bool
}{
	TypeProduct: {
		isProduct: true,
		name:      "товар",
		navType:   NavTypeProduct,
	},
	TypeSubcontractServiceForProduct: {
		isService:   true,
		serviceType: ServiceTypeSubcontract,
		name:        "услуга установки",
		navType:     NavTypeService,
	},
	TypeInsuranceServiceForProduct: {
		isService:   true,
		serviceType: ServiceTypeProductInsurance,
		name:        "защита покупки",
		navType:     NavTypeService,
	},
	TypeDigitalService: {
		isService:   true,
		serviceType: ServiceTypeDigital,
		name:        "цифровая услуга",
		navType:     NavTypeService,
	},
	TypePresent: {
		isPresent:       true,
		name:            "подарок",
		navType:         NavTypeProduct,
		isAddedByServer: true,
	},
	TypePropertyInsurance: {
		isService:   true,
		serviceType: ServiceTypeInsuranceOfProperty,
		name:        "услуга страхования имущества",
		navType:     NavTypeService,
	},
	TypeConfiguration: {
		isConfiguration: true,
		name:            "сборка компьютера",
		navType:         NavTypeProduct,
	},
	TypeConfigurationProduct: {
		isProduct:             true,
		name:                  "товар",
		navType:               NavTypeProduct,
		isPartOfConfiguration: true,
	},
	TypeConfigurationProductService: {
		isService:             true,
		serviceType:           ServiceTypeForConfigurationFeature,
		name:                  "услуга для комплектующих сборки",
		navType:               NavTypeService,
		isPartOfConfiguration: true,
	},
	TypeConfigurationAssemblyService: {
		isService: true,
		navType:   NavTypeProduct,
		// странно, что тут нет сервис типа. Типа если это конфигурация, и это услуга и нет типа услуги, то это услуга сборки
		name:                  "услуга сборки компьютера",
		isPartOfConfiguration: true,
	},
	TypeDeliveryService: {
		isService:       true,
		name:            "услуга доставки",
		navType:         NavTypeService,
		serviceType:     ServiceTypeDelivery,
		isAddedByServer: true,
	},
	TypeLiftingService: {
		isService:       true,
		name:            "услуга подъема на этаж",
		navType:         NavTypeService,
		isAddedByServer: true,
	},
}

func (t Type) Validate() error {
	_, ok := typeOptions[t]
	if !ok {
		return fmt.Errorf("type is not in valid state %s", t)
	}

	return nil
}

func (t Type) Spec() *Spec {
	if spec, ok := typeToSpecs[t]; ok {
		return spec
	}

	panic(fmt.Sprintf("type %s not have spec", t))
}

func (t Type) IsProduct() bool {
	return typeOptions[t].isProduct
}

func (t Type) IsAddedByServer() bool {
	return typeOptions[t].isAddedByServer
}

func (t Type) IsService() bool {
	return typeOptions[t].isService
}

func (t Type) IsPresent() bool {
	return typeOptions[t].isPresent
}

func (t Type) IsConfiguration() bool {
	return typeOptions[t].isConfiguration
}

func (t Type) IsPartOfConfiguration() bool {
	return typeOptions[t].isPartOfConfiguration
}

func (t Type) Name() string {
	return typeOptions[t].name
}

// ServiceType возвраащет служебный тип услуги, необходимый для БД
func (t Type) ServiceType() ServiceType {
	return typeOptions[t].serviceType
}

// NavType возвращает служебный тип позиции для NAV
func (t Type) NavType() NavType {
	return typeOptions[t].navType
}
