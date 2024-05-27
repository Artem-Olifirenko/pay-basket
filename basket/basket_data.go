package basket

import (
	"encoding/binary"
	"fmt"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/store_types"
	"hash/fnv"
	"sort"
	"strconv"
)

type CityId string

type BasketData struct {
	// Идентификатор региона, относительно которого рассчитывается наличие и цены в корзине
	spaceId store_types.SpaceId // 1
	items   basket_item.ItemMap // 2
	// 3 deleted
	// 4 deleted
	// ценовая колонка относительно которой рассчитываются цены в корзине
	priceColumn       catalog_types.PriceColumn // 5
	infos             []*Info                   // 6
	commitFingerprint string                    // 7
	cityId            CityId                    // 8
	// Флаг показывающий можно ли из товаров корзины собрать конфигурацию
	hasPossibleConfiguration bool // 9

}

// NewBasketData создает данные корзины
//
// Будьте внимательны, при добавлении новых свойств не забывайте также мапить их в методе копирования
// BasketData.Copy(), так как там тоже производится создание корзины
func NewBasketData(spaceId store_types.SpaceId, priceColumn catalog_types.PriceColumn, cityId CityId) *BasketData {
	return &BasketData{
		items:       make(map[basket_item.UniqId]*basket_item.Item),
		spaceId:     spaceId,
		priceColumn: priceColumn,
		cityId:      cityId,
	}
}

// Fingerprint собирает информацию по всей корзине и выводит это в виде хэша. Данный хэш при сборе так же
// сортирует позиции заказа по идентификатору позиции, таким образом увеличивается кол-во одинаковых отпечатков у
// одинаковых корзин
func (b *BasketData) Fingerprint() string {
	// данная реализация никогда не сможет вернуть ошибку, поэтому они здесь опущены + так же это крайне затрудняет
	// тестирование, так как для подмены хэша придется городить какую-то фабрику или любую другую решение, которое
	// только усложнит код.
	h := fnv.New64a()
	_, _ = h.Write([]byte(b.SpaceId()))
	_ = binary.Write(h, binary.LittleEndian, int32(b.PriceColumn()))

	hashes := make([]string, 0)
	for _, item := range b.items {
		hashes = append(hashes, item.Fingerprint())
	}

	sort.Strings(hashes)
	for _, i := range hashes {
		_, _ = h.Write([]byte(i))
	}

	return strconv.FormatUint(h.Sum64(), 10)
}

// Add добавляет позицию в корзину и производит набор проверок на основе спецификаций типа и правил самой позиции.
//
// Одним из возвращаемых значений является позиция, в большинстве случаев она является той же позицией, что и
// добавляемая, НО, если производится добавление уже существующей позиции (с одним и тем же идентификатором позиции и
// родительской позицией), то вернется уже существующая позиция, а не добавляемая.
func (b *BasketData) Add(item *basket_item.Item) (*basket_item.Item, error) {
	for _, existItem := range b.items {
		if existItem.ItemId() == item.ItemId() && existItem.ParentUniqId() == item.ParentUniqId() {
			newCount := existItem.Count() + item.Count()
			if existItem.Rules().MaxCount() > 0 && newCount > existItem.Rules().MaxCount() {
				newCount = existItem.Rules().MaxCount()
			}

			if existItem.Spec().IsCountChangeable() {
				err := existItem.SetCount(newCount)
				if err != nil {
					return nil, fmt.Errorf("can't change count: %w", err)
				}
			}

			return existItem, nil
		}
	}

	var parentItem *basket_item.Item
	if item.IsChild() {
		parentItem = b.FindOneById(item.ParentUniqId())
		if parentItem == nil {
			return nil, fmt.Errorf("can't find parent item of %s by id %s", item.UniqId(), item.ParentUniqId())
		}
	}

	if parentItem != nil && item.Spec().IsCountLessOrEqualThenParent() {
		if item.Count() > parentItem.Count() {
			err := item.SetCount(parentItem.Count())
			if err != nil {
				return nil, fmt.Errorf("can't set item count : %w", err)
			}
		}

		item.Rules().SetMaxCount(parentItem.Count())
	}

	// если такой тип позиции может быть только один у родителя, тогда удалим уже существующую позицию
	if parentItem != nil && item.Spec().IsOnlyOnePositionPerParent() {
		for _, foundedItem := range b.All() {
			if foundedItem.Type() == item.Type() && foundedItem.ParentUniqId() == parentItem.UniqId() {
				b.Remove(foundedItem)
			}
		}
	}

	// если родительская позиция не выбрана для выкупа - отмечаем ее выбранной обязательно, как и все ее дочерние
	if parentItem != nil && !parentItem.IsSelected() {
		parentItem.SetIsSelected(true)
		for _, bItem := range b.items {
			if bItem.ParentUniqId() == parentItem.UniqId() {
				bItem.SetIsSelected(true)
			}
		}
	}

	if item.Spec().IsOnlyOnePositionPossible() {
		// в случае, если добавляется еще одна позиция, которая может быть только одна в корзине, то предыдущая
		// позиция с этим типом удаляется из корзины и таким образом мы ее как будто"заменяем"
		itemsOfSameType := b.Find(Finders.ByType(item.Type()))
		for _, foundedItem := range itemsOfSameType {
			b.Remove(foundedItem)
		}
	}

	b.items[item.UniqId()] = item

	return item, nil
}

func (b *BasketData) FindOneById(id basket_item.UniqId) *basket_item.Item {
	item, ok := b.items[id]
	if !ok {
		return nil
	}

	return item
}

func (b *BasketData) FindByIds(ids ...basket_item.UniqId) []*basket_item.Item {
	var foundedItems []*basket_item.Item
	for _, id := range ids {
		if item, ok := b.items[id]; ok {
			foundedItems = append(foundedItems, item)
		}
	}

	return foundedItems
}

func (b *BasketData) Find(finder Finder) basket_item.Items {
	return finder(b.All())
}

func (b *BasketData) FindSelected(finder Finder) basket_item.Items {
	return finder(b.SelectedItems())
}

func (b *BasketData) Remove(item *basket_item.Item) {
	childrenRecursive := b.Find(Finders.ChildrenOfRecursive(item))
	for _, child := range childrenRecursive {
		delete(b.items, child.UniqId())
	}

	delete(b.items, item.UniqId())
}

func (b *BasketData) All() basket_item.Items {
	return b.items.ToSlice()
}

func (b *BasketData) SelectedItems() basket_item.Items {
	return b.items.ToSliceOnlySelected()
}

func (b *BasketData) CommitChanges() {
	b.commitFingerprint = b.Fingerprint()
	for _, item := range b.items {
		item.CommitChanges()
	}
}

// HasPrepaymentMandatoryItems проверяет, есть ли в корзине товары, покупка которых возможна только по предоплате
func (b *BasketData) HasPrepaymentMandatoryItems() bool {
	for _, item := range b.SelectedItems() {
		if item.IsPrepaymentMandatory() {
			return true
		}
	}
	return false
}

// HasConfiguration проверяет, есть ли в корзине конфигурации
func (b *BasketData) HasConfiguration() bool {
	for _, item := range b.SelectedItems() {
		if item.Type().IsConfiguration() {
			return true
		}
	}
	return false
}

func (b *BasketData) IsChanged() bool {
	if b.commitFingerprint != b.Fingerprint() {
		return true
	}

	for _, item := range b.items {
		if item.IsChanged() {
			return true
		}
	}

	return false
}

func (b *BasketData) SetHasPossibleConfiguration(v bool) {
	b.hasPossibleConfiguration = v
}

func (b *BasketData) HasPossibleConfiguration() bool {
	return b.hasPossibleConfiguration
}

func (b *BasketData) AvailableConfiguration() bool {
	hasConf := false
	for _, item := range b.All() {
		if !item.Type().IsProduct() {
			continue
		}
		if item.Type().IsPartOfConfiguration() {
			hasConf = true
			break
		}
	}

	return hasConf
}

func (b *BasketData) SpaceId() store_types.SpaceId {
	return b.spaceId
}

func (b *BasketData) CityId() CityId {
	return b.cityId
}

func (b *BasketData) Clear() {
	b.items = make(map[basket_item.UniqId]*basket_item.Item)
}

func (b *BasketData) Cost() int {
	var cost = 0
	var isAvailable bool
	// Стоимость рассчитываем только для selected позиций
	for _, item := range b.SelectedItems() {
		// так как сама конфигурация обладает стоимостью, то имеет смысл только ее и считать, а позиции в составе
		// конфигурации нужно пропускать, иначе итоговая сумма будет всегда x2
		if item.Type().IsPartOfConfiguration() {
			continue
		}

		isAvailable = true
		for _, problem := range item.Problems() {
			if problem.Id() == basket_item.ProblemNotAvailable {
				isAvailable = false
				break
			}
		}

		if !isAvailable {
			continue
		}

		cost += item.Count() * item.Price()
	}

	return cost
}

func (b *BasketData) Count() int {
	return len(b.items)
}

func (b *BasketData) CountSelected() int {
	return len(b.SelectedItems())
}

func (b *BasketData) ToXItems() []*basket_item.XItem {
	xitems := make([]*basket_item.XItem, 0, len(b.items))
	for _, item := range b.SelectedItems() {
		xitems = append(xitems, item.ToXItem())
	}

	return xitems
}

func (b *BasketData) Problems() []*Problem {
	var problems []*Problem
	for _, item := range b.items {
		for _, itemProblem := range item.Problems() {
			problems = append(problems, NewProblem(item, itemProblem))
		}
	}

	return problems
}

func (b *BasketData) Infos() []*Info {
	var infos []*Info
	infos = b.infos
	for _, item := range b.items {
		for _, itemInfo := range item.Infos() {
			infos = append(infos, NewInfo(item, itemInfo))
		}
	}

	return infos
}

func (b *BasketData) AccruedBonus() int {
	bonus := 0
	// Начисляемые бонусы рассчитываем только для selected позиций
	for _, item := range b.SelectedItems() {
		// Пропускаем позиции, являющиеся комплектующими для конфигурации, так как конфигурация агрегирует в
		// себе кол-во бонусов, которые будут получены за выкуп заказа.
		if item.Type().IsPartOfConfiguration() {
			continue
		}
		bonus += item.Bonus() * item.Count()
	}

	return bonus
}

// IsAllProductsInStore узнает все ли позиции с типом "товар" есть в наличии в магазинах
func (b *BasketData) IsAllProductsInStore() bool {
	if len(b.SelectedItems()) == 0 {
		return false
	}

	productItems := b.FindSelected(Finders.ByType(basket_item.TypeProduct))
	if len(productItems) == 0 {
		return false
	}

	for _, item := range productItems {
		if !item.Additions().GetProduct().IsAvailInStore() {
			return false
		}
	}

	return true
}

// ItemIds список id позиций
func (b *BasketData) ItemIds(itemTypes []basket_item.Type) []string {
	if len(b.SelectedItems()) == 0 {
		return make([]string, 0)
	}

	productItems := b.FindSelected(Finders.ByType(itemTypes...))
	ids := make([]string, 0, len(productItems))
	for _, item := range productItems {
		ids = append(ids, string(item.ItemId()))
	}

	return ids
}

// setSpaceId меняет регион относительно которого ведет расчеты корзина.
func (b *BasketData) setSpaceId(spaceId store_types.SpaceId) {
	if b.spaceId == spaceId {
		return
	}

	b.spaceId = spaceId
	for _, item := range b.SelectedItems() {
		item.SetSpaceId(spaceId)
	}
}

// setPriceColumn задает новую ценовую колонку, относительно которой рассчитываются цены в корзине
func (b *BasketData) setPriceColumn(priceColumn catalog_types.PriceColumn) {
	if b.priceColumn == priceColumn {
		return
	}

	b.priceColumn = priceColumn
	for _, item := range b.SelectedItems() {
		item.SetPriceColumn(priceColumn)
	}
}

func (b *BasketData) PriceColumn() catalog_types.PriceColumn {
	return b.priceColumn
}
