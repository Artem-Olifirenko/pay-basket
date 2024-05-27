package mssql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	citizapFactory "go.citilink.cloud/citizap/factory"
	citizapFactoryCtx "go.citilink.cloud/citizap/factory/ctx"
	database "go.citilink.cloud/libdatabase"
	simplestorage "go.citilink.cloud/libsimple-storage/v2"
	"go.citilink.cloud/order/internal/order/mock"
	"go.citilink.cloud/store_types"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"reflect"
	"testing"
)

func TestOnlinePaymentInfo_GetRegions(t *testing.T) {

	loggerFactory := citizapFactoryCtx.New(zap.NewNop())

	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	type constructorArgs struct {
		db            database.DB
		loggerFactory citizapFactory.Factory
		storage       func(ctrl *gomock.Controller) simplestorage.Storage
	}

	type testStruct struct {
		name            string
		constructorArgs constructorArgs
		setOptions      func(dbMock sqlmock.Sqlmock)
		loggerFactory   func() (citizapFactory.Factory, *observer.ObservedLogs)
		want            func() (map[store_types.SpaceId]*YandexPayShopData, error, []string)
	}

	tests := []testStruct{
		{
			name: "can't get saved data from cache and get data from db without errors",
			constructorArgs: constructorArgs{
				db:            sqlxDB,
				loggerFactory: loggerFactory,
				storage: func(ctrl *gomock.Controller) simplestorage.Storage {
					s := mock.NewMockStorage(ctrl)
					s.EXPECT().
						Get(gomock.Any(), gomock.Any()).
						Return(nil, fmt.Errorf("error on geting: %w", errors.New("can't get value as []byte"))).
						Times(1)

					s.EXPECT().
						Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil).
						Times(1)

					return s
				},
			},
			setOptions: func(
				dbMock sqlmock.Sqlmock,
			) {
				dbMock.ExpectQuery("dbo.pr_store_shop_data").
					WithArgs(
						sql.Named("operation", "list")).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"space_id", "pay_flag"}).
							AddRow("cl_msk", false).
							AddRow("cl_nvg", false).
							RowError(0, nil),
					)
			},
			loggerFactory: func() (citizapFactory.Factory, *observer.ObservedLogs) {
				zapCore, logs := observer.New(zap.DebugLevel)
				return citizapFactoryCtx.New(zap.New(zapCore)), logs
			},
			want: func() (map[store_types.SpaceId]*YandexPayShopData, error, []string) {
				return map[store_types.SpaceId]*YandexPayShopData{
					"cl_msk": {
						SpaceId:   "cl_msk",
						IsEnabled: false,
					},
					"cl_nvg": {
						SpaceId:   "cl_nvg",
						IsEnabled: false,
					}}, nil, []string{}
			},
		},
		{
			name: "can't get saved data from cache and error from db",
			constructorArgs: constructorArgs{
				db:            sqlxDB,
				loggerFactory: loggerFactory,
				storage: func(ctrl *gomock.Controller) simplestorage.Storage {
					s := mock.NewMockStorage(ctrl)
					s.EXPECT().
						Get(gomock.Any(), gomock.Any()).
						Return(nil, fmt.Errorf("error on geting: %w", errors.New("can't get value as []byte"))).
						Times(1)

					return s
				},
			},
			setOptions: func(
				dbMock sqlmock.Sqlmock,
			) {
				dbMock.ExpectQuery("dbo.pr_store_shop_data").
					WithArgs(
						sql.Named("operation", "list")).
					WillReturnError(fmt.Errorf("some error"))
			},
			loggerFactory: func() (citizapFactory.Factory, *observer.ObservedLogs) {
				zapCore, logs := observer.New(zap.DebugLevel)
				return citizapFactoryCtx.New(zap.New(zapCore)), logs
			},
			want: func() (map[store_types.SpaceId]*YandexPayShopData, error, []string) {
				return nil, fmt.Errorf("unable to get shop data from DB: %w", fmt.Errorf("some error")), []string{
					"can't get saved data from cache",
				}
			},
		},
		{
			name: "get saved data from cache without errors and a row from db is unable to parse",
			constructorArgs: constructorArgs{
				db:            sqlxDB,
				loggerFactory: loggerFactory,
				storage: func(ctrl *gomock.Controller) simplestorage.Storage {
					s := mock.NewMockStorage(ctrl)
					s.EXPECT().
						Get(gomock.Any(), gomock.Any()).
						Return([]byte{0x55}, nil).
						Times(1)

					s.EXPECT().
						Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil).
						Times(1)

					return s
				},
			},
			setOptions: func(
				dbMock sqlmock.Sqlmock,
			) {
				dbMock.ExpectQuery("dbo.pr_store_shop_data").
					WithArgs(
						sql.Named("operation", "list")).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"space_id", "pay_flag"}).
							AddRow("cl_msk", false).
							AddRow(0, "incorrect"),
					)
			},
			loggerFactory: func() (citizapFactory.Factory, *observer.ObservedLogs) {
				zapCore, logs := observer.New(zap.DebugLevel)
				return citizapFactoryCtx.New(zap.New(zapCore)), logs
			},
			want: func() (map[store_types.SpaceId]*YandexPayShopData, error, []string) {
				return map[store_types.SpaceId]*YandexPayShopData{
						"cl_msk": {
							SpaceId:   "cl_msk",
							IsEnabled: false,
						}}, nil, []string{
						`can't unmarshal saved data from cache`,
						`unable to struct scan`,
					}
			},
		},
		{
			name: "can't get saved data from cache and can't cache data",
			constructorArgs: constructorArgs{
				db:            sqlxDB,
				loggerFactory: loggerFactory,
				storage: func(ctrl *gomock.Controller) simplestorage.Storage {
					s := mock.NewMockStorage(ctrl)
					s.EXPECT().
						Get(gomock.Any(), gomock.Any()).
						Return([]byte{0x55}, nil).
						Times(1)

					s.EXPECT().
						Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(fmt.Errorf("error while caching data")).
						Times(1)

					return s
				},
			},
			setOptions: func(
				dbMock sqlmock.Sqlmock,
			) {
				dbMock.ExpectQuery("dbo.pr_store_shop_data").
					WithArgs(
						sql.Named("operation", "list")).
					WillReturnRows(
						sqlmock.
							NewRows([]string{"space_id", "pay_flag"}).
							AddRow("cl_msk", false),
					)
			},
			loggerFactory: func() (citizapFactory.Factory, *observer.ObservedLogs) {
				zapCore, logs := observer.New(zap.DebugLevel)
				return citizapFactoryCtx.New(zap.New(zapCore)), logs
			},
			want: func() (map[store_types.SpaceId]*YandexPayShopData, error, []string) {
				return map[store_types.SpaceId]*YandexPayShopData{
						"cl_msk": {
							SpaceId:   "cl_msk",
							IsEnabled: false,
						}}, nil, []string{
						`can't set regions in cache`,
					}
			},
		},
		{
			name: "can't get data from db",
			constructorArgs: constructorArgs{
				db:            sqlxDB,
				loggerFactory: loggerFactory,
				storage: func(ctrl *gomock.Controller) simplestorage.Storage {
					s := mock.NewMockStorage(ctrl)
					s.EXPECT().
						Get(gomock.Any(), gomock.Any()).
						Return(nil, fmt.Errorf("error while getting data from cache")).
						Times(1)

					return s
				},
			},
			setOptions: func(
				dbMock sqlmock.Sqlmock,
			) {
				dbMock.ExpectQuery("dbo.pr_store_shop_data").
					WithArgs(
						sql.Named("operation", "list")).
					WillReturnError(fmt.Errorf("error from db"))
			},
			loggerFactory: func() (citizapFactory.Factory, *observer.ObservedLogs) {
				zapCore, logs := observer.New(zap.DebugLevel)
				return citizapFactoryCtx.New(zap.New(zapCore)), logs
			},
			want: func() (map[store_types.SpaceId]*YandexPayShopData, error, []string) {
				return nil, fmt.Errorf("unable to get shop data from DB: %w", fmt.Errorf("error from db")), []string{
					`can't get saved data from cache`,
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			loggerFactory, logs := test.loggerFactory()
			want, wantErr, wantLog := test.want()
			test.setOptions(mockDb)

			p := NewOnlinePaymentInfo(sqlxDB, loggerFactory, test.constructorArgs.storage(ctrl))
			got, err := p.GetRegions(context.Background())

			assert.Equal(t, wantErr, err)
			if !reflect.DeepEqual(want, got) {
				t.Errorf("GetIds() got = %v, want %v", got, want)
			}
			for _, msg := range wantLog {
				assert.Equal(t, 1, logs.FilterMessage(msg).Len())
			}
		})
	}
}
