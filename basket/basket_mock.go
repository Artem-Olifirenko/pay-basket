// Code generated by MockGen. DO NOT EDIT.
// Source: basket.go
//
// Generated by this command:
//
//	mockgen -source=basket.go -destination=basket_mock.go -package=basket
//
// Package basket is a generated GoMock package.
package basket

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	basket_item "go.citilink.cloud/order/internal/order/basket/basket_item"
	v1 "go.citilink.cloud/order/internal/specs/grpcclient/gen/citilink/profile/user/v1"
	store_types "go.citilink.cloud/store_types"
	zap "go.uber.org/zap"
)

// MockRefresherBasket is a mock of RefresherBasket interface.
type MockRefresherBasket struct {
	ctrl     *gomock.Controller
	recorder *MockRefresherBasketMockRecorder
}

// MockRefresherBasketMockRecorder is the mock recorder for MockRefresherBasket.
type MockRefresherBasketMockRecorder struct {
	mock *MockRefresherBasket
}

// NewMockRefresherBasket creates a new mock instance.
func NewMockRefresherBasket(ctrl *gomock.Controller) *MockRefresherBasket {
	mock := &MockRefresherBasket{ctrl: ctrl}
	mock.recorder = &MockRefresherBasketMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRefresherBasket) EXPECT() *MockRefresherBasketMockRecorder {
	return m.recorder
}

// AddInfo mocks base method.
func (m *MockRefresherBasket) AddInfo(infos ...*Info) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range infos {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "AddInfo", varargs...)
}

// AddInfo indicates an expected call of AddInfo.
func (mr *MockRefresherBasketMockRecorder) AddInfo(infos ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddInfo", reflect.TypeOf((*MockRefresherBasket)(nil).AddInfo), infos...)
}

// All mocks base method.
func (m *MockRefresherBasket) All() basket_item.Items {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "All")
	ret0, _ := ret[0].(basket_item.Items)
	return ret0
}

// All indicates an expected call of All.
func (mr *MockRefresherBasketMockRecorder) All() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "All", reflect.TypeOf((*MockRefresherBasket)(nil).All))
}

// Configuration mocks base method.
func (m *MockRefresherBasket) Configuration() *Configuration {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Configuration")
	ret0, _ := ret[0].(*Configuration)
	return ret0
}

// Configuration indicates an expected call of Configuration.
func (mr *MockRefresherBasketMockRecorder) Configuration() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Configuration", reflect.TypeOf((*MockRefresherBasket)(nil).Configuration))
}

// Find mocks base method.
func (m *MockRefresherBasket) Find(finder Finder) basket_item.Items {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Find", finder)
	ret0, _ := ret[0].(basket_item.Items)
	return ret0
}

// Find indicates an expected call of Find.
func (mr *MockRefresherBasketMockRecorder) Find(finder any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Find", reflect.TypeOf((*MockRefresherBasket)(nil).Find), finder)
}

// FindOneById mocks base method.
func (m *MockRefresherBasket) FindOneById(id basket_item.UniqId) *basket_item.Item {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOneById", id)
	ret0, _ := ret[0].(*basket_item.Item)
	return ret0
}

// FindOneById indicates an expected call of FindOneById.
func (mr *MockRefresherBasketMockRecorder) FindOneById(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOneById", reflect.TypeOf((*MockRefresherBasket)(nil).FindOneById), id)
}

// HasPossibleConfiguration mocks base method.
func (m *MockRefresherBasket) HasPossibleConfiguration() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasPossibleConfiguration")
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasPossibleConfiguration indicates an expected call of HasPossibleConfiguration.
func (mr *MockRefresherBasketMockRecorder) HasPossibleConfiguration() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasPossibleConfiguration", reflect.TypeOf((*MockRefresherBasket)(nil).HasPossibleConfiguration))
}

// Remove mocks base method.
func (m *MockRefresherBasket) Remove(item *basket_item.Item, force bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", item, force)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockRefresherBasketMockRecorder) Remove(item, force any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockRefresherBasket)(nil).Remove), item, force)
}

// SelectedItems mocks base method.
func (m *MockRefresherBasket) SelectedItems() basket_item.Items {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectedItems")
	ret0, _ := ret[0].(basket_item.Items)
	return ret0
}

// SelectedItems indicates an expected call of SelectedItems.
func (mr *MockRefresherBasketMockRecorder) SelectedItems() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectedItems", reflect.TypeOf((*MockRefresherBasket)(nil).SelectedItems))
}

// SetHasPossibleConfiguration mocks base method.
func (m *MockRefresherBasket) SetHasPossibleConfiguration(v bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetHasPossibleConfiguration", v)
}

// SetHasPossibleConfiguration indicates an expected call of SetHasPossibleConfiguration.
func (mr *MockRefresherBasketMockRecorder) SetHasPossibleConfiguration(v any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHasPossibleConfiguration", reflect.TypeOf((*MockRefresherBasket)(nil).SetHasPossibleConfiguration), v)
}

// SpaceId mocks base method.
func (m *MockRefresherBasket) SpaceId() store_types.SpaceId {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SpaceId")
	ret0, _ := ret[0].(store_types.SpaceId)
	return ret0
}

// SpaceId indicates an expected call of SpaceId.
func (mr *MockRefresherBasketMockRecorder) SpaceId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SpaceId", reflect.TypeOf((*MockRefresherBasket)(nil).SpaceId))
}

// User mocks base method.
func (m *MockRefresherBasket) User() *v1.User {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "User")
	ret0, _ := ret[0].(*v1.User)
	return ret0
}

// User indicates an expected call of User.
func (mr *MockRefresherBasketMockRecorder) User() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "User", reflect.TypeOf((*MockRefresherBasket)(nil).User))
}

// MockitemRefresher is a mock of itemRefresher interface.
type MockitemRefresher struct {
	ctrl     *gomock.Controller
	recorder *MockitemRefresherMockRecorder
}

// MockitemRefresherMockRecorder is the mock recorder for MockitemRefresher.
type MockitemRefresherMockRecorder struct {
	mock *MockitemRefresher
}

// NewMockitemRefresher creates a new mock instance.
func NewMockitemRefresher(ctrl *gomock.Controller) *MockitemRefresher {
	mock := &MockitemRefresher{ctrl: ctrl}
	mock.recorder = &MockitemRefresherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockitemRefresher) EXPECT() *MockitemRefresherMockRecorder {
	return m.recorder
}

// Refresh mocks base method.
func (m *MockitemRefresher) Refresh(ctx context.Context, basket RefresherBasket, logger *zap.Logger) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Refresh", ctx, basket, logger)
	ret0, _ := ret[0].(error)
	return ret0
}

// Refresh indicates an expected call of Refresh.
func (mr *MockitemRefresherMockRecorder) Refresh(ctx, basket, logger any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Refresh", reflect.TypeOf((*MockitemRefresher)(nil).Refresh), ctx, basket, logger)
}

// MockBasketReadOnly is a mock of BasketReadOnly interface.
type MockBasketReadOnly struct {
	ctrl     *gomock.Controller
	recorder *MockBasketReadOnlyMockRecorder
}

// MockBasketReadOnlyMockRecorder is the mock recorder for MockBasketReadOnly.
type MockBasketReadOnlyMockRecorder struct {
	mock *MockBasketReadOnly
}

// NewMockBasketReadOnly creates a new mock instance.
func NewMockBasketReadOnly(ctrl *gomock.Controller) *MockBasketReadOnly {
	mock := &MockBasketReadOnly{ctrl: ctrl}
	mock.recorder = &MockBasketReadOnlyMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBasketReadOnly) EXPECT() *MockBasketReadOnlyMockRecorder {
	return m.recorder
}

// All mocks base method.
func (m *MockBasketReadOnly) All() basket_item.Items {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "All")
	ret0, _ := ret[0].(basket_item.Items)
	return ret0
}

// All indicates an expected call of All.
func (mr *MockBasketReadOnlyMockRecorder) All() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "All", reflect.TypeOf((*MockBasketReadOnly)(nil).All))
}

// Count mocks base method.
func (m *MockBasketReadOnly) Count() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Count")
	ret0, _ := ret[0].(int)
	return ret0
}

// Count indicates an expected call of Count.
func (mr *MockBasketReadOnlyMockRecorder) Count() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Count", reflect.TypeOf((*MockBasketReadOnly)(nil).Count))
}

// Find mocks base method.
func (m *MockBasketReadOnly) Find(finder Finder) basket_item.Items {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Find", finder)
	ret0, _ := ret[0].(basket_item.Items)
	return ret0
}

// Find indicates an expected call of Find.
func (mr *MockBasketReadOnlyMockRecorder) Find(finder any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Find", reflect.TypeOf((*MockBasketReadOnly)(nil).Find), finder)
}

// FindByIds mocks base method.
func (m *MockBasketReadOnly) FindByIds(ids ...basket_item.UniqId) []*basket_item.Item {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range ids {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FindByIds", varargs...)
	ret0, _ := ret[0].([]*basket_item.Item)
	return ret0
}

// FindByIds indicates an expected call of FindByIds.
func (mr *MockBasketReadOnlyMockRecorder) FindByIds(ids ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByIds", reflect.TypeOf((*MockBasketReadOnly)(nil).FindByIds), ids...)
}

// FindOneById mocks base method.
func (m *MockBasketReadOnly) FindOneById(id basket_item.UniqId) *basket_item.Item {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindOneById", id)
	ret0, _ := ret[0].(*basket_item.Item)
	return ret0
}

// FindOneById indicates an expected call of FindOneById.
func (mr *MockBasketReadOnlyMockRecorder) FindOneById(id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOneById", reflect.TypeOf((*MockBasketReadOnly)(nil).FindOneById), id)
}

// FindSelected mocks base method.
func (m *MockBasketReadOnly) FindSelected(finder Finder) basket_item.Items {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindSelected", finder)
	ret0, _ := ret[0].(basket_item.Items)
	return ret0
}

// FindSelected indicates an expected call of FindSelected.
func (mr *MockBasketReadOnlyMockRecorder) FindSelected(finder any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindSelected", reflect.TypeOf((*MockBasketReadOnly)(nil).FindSelected), finder)
}
