package mssql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	citizapFactoryCtx "go.citilink.cloud/citizap/factory/ctx"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"testing"
)

func TestCommodityGroupChecker_CheckAvailabilityForResale(t *testing.T) {
	type args struct {
		ctx           context.Context
		itemIds       []basket_item.ItemId
		contractorInn string
	}

	commodityGroupCheckProcedure := "ItemsCatalog.get_check_items_resale"
	commodityGroupCheckRows := []string{"item_id", "is_allowed", "product_group"}

	tests := []struct {
		name           string
		args           args
		prepare        func(dbMock sqlmock.Sqlmock)
		want           func() (map[basket_item.ItemId]*CommodityGroupAvailability, error)
		wantLogMessage string
		wantErrMsg     string
	}{
		{
			name: "empty items ids",
			args: args{
				ctx:           context.Background(),
				itemIds:       []basket_item.ItemId{},
				contractorInn: "555777",
			},
			prepare: func(dbMock sqlmock.Sqlmock) {},
			want: func() (map[basket_item.ItemId]*CommodityGroupAvailability, error) {
				return make(map[basket_item.ItemId]*CommodityGroupAvailability, 0), nil
			},
		},
		{
			name: "mssql db error",
			args: args{
				ctx:           context.Background(),
				itemIds:       []basket_item.ItemId{"12345", "23456"},
				contractorInn: "555777",
			},
			prepare: func(dbMock sqlmock.Sqlmock) {
				dbMock.ExpectQuery(commodityGroupCheckProcedure).WillReturnError(errors.New("some db error"))
			},
			want: func() (map[basket_item.ItemId]*CommodityGroupAvailability, error) {
				return nil, fmt.Errorf("can't execute query")
			},
			wantErrMsg: "can't execute query: some db error",
		},
		{
			name: "scan error",
			args: args{
				ctx:           context.Background(),
				itemIds:       []basket_item.ItemId{"12345", "23456"},
				contractorInn: "555777",
			},
			prepare: func(dbMock sqlmock.Sqlmock) {
				dbMock.ExpectQuery(commodityGroupCheckProcedure).
					WithArgs(
						sql.Named("item_ids", "12345|23456"),
						sql.Named("participant_inn", "555777")).
					WillReturnRows(sqlmock.NewRows(commodityGroupCheckRows).
						AddRow("12345", false, "some group name").
						AddRow("23456", 75, "some group name 2"))
			},
			want: func() (map[basket_item.ItemId]*CommodityGroupAvailability, error) {
				commodityGroupAvailabilities := make(map[basket_item.ItemId]*CommodityGroupAvailability, 0)

				commodityGroupAvailabilities["12345"] = &CommodityGroupAvailability{
					IsAvailabilityForResale: false,
					Name:                    "some group name",
				}

				return commodityGroupAvailabilities, nil
			},
			wantLogMessage: "can't scan commodity group availabilities to struct from db",
		},
		{
			name: "success",
			args: args{
				ctx:           context.Background(),
				itemIds:       []basket_item.ItemId{"12345", "23456"},
				contractorInn: "555777",
			},
			prepare: func(dbMock sqlmock.Sqlmock) {
				dbMock.ExpectQuery(commodityGroupCheckProcedure).
					WithArgs(
						sql.Named("item_ids", "12345|23456"),
						sql.Named("participant_inn", "555777")).
					WillReturnRows(sqlmock.NewRows(commodityGroupCheckRows).
						AddRow("12345", false, "some group name").
						AddRow("23456", true, "some group name 2"))
			},
			want: func() (map[basket_item.ItemId]*CommodityGroupAvailability, error) {
				commodityGroupAvailabilities := make(map[basket_item.ItemId]*CommodityGroupAvailability, 0)

				commodityGroupAvailabilities["12345"] = &CommodityGroupAvailability{
					IsAvailabilityForResale: false,
					Name:                    "some group name",
				}
				commodityGroupAvailabilities["23456"] = &CommodityGroupAvailability{
					IsAvailabilityForResale: true,
					Name:                    "some group name 2",
				}

				return commodityGroupAvailabilities, nil
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db, dbMock, err := sqlmock.New()
			if err != nil {
				t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer func() {
				dbMock.ExpectClose()
				assert.NoError(t, db.Close())
			}()

			tt.prepare(dbMock)

			sqlxDB := sqlx.NewDb(db, "sqlmock")
			loggerObserver, logs := observer.New(zap.InfoLevel)
			loggerFactory := citizapFactoryCtx.New(zap.New(loggerObserver))

			commodityGroupChecker := NewCommodityGroupChecker(sqlxDB, loggerFactory)

			want, wantErr := tt.want()
			got, err := commodityGroupChecker.CheckAvailabilityForResale(context.Background(), tt.args.itemIds, tt.args.contractorInn)

			if tt.wantLogMessage != "" {
				assert.Equal(t, 1, logs.FilterMessage(tt.wantLogMessage).Len())
			}

			if wantErr == nil {
				assert.Equal(t, want, got)
			} else {
				assert.EqualError(t, err, tt.wantErrMsg)
				assert.Nil(t, got)
			}
		})
	}
}
