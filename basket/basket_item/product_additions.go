package basket_item

import (
	"go.citilink.cloud/catalog_types"
	"sync"
)

type ProductItemAdditions struct {
	isAvailInStore bool                     // 1
	vat            int                      // 2
	categoryId     catalog_types.CategoryId // 3
	// Является ли товар OEM софтом
	isOEM bool // 6
	// Товаров в наличии
	availTotal int // 7
	// Проверялось ли наличие товара. Данный признак проставляется при сравнении запрашиваемого кол-ва и наличия товара
	// на складах. Данный признак обнуляется при изменении запрашиваемого кол-ва позици.
	isCountMoreThenAvailChecked bool                          // 8
	creditPrograms              []catalog_types.CreditProgram // 9
	// Возможно ли этот товар доставлять в аутсорсовые точки выдачи DPD
	isAvailForDPD bool // 12
	// Является ли товар маркированным
	isMarked bool // 13
	// Цель покупки маркированного товара
	markedPurchaseReason MarkedPurchaseReason // 14
	// Уценённый товар
	isDiscounted bool // 15
	// Название категории
	categoryName string // 16
	// Бренд
	brandName    string // 17
	categoryPath string // 18
	shortName    string // 19
	// Является ли товар прослеживаемым
	isFnsTracked bool         // 20
	mx           sync.RWMutex `msgpack:"-"`
}

func NewProductItemAdditions(
	category catalog_types.CategoryId,
	creditPrograms []catalog_types.CreditProgram,
	vat int,
	availTotal int,
) *ProductItemAdditions {
	return &ProductItemAdditions{
		categoryId:           category,
		vat:                  vat,
		availTotal:           availTotal,
		creditPrograms:       creditPrograms,
		markedPurchaseReason: MarkedPurchaseReasonUnknown,
	}
}

func (p *ProductItemAdditions) IsAvailInStore() bool {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.isAvailInStore
}

func (p *ProductItemAdditions) SetIsAvailInStore(v bool) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.isAvailInStore = v
}

func (p *ProductItemAdditions) Vat() int {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.vat
}

func (p *ProductItemAdditions) SetVat(v int) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.vat = v
}

func (p *ProductItemAdditions) CategoryId() catalog_types.CategoryId {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.categoryId
}

func (p *ProductItemAdditions) SetCategoryId(v catalog_types.CategoryId) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.categoryId = v
}

func (p *ProductItemAdditions) CategoryName() string {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.categoryName
}

func (p *ProductItemAdditions) SetCategoryName(v string) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.categoryName = v
}

func (p *ProductItemAdditions) BrandName() string {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.brandName
}

func (p *ProductItemAdditions) SetBrandName(v string) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.brandName = v
}

func (p *ProductItemAdditions) CategoryPath() string {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.categoryPath
}

func (p *ProductItemAdditions) SetCategoryPath(v string) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.categoryPath = v
}

func (p *ProductItemAdditions) ShortName() string {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.shortName
}

func (p *ProductItemAdditions) SetShortName(v string) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.shortName = v
}

func (p *ProductItemAdditions) IsOEM() bool {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.isOEM
}

func (p *ProductItemAdditions) SetIsOEM(v bool) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.isOEM = v
}

func (p *ProductItemAdditions) AvailTotal() int {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.availTotal
}

func (p *ProductItemAdditions) SetAvailTotal(v int) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.availTotal = v
}

func (p *ProductItemAdditions) IsCountMoreThenAvailChecked() bool {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.isCountMoreThenAvailChecked
}

func (p *ProductItemAdditions) SetIsCountMoreThenAvailChecked(v bool) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.isCountMoreThenAvailChecked = v
}

func (p *ProductItemAdditions) CreditPrograms() []catalog_types.CreditProgram {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.creditPrograms
}

func (p *ProductItemAdditions) SetCreditPrograms(v []catalog_types.CreditProgram) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.creditPrograms = v
}

func (p *ProductItemAdditions) IsAvailForDPD() bool {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.isAvailForDPD
}

func (p *ProductItemAdditions) SetIsAvailForDPD(v bool) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.isAvailForDPD = v
}

func (p *ProductItemAdditions) MarkedPurchaseReason() MarkedPurchaseReason {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.markedPurchaseReason
}

func (p *ProductItemAdditions) SetMarkedPurchaseReason(m MarkedPurchaseReason) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.markedPurchaseReason = m
}

func (p *ProductItemAdditions) IsMarked() bool {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.isMarked
}

func (p *ProductItemAdditions) SetIsMarked(v bool) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.isMarked = v
}

func (p *ProductItemAdditions) IsDiscounted() bool {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.isDiscounted
}

func (p *ProductItemAdditions) SetIsDiscounted(v bool) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.isDiscounted = v
}

func (p *ProductItemAdditions) IsFnsTracked() bool {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return p.isFnsTracked
}

func (p *ProductItemAdditions) SetIsFnsTracked(isFnsTracked bool) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.isFnsTracked = isFnsTracked
}
