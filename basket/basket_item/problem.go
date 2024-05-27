package basket_item

type ProblemId int

const (
	ProblemUnknown ProblemId = 0
	// ProblemNotAvailable позиции нет в наличии или невозможно купить
	ProblemNotAvailable ProblemId = 1
	// ProblemMaxCountExcess превышено максимальное кол-во позиции
	ProblemMaxCountExcess ProblemId = 2
	// ProblemProductItemInConfigurationNotAvailable одна из позиций в составе конфигурации не в наличии или
	// ее невозможно купить
	ProblemProductItemInConfigurationNotAvailable ProblemId = 3
	// ProblemNotAvailableInSelectedCity позиция недоступна в выбранном городе
	ProblemNotAvailableInSelectedCity ProblemId = 4
	// ProblemPurchaseReasonIsNotSelected не выбран тип маркированного товара
	ProblemPurchaseReasonIsNotSelected ProblemId = 5
	// ProblemPurchaseReasonNotAvailableForUser тип маркировки недоступен для пользователя
	ProblemPurchaseReasonNotAvailableForUser ProblemId = 6
	// ProblemFnsTrackedItemNotAvailableForUser отслеживаемый товар недоступен для пользователя
	ProblemFnsTrackedItemNotAvailableForUser ProblemId = 7
)

func NewProblem(id ProblemId, message string) *Problem {
	return &Problem{id: id, message: message, isHidden: false}
}

type Problem struct {
	id      ProblemId // 1
	message string    // 2
	// Дополнительные данные по проблеме
	additions ProblemAdditions // 3
	// Является ли данная позиция скрытой
	isHidden bool // 4
}

func (p *Problem) Id() ProblemId {
	return p.id
}

func (p *Problem) SetIsHidden(v bool) {
	p.isHidden = v
}

func (p *Problem) IsHidden() bool {
	return p.isHidden
}

func (p *Problem) Message() string {
	return p.message
}

// Additions возвращает дополнительные данные по проблеме
func (p *Problem) Additions() *ProblemAdditions {
	return &p.additions
}

// ProblemAdditions дополнительные данные по проблеме
type ProblemAdditions struct {
	ConfigurationProblemAdditions ConfigurationProblemAdditions // 1
}

// ConfigurationProblemAdditions дополнительные данные по проблемам с конфигурацией
type ConfigurationProblemAdditions struct {
	// Идентификаторы позиций, которые не в наличии или невозможно купить.
	//
	// В связи с уникальным поведением самой конфигурации мы не можем указывать проблемы о не наличии/невозможности
	// купить к самим позициям внутри конфигурации, так как их нельзя удалять и получится так, что проблема нерешаемая.
	// Проблему решить возможно только удалением/разборкой самой конфигурации
	//
	// Эти данные проставляются вместе с проблемой ProblemProductItemInConfigurationNotAvailable
	NotAvailableProductItemIds []ItemId // 1
}
