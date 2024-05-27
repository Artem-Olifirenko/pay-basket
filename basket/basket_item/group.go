package basket_item

import "fmt"

// Group Группа позиции
type Group string

const (
	GroupInvalid       Group = "invalid"       // Некорретный тип
	GroupProduct       Group = "product"       // Товар
	GroupService       Group = "service"       // Услуга
	GroupConfiguration Group = "configuration" // Контейнер конфигурации
)

func (g Group) Validate() error {
	if g == GroupInvalid {
		return fmt.Errorf("group is not in valid state %s", g)
	}

	return nil
}
