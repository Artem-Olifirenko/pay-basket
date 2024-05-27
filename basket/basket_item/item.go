package basket_item

//go:generate mockgen -source=item.go -destination=./item_mock.go -package basket_item

import (
	"context"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"go.citilink.cloud/catalog_types"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	"go.citilink.cloud/store_types"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

// ItemId идентификатор позиции (не уникален)
type ItemId string

func (i ItemId) IsConfiguration() bool {
	return i == ConfSpecialItemId
}

type NavType int

const (
	NavTypeProduct NavType = 2
	NavTypeService NavType = 13

	// LimitTotalGoods лимит количества добавленных товаров в корзину
	LimitTotalGoods  = 1000
	LimitOnePosition = 999
)

// UniqId уникальный идентификатор позиции
type UniqId string

type ItemFactory interface {
	Creatable(itemType Type) bool
	Create(
		ctx context.Context,
		itemId ItemId,
		spaceId store_types.SpaceId,
		itemType Type,
		count int,
		parentItem *Item,
		priceColumn catalog_types.PriceColumn,
		user *userv1.User,
		ignoreFairPrice bool,
	) (*Item, error)
}

func NewBasketItemFactory(factory ...ItemFactory) *CompositeFactory {
	return &CompositeFactory{factory}
}

type CompositeFactory struct {
	factories []ItemFactory
}

func (f *CompositeFactory) Add(factory ...ItemFactory) {
	f.factories = append(f.factories, factory...)
}

func (f *CompositeFactory) Creatable(itemType Type) bool {
	for _, fc := range f.factories {
		if fc.Creatable(itemType) {
			return true
		}
	}

	return false
}

func (f *CompositeFactory) Create(
	ctx context.Context,
	itemId ItemId,
	spaceId store_types.SpaceId,
	itemType Type,
	count int,
	parentItem *Item,
	priceColumn catalog_types.PriceColumn,
	user *userv1.User,
	ignoreFairPrice bool,
) (*Item, error) {
	for _, fc := range f.factories {
		if fc.Creatable(itemType) {
			item, err := fc.Create(ctx, itemId, spaceId, itemType, count, parentItem, priceColumn, user, ignoreFairPrice)
			if err != nil {
				return nil, fmt.Errorf("can't create item '%s' with factory: %w", itemType, err)
			}

			return item, nil
		}
	}

	return nil, fmt.Errorf("can't find factory to create item '%s", itemType)
}

type ServiceType int

const (
	ServiceTypeSubcontract             ServiceType = 22
	ServiceTypeDigital                 ServiceType = 62
	ServiceTypeInsuranceOfProperty     ServiceType = 5
	ServiceTypeProductInsurance        ServiceType = 52
	ServiceTypeDelivery                ServiceType = 53
	ServiceTypeForConfigurationFeature ServiceType = 54
)

// MarkedPurchaseReason Цель покупки маркированного товара

type MarkedPurchaseReason int

// Значения аналогичны БД, БД берут у Нава, также есть Госконтракт - 3
const (
	// MarkedPurchaseReasonUnknown не указан
	MarkedPurchaseReasonUnknown MarkedPurchaseReason = 0
	// MarkedPurchaseReasonForSelfNeeds для собственных нужд
	MarkedPurchaseReasonForSelfNeeds MarkedPurchaseReason = 1
	// MarkedPurchaseReasonForResale для перепродажи
	MarkedPurchaseReasonForResale MarkedPurchaseReason = 2
)

const (
	// GiftCertificateCategoryID название категории - "Подарочные сертификаты"
	GiftCertificateCategoryID catalog_types.CategoryId = 219
	// PCCaseCategoryID id категории компьютерного корпуса
	PCCaseCategoryID catalog_types.CategoryId = 41
)

// Item позиция корзины. Никогда не создавайте эту структуру напрямую, всегда пользуйтесь методом-конструктором NewItem
type Item struct {
	uniqId       UniqId // 1
	itemId       ItemId // 2
	itemType     Type   // 3
	parentUniqId UniqId // 4
	parentItemId ItemId // 5
	name         string // 6
	image        string // 7

	count             int // 8
	price             int // 9
	bonus             int // 10
	countMultiplicity int // 11 - количество в коробке/упаковке

	problems  []*Problem       // 13
	infos     map[InfoId]*Info // 14
	rules     Rules            // 15
	additions ItemAdditions    // 16

	// список постоянных проблем
	//
	// данный список для отладки проблем и функционала, который с ними связан. Он нужен в связи с тем, что порой крайне
	// затруднительно воспроизвести ту или иную проблему, и чтобы облегчить этот труд придумали фиксированные проблемы
	permanentProblems []*Problem // 17

	// 18 deleted
	// 19 deleted
	spaceId               store_types.SpaceId       // 20
	priceColumn           catalog_types.PriceColumn // 21
	commitFingerprint     string                    // 22
	isPrepaymentMandatory bool                      // 23
	// Флаг отвечающий за наличие у товара честной цены. Если true - значит в old_price хранится СТАРАЯ цена, а в ПЕРВОЙ
	// ценовой колонке находится честная(клубная) цена. Если false - необходимо игнорировать поведение описанное
	// для ignoreFairPrice.
	hasFairPrice bool // 24
	// Флаг отвечающий за желание пользователя игнорировать честную цену для данного товара. Если true - пользователь
	// игнорирует честную цену и необходимо использовать old_price для получения цены товара. Иначе необходимо
	// использовать ПЕРВУЮ ценовую колонку как цену товара.
	ignoreFairPrice bool // 25
	// Не сохраняемый флаг. Отвечает за статус изменения флага ignoreFairPrice
	ignoreFairPriceChanged bool
	// Информация о возможности приобретать данную позицию для перепродажи.
	allowResale *AllowResale // 26
	// Маркированный товар
	markedPurchaseReason MarkedPurchaseReason
	// Скидки по данной позиции
	discount ItemDiscount // 27
	// Флаг показывающий, можно ли переместить товар в конфигурацию с учетом текущей корзины
	movableToConfiguration bool // 28
	// Флаг показывающий, можно ли переместить товар из конфигурации в корзину
	movableFromConfiguration bool // 29
	// Флаг показывающий выбрана ли позиция для покупки
	isSelected bool // 30
}

type ItemDiscount struct {
	// Скидка позиции за примененный купон
	Coupon int
	// Скидка позиции по акциям
	Action int
	// Общая сумма всех скидок позиции
	Total int
	// Список примененных к позиции акций строкой через запятую
	AppliedPromotions string
}

// NewItem создает новую позицию.
func NewItem(
	itemId ItemId,
	itemType Type,
	name string,
	image string,
	count int,
	price int,
	bonus int,
	spaceId store_types.SpaceId,
	priceColumn catalog_types.PriceColumn,
) *Item {
	id := UniqId(uuid.NewString())
	return &Item{
		uniqId:               id,
		itemId:               itemId,
		itemType:             itemType,
		name:                 name,
		image:                image,
		count:                count,
		price:                price,
		bonus:                bonus,
		infos:                make(map[InfoId]*Info),
		spaceId:              spaceId,
		priceColumn:          priceColumn,
		markedPurchaseReason: MarkedPurchaseReasonUnknown,
		isSelected:           true,
	}
}

// NewConfigurationItem создает новую позицию для конфигуратора.
// Добавлено, так как сборка создается пользователем, это не готовый товар и именно поэтому,
// в отличие от конструктора NewItem, тут устанавливается countMultiplicity = 1
func NewConfigurationItem(
	price int,
	spaceId store_types.SpaceId,
	priceColumn catalog_types.PriceColumn,
) *Item {
	id := UniqId(uuid.NewString())
	return &Item{
		uniqId:               id,
		itemId:               ConfSpecialItemId,
		itemType:             TypeConfiguration,
		count:                1,
		price:                price,
		bonus:                0,
		countMultiplicity:    1,
		infos:                make(map[InfoId]*Info),
		spaceId:              spaceId,
		priceColumn:          priceColumn,
		markedPurchaseReason: MarkedPurchaseReasonUnknown,
		isSelected:           true,
	}
}

func (i *Item) IsSelected() bool {
	return i.isSelected
}

func (i *Item) SetIsSelected(isSelected bool) {
	i.isSelected = isSelected
}

func (i *Item) UniqId() UniqId {
	return i.uniqId
}

func (i *Item) ItemId() ItemId {
	return i.itemId
}

// Fingerprint собирает информацию по всей корзине и выводит это в виде хэша. Данный хэш при сборе так же
// сортирует позиции заказа по идентификатору позиции, таким образом увеличивается кол-во одинаковых отпечатков у
// одинаковых корзин
func (i *Item) Fingerprint() string {
	// данная реализация никогда не сможет вернуть ошибку, поэтому они здесь опущены + так же это крайне затрудняет
	// тестирование, так как для подмены хэша придется городить какую-то фабрику или любую другую решение, которое
	// только усложнит код.
	h := fnv.New64a()
	_, _ = h.Write([]byte(i.SpaceId()))
	_ = binary.Write(h, binary.LittleEndian, int32(i.PriceColumn()))
	_, _ = h.Write([]byte(i.ItemId()))
	_ = binary.Write(h, binary.LittleEndian, int32(i.Count()))
	_ = binary.Write(h, binary.LittleEndian, int32(i.Price()))

	return strconv.FormatUint(h.Sum64(), 10)
}

// AllowUnselect определяет, можно ли снимать галочку выкупа товара. При снятии товар не попадет в оформленный заказ,
// но останется в корзине.
func (i *Item) AllowUnselect() bool {
	// Нельзя убирать галочку для дочерних позиций
	if i.ParentUniqId() != "" {
		return false
	}
	// Нельзя убирать галочку для услуг доставки, подъема на этаж и для подарков
	if i.Type() == TypeLiftingService || i.Type() == TypeDeliveryService || i.Type() == TypePresent {
		return false
	}

	return true
}

func (i *Item) SetMovableToConfiguration(v bool) {
	i.movableToConfiguration = v
}

func (i *Item) IsMovableToConfiguration() bool {
	return i.movableToConfiguration
}

func (i *Item) SetMovableFromConfiguration(v bool) {
	i.movableFromConfiguration = v
}

func (i *Item) IsMovableFromConfiguration() bool {
	return i.movableFromConfiguration
}

func (i *Item) Type() Type {
	return i.itemType
}

func (i *Item) SetType(v Type) {
	i.itemType = v
}

func (i *Item) AllowResale() *AllowResale {
	return i.allowResale
}

func (i *Item) SetAllowResale(v *AllowResale) {
	i.allowResale = v
}

func (i *Item) SetDiscount(v ItemDiscount) {
	i.discount = v
}

func (i *Item) GetDiscount() ItemDiscount {
	return i.discount
}

func (i *Item) ParentUniqId() UniqId {
	return i.parentUniqId
}

func (i *Item) IsChild() bool {
	return i.parentUniqId != ""
}

func (i *Item) SetBonus(bonus int) {
	i.bonus = bonus
}

func (i *Item) FixBonus(bonus int) {
	i.bonus = bonus
}

func (i *Item) FixPrice(price int) {
	i.price = price
}

func (i *Item) FixCount(count int) {
	i.count = count
}

func (i *Item) FixName(name string) {
	i.name = name
}

func (i *Item) Bonus() int {
	return i.bonus
}

func (i *Item) setParent(parent *Item) error {
	if !parent.Spec().CanHaveChild(i.Type()) {
		return fmt.Errorf("item with type '%s' can't be child of item with type '%s'", i.Type(), parent.Type())
	}

	i.parentUniqId = parent.UniqId()
	i.parentItemId = parent.ItemId()

	return nil
}

func (i *Item) AddChild(child *Item) error {
	err := child.setParent(i)
	if err != nil {
		return fmt.Errorf("item '%s' can't add '%s' as child: %w", i.Type(), child.Type(), err)
	}

	return nil
}

func (i *Item) MakeChildOf(parent *Item) error {
	err := i.setParent(parent)
	if err != nil {
		return fmt.Errorf("item '%s' can't add '%s' as child: %w", parent.Type(), i.Type(), err)
	}

	return nil
}

func (i *Item) IsChildOf(parent *Item) bool {
	return i.ParentUniqId() == parent.UniqId()
}

func (i *Item) Name() string {
	return i.name
}

func (i *Item) Image() string {
	return i.image
}

func (i *Item) SetImage(image string) {
	i.image = image
}

func (i *Item) Count() int {
	return i.count
}

// SetCount установка заданного кол-ва товара
func (i *Item) SetCount(count int) error {
	if !i.Spec().IsCountChangeable() {
		return fmt.Errorf("count of position with type '%s' can't be changed", i.Type())
	}

	if count > LimitTotalGoods {
		count = LimitTotalGoods
	}

	count = i.CorrectCountFromMultiplicity(count)

	maxCount := i.Rules().MaxCount()
	if i.Rules().IsMaxCount() && count > maxCount {
		return NewMaxItemCountError(fmt.Errorf("can't set count %d more then max count %d for this item", count, maxCount), maxCount)
	}

	i.count = count
	// так как мы должны сообщать о том, что запрашиваемое кол-во товаров больше, чем товаров в наличии мы должны
	// отслеживать изменение запрашиваемого кол-ва, на данный момент это единственный способ это сделать без создания
	// тысячи абстракций, если таких проверок станет больше, то необходимо будет это решать при помощи эвентов
	if i.Type().IsProduct() {
		i.Additions().GetProduct().SetIsCountMoreThenAvailChecked(false)
	}

	return nil
}

// CorrectCountFromMultiplicity корректировка кол-ва товара исходя из показателя multiplicity(коробки/кратность товара)
func (i *Item) CorrectCountFromMultiplicity(count int) int {
	multiplicity := i.countMultiplicity
	if multiplicity == 0 || count%multiplicity == 0 {
		return count
	}
	if count <= multiplicity {
		return multiplicity
	}
	return count / multiplicity * multiplicity
}

func (i *Item) CommitChanges() {
	i.commitFingerprint = i.Fingerprint()
}

func (i *Item) IsChanged() bool {
	return i.Fingerprint() != i.commitFingerprint
}

func (i *Item) SetPrepaymentMandatory(isPrepaymentMandatory bool) {
	i.isPrepaymentMandatory = isPrepaymentMandatory
}

func (i *Item) IsPrepaymentMandatory() bool {
	return i.isPrepaymentMandatory
}

func (i *Item) SetMarkedPurchaseReason(purchaseReason MarkedPurchaseReason) {
	i.markedPurchaseReason = purchaseReason
}

func (i *Item) MarkedPurchaseReason() MarkedPurchaseReason {
	return i.markedPurchaseReason
}

func (i *Item) HasFairPrice() bool {
	return i.hasFairPrice
}

func (i *Item) SetHasFairPrice(hasFairPrice bool) {
	i.hasFairPrice = hasFairPrice
}

func (i *Item) IgnoreFairPrice() bool {
	return i.ignoreFairPrice
}

func (i *Item) SetIgnoreFairPrice(isIgnoreFairPrice bool) {
	i.ignoreFairPriceChanged = i.ignoreFairPrice != isIgnoreFairPrice
	i.ignoreFairPrice = isIgnoreFairPrice
}

func (i *Item) IgnoreFairPriceChanged() bool {
	return i.ignoreFairPriceChanged
}

func (i *Item) Price() int {
	return i.price
}

func (i *Item) Spec() *Spec {
	return i.Type().Spec()
}

func (i *Item) ToXItem() *XItem {
	isPresent := 0
	if i.Type() == TypePresent {
		isPresent = 1
	}

	isService := 0
	if i.Type().IsService() {
		isService = 1
	}

	xItem := &XItem{
		ItemId:                    string(i.ItemId()),
		Count:                     i.count,
		Count1:                    i.count,
		Price1:                    i.Price(),
		Price2:                    i.Price(),
		Price3:                    i.Price(),
		Discount:                  0,
		IsPresent:                 isPresent,
		Bonus:                     i.Bonus(),
		NavisionType:              int(i.Type().NavType()),
		ParentItemId:              string(i.parentItemId),
		IsService:                 isService,
		Price2WithoutLoyaltyBonus: i.Price(),
	}

	configurationAddition := i.Additions().GetConfiguration()
	if configurationAddition != nil {
		xItem.ConfId = configurationAddition.GetConfId()
		xItem.ConfType = int(configurationAddition.GetConfType())
		if i.Type() == TypeConfiguration {
			xItem.IsConfiguration = 1
		}
	}

	return xItem
}

func (i *Item) ToXmlItem() *XMLItem {
	return &XMLItem{
		ItemId: string(i.ItemId()),
		Count:  i.count,
	}
}

// Problems возвращает проблемы позиции
// Стоит обратить внимание, что так же этот метод возвращает и постоянные проблемы для отладки
func (i *Item) Problems() []*Problem {
	return append(i.problems, i.permanentProblems...)
}

// DeleteProblems удаляет все проблемы позиции (обычно производится перед очередной проверкой на наличие проблем у позиции)
func (i *Item) DeleteProblems() {
	i.problems = nil
}

func (i *Item) SimulateProblem(problems ...*Problem) {
	i.permanentProblems = append(i.permanentProblems, problems...)
}

func (i *Item) CancelSimulateProblems() {
	i.permanentProblems = nil
}

func (i *Item) CountMultiplicity() int {
	return i.countMultiplicity
}

func (i *Item) SetCountMultiplicity(countMultiplicity int) {
	i.countMultiplicity = countMultiplicity
}

func (i *Item) AddProblem(problems ...*Problem) {
	// Если позиция не выбрана (unselected), то для нее Problems делаем скрытыми, чтобы они не отображались на клиентах
	if !i.IsSelected() {
		for _, p := range problems {
			p.SetIsHidden(true)
		}
	}

	i.problems = append(i.problems, problems...)
}

func (i *Item) Infos() map[InfoId]*Info {
	return i.infos
}

func (i *Item) AddInfo(infos ...*Info) {
	for _, info := range infos {
		i.infos[info.Id()] = info
	}
}

func (i *Item) CommitInfo(id InfoId) {
	delete(i.infos, id)
}

func (i *Item) Additions() *ItemAdditions {
	return &i.additions
}

func (i *Item) Rules() *Rules {
	return &i.rules
}

func (i *Item) Cost() int {
	return i.price * i.count
}

// SpaceId возвращает идентификатор региона, относительно которого посчитано наличие и цена позиции
func (i *Item) SpaceId() store_types.SpaceId {
	return i.spaceId
}

// SetSpaceId задает идентификатор региона, относительно которого посчитаны цена и наличие для данной позиции.
// Данный метод нужно использовать только и только в том случае, когда производится смена региона для всего заказа
// (например пользователь сменил город тем или иным способом). Будьте крайне осторожны, и не используйте этот метод
// напрямую без знания дела
func (i *Item) SetSpaceId(spaceId store_types.SpaceId) {
	i.spaceId = spaceId
}

// SetPriceColumn задает ценовую колонку, относительно которой подсчитаны цена для позиции. Данный метод можно
// использовать только со знанием дела, нельзя просто так взять и поменять ценовую колонку, это нужно делать только
// тогда, когда стало ясно, что у пользователя, к которому прикреплен заказ, изменилась ценовая колонка
func (i *Item) SetPriceColumn(priceColumn catalog_types.PriceColumn) {
	i.priceColumn = priceColumn
}

// PriceColumn возвращает ценовую колонку, относительно которой подсчитана цена
func (i *Item) PriceColumn() catalog_types.PriceColumn {
	return i.priceColumn
}

func (i *Item) SetPrice(price int) {
	i.price = price
}

func (i *Item) SortTypeValue() int {
	switch i.Type() {
	case TypeProduct:
		return 1
	case TypeSubcontractServiceForProduct:
		return 2
	case TypeInsuranceServiceForProduct:
		return 2
	case TypeDigitalService:
		return 2
	case TypePropertyInsurance:
		return 2
	case TypeDeliveryService:
		return 2
	case TypeLiftingService:
		return 2
	case TypeConfiguration:
		return 3
	case TypeConfigurationProduct:
		return 4
	case TypeConfigurationProductService:
		return 5
	case TypeConfigurationAssemblyService:
		return 5
	default:
		return 100
	}
}

// CalculateMaxCount Рассчитывает максимальное количество товара учитывая наличие в ПВЗ и на складе
func (i *Item) CalculateMaxCount(maxAvailable int) int {
	return maxAvailable * i.CountMultiplicity()
}

type XItemer interface {
	ToXItem() *XItem
}

type XItemRecursiver interface {
	ToXItems() []*XItem
}

// XItem структура для представления в БД
//
// Описание структуры из PHP, оставить пока не исправим всем ошибки
//
//		return [
//		   'item_id' => $this->itemId,
//		   'qty' => $this->quantity,
//		   'qty1' => $this->quantity,
//		   'price' => $this->price,
//		   'discount' => $this->discount,
//		   'is_present' => (int)$this->isPresent ?: '',
//		   'conf_id' => $this->confId,
//		   'bonus' => $this->bonus,
//		   'price2' => $this->price2,
//		   'price3' => $this->price3,
//		   'price2_wo_loyaltybonus' => $this->priceTwoWithoutLoyaltyBonus,
//		   'type' => $this->navisionType,
//		   'parent_id' => $this->parentId,
//		   'is_conf' => (int)$this->isConfiguration ?: '',
//		   'conf_type' => $this->confType ?? null,
//		   'is_serv' => (int)$this->isService ?: '',
//		   'sub_name' => $this->subName,
//	  ];
type XItem struct {
	XMLName                   xml.Name `xml:"item"`
	ItemId                    string   `xml:"item_id"`
	Count                     int      `xml:"qty"`
	Count1                    int      `xml:"qty1"`
	Price1                    int      `xml:"price"`
	Price2                    int      `xml:"price2"`
	Price3                    int      `xml:"price3"`
	Discount                  int      `xml:"discount"`
	IsPresent                 int      `xml:"is_present"`
	ConfId                    string   `xml:"conf_id"`
	Bonus                     int      `xml:"bonus"`
	NavisionType              int      `xml:"type"`
	ParentItemId              string   `xml:"parent_id"`
	IsConfiguration           int      `xml:"is_conf"`
	ConfType                  int      `xml:"conf_type"`
	IsService                 int      `xml:"is_serv"`
	Price2WithoutLoyaltyBonus int      `xml:"price2_wo_loyaltybonus"`
	SubName                   string   `xml:"sub_name"`
}

type XMLItem struct {
	XMLName xml.Name `xml:"item"`
	ItemId  string   `xml:"id"`
	Count   int      `xml:"quantity"`
}

type ItemList struct {
	XMLName xml.Name `xml:"items"`
	Items   []*XMLItem
}

// MaxItemCountError ошибка превышения максимального количества позиции корзины.
type MaxItemCountError struct {
	err      error
	maxCount int
}

func NewMaxItemCountError(err error, maxCount int) *MaxItemCountError {
	return &MaxItemCountError{err, maxCount}
}

// MaxCount возвращает превышенное максимальное количество позиции корзины.
func (e *MaxItemCountError) MaxCount() int {
	return e.maxCount
}

func (e *MaxItemCountError) Error() string {
	return e.err.Error()
}

func (e *MaxItemCountError) Unwrap() error {
	return e.err
}

type ItemAdditions struct {
	// Данные о товаре, заполняются для всех позиций, которые являются товаром (например товар в корзине
	// и товар как дочерняя позиции конфигурации). У каждой позиции у которой isProduct истина должна обладать
	// дополнительной информацией о товаре
	Product *ProductItemAdditions // 1
	// Данные по конфигурации. Эти данные применяются для всех типов, связанных с конфигурацией
	// (товар, услуга, сама конфигурация)
	Configuration                *ConfiguratorItemAdditions // 2
	SubcontractServiceForProduct *SubcontractItemAdditions  // 3
	// Данные о любой услуге
	Service *Service     // 4
	mx      sync.RWMutex `msgpack:"-"`
}

func (i *ItemAdditions) GetProduct() *ProductItemAdditions {
	i.mx.RLock()
	defer i.mx.RUnlock()

	return i.Product
}

func (i *ItemAdditions) SetProduct(v *ProductItemAdditions) {
	i.mx.Lock()
	defer i.mx.Unlock()

	i.Product = v
}

func (i *ItemAdditions) GetConfiguration() *ConfiguratorItemAdditions {
	i.mx.RLock()
	defer i.mx.RUnlock()

	return i.Configuration
}

func (i *ItemAdditions) SetConfiguration(v *ConfiguratorItemAdditions) {
	i.mx.Lock()
	defer i.mx.Unlock()

	i.Configuration = v
}

func (i *ItemAdditions) GetSubcontractServiceForProduct() *SubcontractItemAdditions {
	i.mx.RLock()
	defer i.mx.RUnlock()

	return i.SubcontractServiceForProduct
}

func (i *ItemAdditions) SetSubcontractServiceForProduct(info *SubcontractItemAdditions) {
	i.mx.Lock()
	defer i.mx.Unlock()

	i.SubcontractServiceForProduct = info
}

func (i *ItemAdditions) GetService() *Service {
	i.mx.RLock()
	defer i.mx.RUnlock()

	return i.Service
}

func (i *ItemAdditions) SetService(v *Service) {
	i.mx.Lock()
	defer i.mx.Unlock()

	i.Service = v
}

type Specer interface {
	Spec() *Spec
}

type bitmap uint16

const (
	specPersonType  bitmap = 1
	specB2bUserType bitmap = 2
)

// Spec спецификация позиции.
//
// Отвечает за поведение позиции, ее расположение в дереве позиций, возможные дети и прочие условия, которую определяют
// позицию
type Spec struct {
	// какие типы детей могут быть у позиции. Именно по этому признаку так же определяется сам факт возможности
	// наличия детей у позиции
	childrenTypes []Type
	// типы пользователей, к которым можно применить позицию
	allowedUserTypes bitmap
	// кол-во позиций не должно превышать кол-во родительской позиции (для случаев когда позиция является дочерней
	// к другой позиции)
	isCountLessOrEqualThenParent bool
	// возможно ли просто взять и удалить позицию
	isDeletable bool
	// возможно ли изменение кол-ва для позиции
	isCountChangeable bool
	// только одна позиция с этим типом может присутствовать в корзине
	isOnlyOnePositionPossible bool
	// только одна позиция с таким типом может быть у родительской позиции. Например, только одна позиция услуги
	// страхования товара может быть у товара
	isOnlyOnePositionPerParent bool
	// Количество единиц в корзине всегда совпадает с количеством родительской позиции
	isCountEqualToParentCount bool
	// позиция обязательно должна быть чьим то дитем
	mustBeAChild bool
}

var typeToSpecs = map[Type]*Spec{
	TypeProduct: {
		childrenTypes:     []Type{TypeInsuranceServiceForProduct, TypeSubcontractServiceForProduct, TypeDigitalService, TypePresent},
		isDeletable:       true,
		isCountChangeable: true,
		allowedUserTypes:  specPersonType | specB2bUserType,
	},
	TypeInsuranceServiceForProduct: {
		mustBeAChild:                 true,
		isCountLessOrEqualThenParent: true,
		isDeletable:                  true,
		isCountChangeable:            true,
		isOnlyOnePositionPerParent:   true,
		allowedUserTypes:             specPersonType | specB2bUserType,
		isCountEqualToParentCount:    true,
	},
	TypeSubcontractServiceForProduct: {
		mustBeAChild:                 true,
		isCountLessOrEqualThenParent: true,
		isDeletable:                  true,
		isCountChangeable:            true,
		isOnlyOnePositionPerParent:   true,
		allowedUserTypes:             specPersonType | specB2bUserType,
		isCountEqualToParentCount:    true,
	},
	TypeDigitalService: {
		isCountLessOrEqualThenParent: true,
		isDeletable:                  true,
		isCountChangeable:            true,
		allowedUserTypes:             specPersonType | specB2bUserType,
		isCountEqualToParentCount:    true,
	},
	TypePresent: {
		isDeletable:       true,
		isCountChangeable: false,
		allowedUserTypes:  specPersonType | specB2bUserType,
	},
	TypePropertyInsurance: {
		isDeletable:               true,
		isCountChangeable:         true,
		isOnlyOnePositionPossible: true,
		allowedUserTypes:          specPersonType,
	},
	TypeConfiguration: {
		childrenTypes:             []Type{TypeConfigurationProduct, TypeConfigurationAssemblyService},
		isDeletable:               true,
		isCountChangeable:         true,
		isOnlyOnePositionPossible: true,
		allowedUserTypes:          specPersonType | specB2bUserType,
	},
	TypeConfigurationProduct: {
		mustBeAChild:      true,
		childrenTypes:     []Type{TypeConfigurationProductService},
		isDeletable:       false,
		isCountChangeable: false,
		allowedUserTypes:  specPersonType | specB2bUserType,
	},
	TypeConfigurationProductService: {
		mustBeAChild:      true,
		isDeletable:       false,
		isCountChangeable: false,
		allowedUserTypes:  specPersonType | specB2bUserType,
	},
	TypeConfigurationAssemblyService: {
		mustBeAChild:               true,
		isDeletable:                false,
		isCountChangeable:          false,
		isOnlyOnePositionPerParent: true,
		allowedUserTypes:           specPersonType | specB2bUserType,
	},
	TypeDeliveryService: {
		allowedUserTypes:          specPersonType | specB2bUserType,
		isOnlyOnePositionPossible: true,
	},
	TypeLiftingService: {
		allowedUserTypes:          specPersonType | specB2bUserType,
		isOnlyOnePositionPossible: false,
	},
}

func (s *Spec) CanHaveChildren() bool {
	return len(s.ChildrenTypes()) > 0
}

func (s *Spec) MustBeAChild() bool {
	return s.mustBeAChild
}

func (s *Spec) CanHaveChild(itemType Type) bool {
	if !s.CanHaveChildren() {
		return false
	}

	err := itemType.Validate()
	if err != nil {
		return false
	}

	for _, childType := range s.childrenTypes {
		if childType == itemType {
			return true
		}
	}

	return false
}

func (s *Spec) ChildrenTypes() []Type {
	return s.childrenTypes
}

// IsCountLessOrEqualThenParent кол-во позиций не должно превышать кол-во родительской позиции (для случаев когда позиция
// является дочерней к другой позиции)
func (s *Spec) IsCountLessOrEqualThenParent() bool {
	return s.isCountLessOrEqualThenParent
}

// IsDeletable возможно ли просто взять и удалить позицию
func (s *Spec) IsDeletable() bool {
	return s.isDeletable
}

// IsCountChangeable возможно ли изменение кол-ва для позиции
func (s *Spec) IsCountChangeable() bool {
	return s.isCountChangeable
}

// IsOnlyOnePositionPossible означает, что только одна позиция с этим типом может присутствовать в корзине
func (s *Spec) IsOnlyOnePositionPossible() bool {
	return s.isOnlyOnePositionPossible
}

// IsOnlyOnePositionPerParent только одна позиция с таким типом может быть у родительской позиции. Например, только
// одна позиция услуги страхования товара может быть у товара
func (s *Spec) IsOnlyOnePositionPerParent() bool {
	return s.isOnlyOnePositionPerParent
}

func (s *Spec) IsAllowedForPerson() bool {
	return (specPersonType & s.allowedUserTypes) > 0
}

func (s *Spec) IsAllowedForB2bUser() bool {
	return (specB2bUserType & s.allowedUserTypes) > 0
}

// IsCountEqualToParentCount количество единиц в корзине всегда совпадает с количеством родительской позиции
func (s *Spec) IsCountEqualToParentCount() bool {
	return s.isCountEqualToParentCount
}

type Items []*Item

type id string

func (is Items) Ids() []id {
	ids := make([]id, 0, len(is))
	for _, item := range is {
		ids = append(ids, id(item.ItemId()))
	}

	return ids
}

func (is Items) ItemIds() []ItemId {
	ids := make([]ItemId, 0, len(is))
	for _, item := range is {
		ids = append(ids, item.ItemId())
	}

	return ids
}

func (is Items) ItemIdsStrings() []string {
	ids := make([]string, 0, len(is))
	for _, item := range is {
		ids = append(ids, string(item.ItemId()))
	}

	return ids
}

func (is Items) UniqIds() []UniqId {
	ids := make([]UniqId, 0, len(is))
	for _, item := range is {
		ids = append(ids, item.UniqId())
	}

	return ids
}

func (is Items) ToXItems() []*XItem {
	xItems := make([]*XItem, 0, len(is))
	for _, i := range is {
		xItems = append(xItems, i.ToXItem())
	}

	return xItems
}

func (is Items) ToItemList() *ItemList {
	xmlItems := make([]*XMLItem, 0, len(is))
	for _, i := range is {
		xmlItems = append(xmlItems, i.ToXmlItem())
	}

	return &ItemList{Items: xmlItems}
}

func (is Items) First() *Item {
	if len(is) == 0 {
		return nil
	}

	return is[0]
}

func (is Items) Sort(parent *Item) Items {
	children := Items{}
	for _, item := range is {
		if parent == nil && item.ParentUniqId() == "" {
			children = append(children, item)
		} else if parent != nil && item.ParentUniqId() == parent.UniqId() {
			children = append(children, item)
		}
	}

	sort.SliceStable(children, func(i, j int) bool {
		if children[i].SortTypeValue() == children[j].SortTypeValue() {
			return children[i].Name() < children[j].Name()
		}

		return children[i].SortTypeValue() < children[j].SortTypeValue()
	})

	result := Items{}
	for _, child := range children {
		result = append(result, child)
		result = append(result, is.Sort(child)...)
	}

	return result
}

type ItemMap map[UniqId]*Item

func (im ItemMap) ToSlice() Items {
	is := make(Items, 0, len(im))
	for _, i := range im {
		is = append(is, i)
	}

	return is
}

func (im ItemMap) ToSliceOnlySelected() Items {
	selectedItems := make(Items, 0, len(im))
	for _, i := range im {
		if i.IsSelected() {
			selectedItems = append(selectedItems, i)
		}
	}

	return selectedItems
}

func CalculateBonus(user *userv1.User, prices *productv1.ProductPriceByRegion) int {
	bonusToUse := 0
	for _, bonus := range prices.GetBonuses() {
		if user != nil && bonus.LoyaltyStatus.GetCode() != user.GetLpStatusAsString() {
			continue
		}
		bonusToUse = int(bonus.GetBonusB2C())
		if user.GetB2B().GetIsB2BState() {
			bonusToUse = int(bonus.GetBonusB2B())
		}
		break
	}
	return bonusToUse
}
