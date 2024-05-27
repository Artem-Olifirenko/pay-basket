package basket_item

import "fmt"

// AllowResale структура с информацией о возможности приобретать позицию для перепродажи
type AllowResale struct {
	// Название товарной группы
	commodityGroupName string
	// Признак, можно ли приобретать данную позицию с целью перепродажи
	isAllow bool
}

func NewAllowResale(isAllow bool, commodityGroupName string) *AllowResale {
	return &AllowResale{
		isAllow:            isAllow,
		commodityGroupName: commodityGroupName,
	}
}

func (a *AllowResale) IsAllow() bool {
	return a.isAllow
}

// Message вспомогательное сообщение в случае невозможности перепродажи
func (a *AllowResale) Message() string {
	// Если перепродажа разрешена, то не нужно возвращать сообщение с дополнительной информацией
	if a.isAllow {
		return ""
	}

	// Если известно название запрещенной товарной группы, то добавляем его в сообщение
	if a.commodityGroupName != "" {
		return fmt.Sprintf(
			"Товарная группа %s не подключена. Товар не может быть куплен для цели покупки «Перепродажа». Добавьте товарную группу в системе «Честный знак»",
			a.commodityGroupName,
		)
	}

	return "Товарная группа не подключена. Товар не может быть куплен для цели покупки «Перепродажа». Добавьте товарную группу в системе «Честный знак»"
}
