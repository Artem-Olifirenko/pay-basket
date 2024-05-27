package basket_item

import "sync"

type InfoId int

const (
	InfoIdUnknown InfoId = iota
	// InfoIdPriceChanged цена на позицию изменилась
	InfoIdPriceChanged
	// InfoIdCountMoreThanAvail кол-во товара в наличии меньше, чем запрошенное кол-во позиций в корзине.
	InfoIdCountMoreThanAvail
	// InfoIdPositionRemoved позиция удалена из корзины
	InfoIdPositionRemoved
	// InfoIdPositionChanged позиция заменена
	InfoIdPositionChanged
)

func NewInfo(id InfoId, message string) *Info {
	return &Info{id: id, message: message, additions: &InfoAdditions{}}
}

// Info информация о позиции корзины
type Info struct {
	id        InfoId         // 1
	message   string         // 2
	additions *InfoAdditions // 3
	mx        sync.RWMutex   `msgpack:"-"`
}

func (i *Info) Id() InfoId {
	i.mx.RLock()
	defer i.mx.RUnlock()

	return i.id
}

func (i *Info) Message() string {
	i.mx.RLock()
	defer i.mx.RUnlock()

	return i.message
}

func (i *Info) Additionals() *InfoAdditions {
	i.mx.RLock()
	defer i.mx.RUnlock()

	return i.additions
}

func (i *Info) SetAdditions(additions *InfoAdditions) {
	i.mx.Lock()
	defer i.mx.Unlock()

	i.additions = additions
}

// InfoAdditions уточняющая информация по позиции
type InfoAdditions struct {
	PriceChanged       PriceChangedInfoAddition        // 1
	CountMoreThenAvail CountMoreThenAvailInfoAdditions // 2
	ChangedItem        ChangedItemInfoAdditions        // 3
}

// PriceChangedInfoAddition информация об изменении цены
type PriceChangedInfoAddition struct {
	From int // 1
	To   int // 2
}

// CountMoreThenAvailInfoAdditions информация о том, что по позиции доступно товаров меньше, чем добавлено в корзину
type CountMoreThenAvailInfoAdditions struct {
	AvailCount int // 1
}

// ChangedItemInfoAdditions Информация о замененном товаре/услуге.
// Указывается для идентификатора INFO_ID_POSITION_CHANGED
type ChangedItemInfoAdditions struct {
	// Идентификатор позиции (не уникален)
	ItemId string // 1
	// Уникальный идентификатор позиции
	UniqId string // 2
	// Кол-во позиции (например 2 телефона)
	Count int // 3
	// Название позиции
	Name string // 4
	// Цена за позицию
	Price int // 5
}
