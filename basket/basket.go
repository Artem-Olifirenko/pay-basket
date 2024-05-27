package basket

//go:generate mockgen -source=basket.go -destination=basket_mock.go -package=basket

import (
	"context"
	"errors"
	"fmt"
	"go.citilink.cloud/catalog_types"
	"go.citilink.cloud/citizap"
	citizap_factory "go.citilink.cloud/citizap/factory"
	database "go.citilink.cloud/libdatabase"
	"go.citilink.cloud/order/internal"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.citilink.cloud/order/internal/order/bonus/bonuses_for_payment"
	productv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/product/v1"
	servicev1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/catalog/service/v1"
	userv1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	"go.citilink.cloud/store_types"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RefresherBasket interface {
	SpaceId() store_types.SpaceId
	AddInfo(infos ...*Info)
	Remove(item *basket_item.Item, force bool) error
	User() *userv1.User
	Find(finder Finder) basket_item.Items
	FindOneById(id basket_item.UniqId) *basket_item.Item
	Configuration() *Configuration
	All() basket_item.Items
	SelectedItems() basket_item.Items
	SetHasPossibleConfiguration(v bool)
	HasPossibleConfiguration() bool
}

type itemRefresher interface {
	Refresh(ctx context.Context, basket RefresherBasket, logger *zap.Logger) error
}

func NewBasket(
	basketData *BasketData,
	itemFactory basket_item.ItemFactory,
	user *userv1.User,
	productClient productv1.ProductAPIClient,
	itemRefresher itemRefresher,
	loggerFactory citizap_factory.Factory,
	db database.DB,
	mrkOptions *markingOptions,
	subcontractOptions *subcontractServiceChangeOptions,
	bonusesForPaymentsCalculator *bonuses_for_payment.BonusesForPaymentAgent,
) *Basket {
	basket := &Basket{
		data:          basketData,
		itemFactory:   itemFactory,
		user:          user,
		itemRefresher: itemRefresher,
		productClient: productClient,
		loggerFactory: loggerFactory,
		markingOptions: &markingOptions{
			markingEnabled:         mrkOptions.markingEnabled,
			markingEnabledInCities: mrkOptions.markingEnabledInCities,
		},
		subcontractServiceChangeOptions: &subcontractServiceChangeOptions{
			subcontractServicesChangeEnabled: subcontractOptions.subcontractServicesChangeEnabled,
		},
		bonusAgent: bonusesForPaymentsCalculator,
	}

	basket.configuration = NewConfiguration(basket, productClient, db)

	// именно в данном месте мы отслеживаем изменения у пользователя. Это необходимо в связи с тем, что сам
	// заказ/корзина никак не могут отслеживать изменения пользователя, хотя у него в любой момент может измениться
	// ценовая колонка и выбранный город (а вместе с ним и регион).
	if user != nil {
		// мы проверяем, что корзина до сих пор обладает тем же самым идентификатором региона, что и
		// пользователь, к которому прикреплена корзина
		if store_types.SpaceId(user.GetSpaceId()) != basket.SpaceId() {
			basketData.setSpaceId(store_types.SpaceId(user.GetSpaceId()))
		}

		// проверяем, что ценовая колонка осталась прежней, что и при создании самой корзины, если это не так,
		// обновляем ценовую колонку в корзине
		if catalog_types.PriceColumn(user.GetPriceColumn()) != basketData.PriceColumn() {
			basketData.setPriceColumn(catalog_types.PriceColumn(user.GetPriceColumn()))
		}
	}

	return basket
}

type markingOptions struct {
	markingEnabledInCities internal.StringsContainer // в каких городах включена маркировка
	markingEnabled         bool                      // включена ли услуга маркировки
}

func NewMarkingOptions(
	markingEnabled bool,
	markingEnabledInCities internal.StringsContainer,
) *markingOptions {
	return &markingOptions{
		markingEnabled:         markingEnabled,
		markingEnabledInCities: markingEnabledInCities,
	}
}

type subcontractServiceChangeOptions struct {
	subcontractServicesChangeEnabled bool // включена ли подмена услуг субподрядчика
}

func NewSubcontractServiceChangeOptions(
	subcontractServicesChangeEnabled bool,
) *subcontractServiceChangeOptions {
	return &subcontractServiceChangeOptions{
		subcontractServicesChangeEnabled: subcontractServicesChangeEnabled,
	}
}

type Basket struct {
	data          *BasketData
	itemFactory   basket_item.ItemFactory
	user          *userv1.User
	configuration *Configuration
	itemRefresher itemRefresher
	productClient productv1.ProductAPIClient
	loggerFactory citizap_factory.Factory
	*markingOptions
	*subcontractServiceChangeOptions
	bonusAgent *bonuses_for_payment.BonusesForPaymentAgent
}

// Add добавляет позицию в корзину. Данный метод сделает всю работу за вас, нужно только передать необходимые параметры.
// Если по какой-то причине у вас уже есть готовая позиция, то стоит воспользоваться методом AddItem.
func (b *Basket) Add(
	ctx context.Context,
	itemId basket_item.ItemId,
	itemType basket_item.Type,
	parentUniqId basket_item.UniqId,
	count int,
	ignoreFairPrice bool,
) (*basket_item.Item, error) {
	if itemId == "" {
		return nil, internal.NewValidationError(errors.New("itemId is empty"))
	}

	err := itemType.Validate()
	if err != nil {
		return nil, internal.NewValidationError(errors.New("itemType invalid"))
	}

	if count <= 0 {
		return nil, internal.NewValidationError(errors.New("count less or equal 0"))
	}

	if itemType.Spec() == nil {
		return nil, fmt.Errorf("not found specification for item type '%s'", itemType)
	}

	if itemType.Spec().MustBeAChild() && parentUniqId == "" {
		return nil, fmt.Errorf("item '%s' must be a child", itemType)
	}

	if b.user != nil && b.user.B2B.IsB2BState && !itemType.Spec().IsAllowedForB2bUser() {
		return nil, fmt.Errorf("item can't be bought by b2b user")
	}

	var parentItem *basket_item.Item
	if parentUniqId != "" {
		parentItem = b.data.FindOneById(parentUniqId)
		if parentItem == nil {
			return nil, internal.NewNotFoundError(fmt.Errorf("parent item '%s' not found in basket", parentUniqId))
		}

		if !parentItem.Spec().CanHaveChild(itemType) {
			return nil, fmt.Errorf(
				"parent item '%s' can't have children of type %s",
				parentItem.Type(), itemType)
		}
	}

	// ограничение на кол-во позиций в корзине для устранения торможения запросов к БД при больших кол-вах товара
	if b.User() == nil && b.Count() >= 20 {
		return nil, internal.NewLogicErrorWithMsg(errors.New("anon user can`t add more than 20 positions"),
			"Вы добавили 20 разных позиций в корзину. Авторизуйтесь или зарегистрируйтесь, чтобы добавить больше")
	} else if b.User() != nil && !b.User().GetB2B().GetIsB2BState() && b.Count() >= 50 {
		return nil, internal.NewLogicErrorWithMsg(errors.New("user can`t add more than 50 positions"),
			"Вы добавили 50 разных позиций в корзину. Удалите ненужные позиции и добавьте товар повторно")
	} else if b.User() != nil && b.User().GetB2B().GetIsB2BState() && b.Count() >= 100 {
		return nil, internal.NewLogicErrorWithMsg(errors.New("B2B user can`t add more than 100 positions"),
			"Вы добавили 100 разных позиций в корзину. Удалите ненужные позиции и добавьте товар повторно")
	}

	item, err := b.itemFactory.Create(
		ctx,
		itemId,
		b.SpaceId(),
		itemType,
		count,
		parentItem,
		b.PriceColumn(),
		b.User(),
		ignoreFairPrice,
	)
	if err != nil {
		return nil, fmt.Errorf("can't create item with item factory: %w", err)
	}

	return b.AddItem(item)
}

func (b *Basket) BonusesForPayment(ctx context.Context) (*bonuses_for_payment.BonusesForPayment, error) {
	bonuses, err := b.bonusAgent.BonusesForPayment(ctx, b.All(), b.Cost())
	if err != nil {
		return nil, fmt.Errorf("can't calculate bonuses for payment: %w", err)
	}

	return bonuses, nil
}

// AddItem добавляет ранее созданную позицию. Если вам просто нужно добавить очередной товар или услугу, воспользуйтесь
// методом Add, а данный метод нужен для служебного использования.
func (b *Basket) AddItem(item *basket_item.Item) (*basket_item.Item, error) {
	if item.Type() == basket_item.TypeProduct && item.Additions().GetProduct().IsOEM() &&
		(b.user == nil || !b.user.GetB2B().GetIsB2BState()) {
		return nil, fmt.Errorf("item can't be bought by not b2b user")
	}

	addedItem, err := b.data.Add(item)
	if err != nil {
		return nil, fmt.Errorf("can't add item to basket: %w", err)
	}

	return addedItem, nil
}

func (b *Basket) Configuration() *Configuration {
	return b.configuration
}

// Remove удаляет позицию. Если в спецификации позиции указано, что нельзя удалять товар, то будет возвращена ошибка,
// это специальная защита от "плохих" пользователей. Но, если при работе возникает необходимость удаления позиции
// (например удалить подарок, или какую-нибудь не удаляемую услугу), то нужно передать флаг force=true
func (b *Basket) Remove(item *basket_item.Item, force bool) error {
	if item.Type() == basket_item.TypeConfiguration {
		// Просто берем и удаляем конфигурацию (метод удаления в корзине сам удалит рекурсивно всех детей). Не прибегая
		// к рекурсивному вызову методов удаления детей, в связи с тем, что у позиции в составе конфигурации
		// стоит правило "удалять нельзя".
		b.data.Remove(item)

		return nil
	}

	if !item.Spec().IsDeletable() && !force {
		return internal.NewLogicError(fmt.Errorf("item with type %s can't be deleted", item.Type()))
	}

	for _, child := range b.data.Find(Finders.ChildrenOf(item)) {
		err := b.Remove(child, force)
		if err != nil {
			return fmt.Errorf("can't delete child(%s:%s) of item(%s:%s): %w", child.ItemId(), child.UniqId(),
				item.ItemId(), item.UniqId(), err)
		}
	}

	b.data.Remove(item)

	return nil
}

// Fingerprint собирает информацию по всей корзине и выводит это в виде хэша. Данный хэш при сборе так же
// сортирует позиции заказа по идентификатору позиции, таким образом увеличивается кол-во одинаковых отпечатков у
// одинаковых корзин
func (b *Basket) Fingerprint() string {
	return b.data.Fingerprint()
}

func (b *Basket) FindOneById(id basket_item.UniqId) *basket_item.Item {
	return b.data.FindOneById(id)
}

func (b *Basket) FindByIds(ids ...basket_item.UniqId) []*basket_item.Item {
	return b.data.FindByIds(ids...)
}

func (b *Basket) Find(finder Finder) basket_item.Items {
	return b.data.Find(finder)
}

func (b *Basket) FindSelected(finder Finder) basket_item.Items {
	return b.data.FindSelected(finder)
}

func (b *Basket) Clear() {
	b.data.Clear()
}

func (b *Basket) Count() int {
	return b.data.Count()
}

func (b *Basket) CountSelected() int {
	return b.data.CountSelected()
}

func (b *Basket) CountUnselected() int {
	return b.Count() - b.CountSelected()
}

func (b *Basket) HasUserAddedItems() bool {
	return len(b.Find(Finders.FindUserAddedItems())) > 0
}

// IsMarkingAvailable проверяет, включена ли маркировка для корзины
func (b *Basket) IsMarkingAvailable() bool {
	if !b.markingEnabled {
		return false
	}
	for _, item := range b.All() {
		if item.Type() == basket_item.TypeProduct &&
			item.Additions() != nil &&
			item.Additions().GetProduct().IsMarked() {
			return true
		}
	}

	return false
}

// HasFnsTrackedProducts проверяет, есть ли в корзине прослеживаемые товары
func (b *Basket) HasFnsTrackedProducts() bool {
	for _, item := range b.All() {
		// не товар или нет аддишинал - пропускаем
		if item.Type() != basket_item.TypeProduct || item.Additions() == nil {
			continue
		}

		// найден прослеживаемый товар
		if item.Additions().GetProduct().IsFnsTracked() {
			return true
		}
	}

	return false
}

func (b *Basket) CancelSimulateProblems() {
	for _, it := range b.All() {
		it.CancelSimulateProblems()
	}
}

func (b *Basket) All() basket_item.Items {
	return b.data.All()
}

func (b *Basket) SelectedItems() basket_item.Items {
	return b.data.SelectedItems()
}

func (b *Basket) ToXItems() []*basket_item.XItem {
	return b.data.ToXItems()
}

func (b *Basket) Problems() []*Problem {
	return b.data.Problems()
}

func (b *Basket) Infos() []*Info {
	return b.data.Infos()
}

func (b *Basket) AddInfo(infos ...*Info) {
	b.data.infos = append(b.data.infos, infos...)
}

func (b *Basket) CommitInfo(infoId basket_item.InfoId) {
	for k, info := range b.data.infos {
		if info.info.Id() == infoId {
			b.data.infos = append(b.data.infos[:k], b.data.infos[k+1:]...)
			return
		}
	}
}

func (b *Basket) CommitAllInfos() {
	b.data.infos = nil
}

func (b *Basket) SpaceId() store_types.SpaceId {
	return b.data.SpaceId()
}

func (b *Basket) SetHasPossibleConfiguration(v bool) {
	b.data.SetHasPossibleConfiguration(v)
}

func (b *Basket) HasPossibleConfiguration() bool {
	return b.data.HasPossibleConfiguration()
}

func (b *Basket) AvailableConfiguration() bool {
	return b.data.AvailableConfiguration()
}

func (b *Basket) PriceColumn() catalog_types.PriceColumn {
	return b.data.priceColumn
}

func (b *Basket) CommitChanges() {
	b.data.CommitChanges()
}

func (b *Basket) IsChanged() bool {
	return b.data.IsChanged()
}

func (b *Basket) IsAllProductsInStore() bool {
	return b.data.IsAllProductsInStore()
}

func (b *Basket) Cost() int {
	return b.data.Cost()
}

func (b *Basket) AccruedBonus() int {
	bonusAmount := 0
	// Чтобы посчитать бонус от заказа для неавторизованных пользователей
	if b.PriceColumn() == catalog_types.PriceColumnRetail ||
		// Условия расчетов бонусов доработаны в рамках задачи WEB-35901. Важный нюанс в том, что ценовая колонка корзины
		// приравнивается к ценовой колонке пользователя в момент её создания из данных internal/order/basket/basket.go:66
		b.PriceColumn() == catalog_types.PriceColumnClub && b.user.GetHasClubCard() {
		bonusAmount = b.data.AccruedBonus()
	}

	return bonusAmount
}

func (b *Basket) Counts() *Counts {
	counts := &Counts{}
	for _, item := range b.All() {
		// на данный момент по бизнес логике в общем кол-ве и кол-во самих позиций никак не участвуют позиции в составе
		// конфигурации, типа конфигурация сама-в-себе позиция
		// Условие item.Type() != basket_item.TypeConfigurationAssemblyService нужно для того,
		// чтобы услугу по сборке считать услугой
		if item.Type().IsPartOfConfiguration() && item.Type() != basket_item.TypeConfigurationAssemblyService {
			continue
		}

		counts.All += item.Count()
		counts.AllPositions += 1
		if item.Type().IsProduct() {
			counts.Products += item.Count()
			counts.ProductPositions += 1
		}

		if item.Type().IsConfiguration() {
			counts.Configurations += item.Count()
		}

		if item.Type().IsService() {
			counts.Services += item.Count()
			counts.ServicePositions += 1
		}

		if item.Type().IsPresent() {
			counts.Presents += item.Count()
		}
	}

	return counts
}

// IsUser прикреплена ли корзина к пользователю
func (b *Basket) IsUser() bool {
	return b.user != nil
}

// User возвращает пользователя. Имейте ввиду, его может и не быть, если корзина не прикреплена к пользователю
func (b *Basket) User() *userv1.User {
	return b.user
}

// SetSpaceId меняет регион относительно которого ведет расчеты корзина. Данный метод возможно безошибочно применять
// только для корзины, которая не прикреплена к пользователю или при задании региона совпадающего с текущим регионом пользователя
func (b *Basket) SetSpaceId(spaceId store_types.SpaceId) error {
	if b.User() != nil && b.User().GetSpaceId() != string(spaceId) {
		return fmt.Errorf("can't change spaceId to basket differenct from user")
	}

	b.data.setSpaceId(spaceId)

	return nil
}

func (b *Basket) Refresh(ctx context.Context, actualizerItems ActualizerItems, logger *zap.Logger) error {
	// предварительно удаляем все проблемы, потому что они будут пересчитываться по ходу алгоритма
	for _, item := range b.All() {
		item.DeleteProblems()
	}

	// Сбрасываем данный флаг, он будет рассчитан далее по ходу алгоритма
	b.SetHasPossibleConfiguration(false)

	// выясняем тип пользователя
	// anon (user == nil) || b2c
	userType := servicev1.CustomerType_CUSTOMER_TYPE_B2C

	// b2b
	if b.IsUser() && b.User().B2B.IsB2BState {
		userType = servicev1.CustomerType_CUSTOMER_TYPE_B2B
	}

	// проверяем на консистентность данных в одном месте, чтобы не распихивать это по всем местам
	// проверяем услуги субподряда, на недоступность для типа пользователя (b2c => b2b / b2b => b2c)
	// + заменяем на аналогичные
	for _, item := range b.All() {
		parentItem := b.FindOneById(item.ParentUniqId())
		if item.IsChild() && parentItem == nil {
			logger.Error("Basket is in inconsistent state! Child item can't find his parent! We should kill "+
				"poor orphan... We fix it, and user don't see error, but it's a serious error",
				zap.String("child_uniq_id", string(item.UniqId())),
				zap.String("child_item_id", string(item.ItemId())),
				zap.String("parent_uniq_id", string(item.ParentUniqId())),
			)

			// удаляем специально из данных, чтобы не нарваться на правила и так далее
			b.data.Remove(item)
			continue
		}

		// подмена услуг субподряда, при смене типа пользователя (b2c => b2b / b2c => b2c)
		if b.subcontractServicesChangeEnabled &&
			item.Type() == basket_item.TypeSubcontractServiceForProduct {
			// получаем информациб о услуге
			response, err := b.productClient.FindServices(ctx, &productv1.FindServicesRequest{
				ProductId: string(parentItem.ItemId()),
				SpaceId:   string(item.SpaceId()),
			})
			if err != nil {
				st, ok := status.FromError(err)
				if !ok {
					return fmt.Errorf("can't get error status from find services response: %w", err)
				}

				if st.Code() == codes.NotFound {
					item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "товара с услугой нет в наличии"))
					continue
				}

				return fmt.Errorf("can't get subcontract services from catalog microservice: %w", err)
			}

			servicesForProduct := response.GetSubcontractServices()
			subcontractService, ok := servicesForProduct[string(item.ItemId())]
			if !ok {
				item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "услуга субподряда не предоставляется"))
				continue
			}

			// проверяем доступность услуги для типа пользователя
			if !b.subcontractServiceAllowedToUser(userType, subcontractService) {
				serviceReplaced := false

				// недоступна - пробуем заменить услугу на аналогичную (если указана)
				if subcontractService.LinkedServiceId != "" {
					// получаем инфу по подменной услуге
					replacementSubcontractService, ok := servicesForProduct[subcontractService.LinkedServiceId]
					if !ok {
						item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "услуга субподряда не предоставляется"))
						continue
					}

					replacesedSubcontractService, err := b.subcontractServiceAdd(ctx, item.ParentUniqId(), item.Count(), replacementSubcontractService)
					if err != nil {
						return fmt.Errorf("can't add subcontract service: %w", err)
					}

					info := basket_item.NewInfo(
						basket_item.InfoIdPositionChanged,
						"Услуга субподряда была заменена на аналогичную, доступную для текущего типа пользователя",
					)
					info.SetAdditions(&basket_item.InfoAdditions{
						ChangedItem: basket_item.ChangedItemInfoAdditions{
							ItemId: string(item.ItemId()),
							UniqId: string(item.UniqId()),
							Count:  item.Count(),
							Name:   item.Name(),
							Price:  item.Price(),
						},
					})
					b.AddInfo(NewInfo(replacesedSubcontractService, info))

					serviceReplaced = true
				}

				// удаляем недоступную услугу
				logger.Info(
					"product can't be founded in catalog. Item will be deleted from basket",
					citizap.ProductId(string(item.ItemId())),
					zap.String("uniq_id", string(item.UniqId())),
					citizap.SpaceId(string(item.SpaceId())),
				)

				// если заменили услугу - 2ое оповещение не нужно
				if !serviceReplaced {
					deleteInfo := basket_item.NewInfo(
						basket_item.InfoIdPositionRemoved,
						"Услуга субподряда удалена, в связи с недоступностью для текущего типа пользователя",
					)
					deleteInfo.SetAdditions(&basket_item.InfoAdditions{
						ChangedItem: basket_item.ChangedItemInfoAdditions{
							ItemId: string(item.ItemId()),
							UniqId: string(item.UniqId()),
							Count:  item.Count(),
							Name:   item.Name(),
							Price:  item.Price(),
						},
					})
					b.AddInfo(NewInfo(item, deleteInfo))
				}

				err := b.Remove(item, false)
				if err != nil {
					return fmt.Errorf("can't remove product: %w", err)
				}
			}
		}
	}

	// обновляем данные позиций: цены, наличие и т.п.
	err := b.itemRefresher.Refresh(ctx, b, logger)
	if err != nil {
		return err
	}

	for _, item := range b.data.All() {
		if item.Count() == 0 {
			b.data.Remove(item)
			continue
		}

		var parentItem *basket_item.Item
		if item.IsChild() {
			parentItem = b.data.FindOneById(item.ParentUniqId())
			if parentItem == nil {
				return fmt.Errorf(
					"can't find parent item %s of item %s:%s:%s",
					item.ParentUniqId(), item.UniqId(), item.ItemId(), item.Type())
			}

			if item.Spec().IsCountLessOrEqualThenParent() && !item.Spec().IsCountEqualToParentCount() {
				item.Rules().SetMaxCount(parentItem.Count())
				if item.Count() > parentItem.Count() {
					item.AddProblem(basket_item.NewProblem(
						basket_item.ProblemMaxCountExcess,
						"Кол-во позиции не должно превышать родительскую"))
				}
			}
		}

		aItem := actualizerItems.FindByItem(item)
		// @todo WEB-54644 одна из причин отсутствия aItem - internal/order/actualizer.go:93 Type() вычисляет неправильный тип обновленной позиции
		if aItem == nil {
			// Если в корзине есть подарок, но актуализатор его не вернул, то помечаем как "нет в наличии"
			if item.Type() == basket_item.TypePresent {
				item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "Товар не в наличии"))
			} else if !item.Type().IsPartOfConfiguration() {
				// Проверяем все типы позиций, кроме тех, которые являются частью состава конфигурации, так как
				// обработка ошибок, связанных с комплектующими конфигурациями берет на себя сама конфигурация и в ней будет указана проблема
				// непонятно почему эту позицию не вернул актуалазер.

				// Если актуализатор не вернул товар, то выставляем соответствующий problem
				item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "Позицию невозможно заказать"))
			}

			continue
		}

		// Актуализируем количество подарков
		if item.Type() == basket_item.TypePresent {
			item.FixCount(aItem.GetCount())
		}

		// при добавлении позиции типа "Сборка компьютера" для конфигарации мы отправляем только идентификатор
		// услуги, всю информацию (даже цену) мы получаем из БД вот таким диким способом...
		if item.Type() == basket_item.TypeConfigurationAssemblyService {
			item.FixName(aItem.GetName())
			if item.Price() == 0 {
				// в случае, если цену мы еще не обновляли, то и нечего сообщать, что цена изменилась с 0 на нормальную
				item.SetPrice(aItem.GetPrice())
			} else if item.Price() != aItem.GetPrice() {
				// приходится вот тут вот проверять а изменилась ли цена...
				info := basket_item.NewInfo(basket_item.InfoIdPriceChanged, "цена на услугу сборки конфигурации изменилась")
				info.Additionals().PriceChanged = basket_item.PriceChangedInfoAddition{
					From: item.Price(),
					To:   aItem.GetPrice(),
				}
				item.AddInfo(info)
				item.SetPrice(aItem.GetPrice())
			}
		}

		// На нашей стороне невозможно и пока неясно как получать информацию по самой конфигурации
		if item.Type() == basket_item.TypeConfiguration {
			item.FixName(aItem.GetName())
		}

		// на данный момент мы хоть и создаем полностью правильные услуги для комплектующих, все же мы не можем никак
		// обновлять цену, кроме как обращаться к данным, полученным из процедуры
		if item.Type() == basket_item.TypeConfigurationProductService {
			item.FixName(aItem.GetName())
			if item.Price() != aItem.GetPrice() {
				// приходится вот тут вот проверять а изменилась ли цена...
				info := basket_item.NewInfo(basket_item.InfoIdPriceChanged, "цена на услугу в конфигурации изменилась")
				info.Additionals().PriceChanged = basket_item.PriceChangedInfoAddition{
					From: item.Price(),
					To:   aItem.GetPrice(),
				}
				item.AddInfo(info)
				item.SetPrice(aItem.GetPrice())
			}
		}

		// На данный момент мы не можем сами руководствоваться количеством позиции в составе конфигурации, так что
		// мы надеемся на правила в БД
		if item.Type().IsPartOfConfiguration() {
			item.FixCount(aItem.GetCount())
		}

		// из БД нам приходит признак, о том, что такого товара не существует в регионе
		if aItem.GetNotExist() {
			// товар не существует в данном регионе
			item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "Позиция не в наличии"))
		}

		// К нам из БД приходят данные об особых ограничениях на кол-во товаров в один заказ
		reduceInfo := aItem.ReduceInfo()
		if reduceInfo != nil {
			item.Rules().SetMaxCount(reduceInfo.Count())
			if item.Count() > reduceInfo.Count() {
				item.AddProblem(basket_item.NewProblem(basket_item.ProblemMaxCountExcess, reduceInfo.Info()))
			}
		}

		// Если не было установлено ограничение по максимальному количеству одной позиции данными из БД
		// и это значение не было определено ранее, то устанавливаем дефолтное значение
		if item.Rules().MaxCount() == 0 {
			item.Rules().SetMaxCount(basket_item.LimitOnePosition)
		}

		if item.Price() == 0 && item.Type() != basket_item.TypePresent && item.Type() != basket_item.TypeConfiguration {
			item.AddProblem(basket_item.NewProblem(basket_item.ProblemNotAvailable, "Товар не в наличии"))
		}
	}

	// добавление заменённой услуги субподряда

	// особая обработка подарков, так как они пока полностью контролируются БД, нам приходится выдумывать, чтобы
	// следить за тем, появились ли подарки или наоборот убрались
	presentAItems := actualizerItems.FindByType(basket_item.TypePresent)
	if len(presentAItems) > 0 {
		presentItems := b.data.Find(Finders.ByType(basket_item.TypePresent))
		if len(presentAItems) > len(presentItems) {
			for _, aItem := range presentAItems {
				found := false
				for _, item := range presentItems {
					if aItem.GetItemId() == item.ItemId() {
						found = true
						break
					}
				}

				if !found {
					addedPresent, err := b.data.Add(
						basket_item.NewItem(
							aItem.GetItemId(),
							basket_item.TypePresent,
							aItem.GetName(),
							"",
							aItem.GetCount(),
							0,
							0,
							b.SpaceId(),
							b.PriceColumn(),
						))
					if err != nil {
						return fmt.Errorf("can`t add present to basket, %w", err)
					}

					// Указываем основной товар для подарка если он есть
					presentParents := b.data.Find(Finders.ByItemIds(aItem.GetParentItemId()))
					if len(presentParents) > 0 {
						err = addedPresent.MakeChildOf(presentParents[0])
						if err != nil {
							return fmt.Errorf("can`t set parent for present, %w", err)
						}
					}
				}
			}
		}
	}

	err = b.fixServiceForConfProduct(actualizerItems)
	if err != nil {
		return err
	}

	// Костыль для правильного подсчета стоимости конфигурации. На данный момент в БД есть баг, связанный с тем, что
	// при добавлении товара, который является ПО, то к нему добавляется услуга на установку, но вот БД считает это
	// дело без добавленной услуги.
	configuration := b.data.Find(Finders.ByType(basket_item.TypeConfiguration)).First()
	if configuration != nil {
		confPrice := 0
		confBonus := 0
		for _, item := range b.data.All() {
			if item.Type().IsPartOfConfiguration() {
				confPrice += item.Cost()
				confBonus += item.Bonus() * item.Count()
			}
		}

		configuration.SetBonus(confBonus)
		if configuration.Price() == 0 {
			// в случае, если цену мы еще не обновляли, то и нечего сообщать, что цена изменилась с 0 на нормальную
			configuration.SetPrice(confPrice)
		} else if configuration.Price() != confPrice {
			// приходится вот тут вот проверять а изменилась ли цена...
			info := basket_item.NewInfo(basket_item.InfoIdPriceChanged, "цена на конфигурацию изменилась")
			info.Additionals().PriceChanged = basket_item.PriceChangedInfoAddition{
				From: configuration.Price(),
				To:   confPrice,
			}
			configuration.AddInfo(info)
			configuration.SetPrice(confPrice)
		}
	}

	b.CommitChanges()

	return nil
}

// fixServiceForConfProduct исправляет отсутствие позиций установки ПО в случае добавления товара, который является ПО
//
// при добавлении позиций в состав конфигурации, которые являются программным обеспечением (windows и т.п.), за их
// установка берется доп. плата в виде услуги установки ПО. Так как все обновления мы получаем из актуалайзера,
// придеться искать новые добавленные позиции и сравнивать с тем что есть.
func (b *Basket) fixServiceForConfProduct(actualizerItems ActualizerItems) error {
	aItems := actualizerItems.FindByType(basket_item.TypeConfigurationProductService)
	if len(aItems) == 0 {
		return nil
	}

	items := b.data.Find(Finders.ByType(basket_item.TypeConfigurationProductService))
	for _, aItem := range aItems {
		var foundedItem *basket_item.Item
		for _, item := range items {
			if item.ItemId() == aItem.GetItemId() {
				foundedItem = item
				break
			}
		}

		if foundedItem != nil {
			continue
		}
		var parentItemOfServiceItem *basket_item.Item
		for _, item := range b.data.All() {
			if item.Type() == basket_item.TypeConfigurationProduct && item.ItemId() == aItem.GetParentItemId() {
				parentItemOfServiceItem = item
				break
			}
		}
		if parentItemOfServiceItem == nil {
			return fmt.Errorf("can't found parent item for configuration service")
		}

		service := basket_item.NewItem(
			aItem.GetItemId(),
			basket_item.TypeConfigurationProductService,
			aItem.GetName(),
			"",
			aItem.GetCount(),
			aItem.GetPrice(),
			0,
			b.SpaceId(),
			b.PriceColumn(),
		)
		err := parentItemOfServiceItem.AddChild(service)
		if err != nil {
			return err
		}
		service.Additions().SetConfiguration(parentItemOfServiceItem.Additions().GetConfiguration())

		_, err = b.data.Add(service)
		if err != nil {
			return err
		}
	}

	return nil
}

// subcontractServiceAllowedToUser проверка доступности услуги субподряда для типа пользователя b2c/b2b
func (b *Basket) subcontractServiceAllowedToUser(userType servicev1.CustomerType, service *servicev1.SubcontractService) bool {
	for _, t := range service.GetCustomerTypes() {
		if userType == t {
			return true
		}
	}

	return false
}

// subcontractServiceAdd добавление услуги субподряда
func (b *Basket) subcontractServiceAdd(
	ctx context.Context,
	parentUniqId basket_item.UniqId,
	count int,
	service *servicev1.SubcontractService,
) (*basket_item.Item, error) {
	item, err := b.Add(
		ctx,
		basket_item.ItemId(service.GetId()),
		basket_item.TypeSubcontractServiceForProduct,
		parentUniqId,
		count,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("can't add subcontractService to basket: %w", err)
	}

	return item, nil
}

func (b *Basket) Data() *BasketData {
	return b.data
}

type Counts struct {
	All              int
	Products         int
	Services         int
	AllPositions     int
	ProductPositions int
	ServicePositions int
	Configurations   int
	Presents         int
}

// BasketReadOnly корзина только на чтение. Данный интерфейс может обладать только методами, которые никаким образом
// не изменяют корзину. Его можно применять когда нужно куда-то передать корзину, но нужно убедиться, что никто и
// никогда ничего не изменит в ней в процессе работы.
type BasketReadOnly interface {
	FindOneById(id basket_item.UniqId) *basket_item.Item
	FindByIds(ids ...basket_item.UniqId) []*basket_item.Item
	Find(finder Finder) basket_item.Items
	FindSelected(finder Finder) basket_item.Items
	Count() int
	All() basket_item.Items
}
