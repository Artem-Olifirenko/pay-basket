package mssql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	citizap_factory "go.citilink.cloud/citizap/factory"
	database "go.citilink.cloud/libdatabase"
	"go.citilink.cloud/order/internal/order/basket/basket_item"
	"go.uber.org/zap"
	"strings"
	"time"
)

func NewCommodityGroupChecker(db database.DB, loggerFactory citizap_factory.Factory) *CommodityGroupChecker {
	return &CommodityGroupChecker{db: db, loggerFactory: loggerFactory}
}

type AvailabilityChecker interface {
	CheckAvailabilityForResale(
		ctx context.Context,
		basketItemIds []basket_item.ItemId,
		contractorInn string,
	) (map[basket_item.ItemId]*CommodityGroupAvailability, error)
}

// CommodityGroupChecker проверяет товарные группы
type CommodityGroupChecker struct {
	loggerFactory citizap_factory.Factory
	db            database.DB
}

// CommodityGroupAvailability структура с информацией о доступности товарной группы
type CommodityGroupAvailability struct {
	// Название товарной группы в системе ГИС МТ
	Name string
	// Доступна ли товарная группа для перепродажи для текущего пользователя в системе ГИС МТ
	IsAvailabilityForResale bool
}

// CheckAvailabilityForResale проверяет, доступна ли товарная группа для перепродажи для b2b пользователя
func (c *CommodityGroupChecker) CheckAvailabilityForResale(
	ctx context.Context,
	basketItemIds []basket_item.ItemId,
	contractorInn string,
) (map[basket_item.ItemId]*CommodityGroupAvailability, error) {
	if len(basketItemIds) == 0 {
		return make(map[basket_item.ItemId]*CommodityGroupAvailability, 0), nil
	}

	logger := c.loggerFactory.Create(ctx)

	itemIdsStr := make([]string, 0, len(basketItemIds))
	for _, itemId := range basketItemIds {
		itemIdsStr = append(itemIdsStr, string(itemId))
	}

	args := []interface{}{
		sql.Named("item_ids", strings.Join(itemIdsStr, "|")),
		sql.Named("participant_inn", contractorInn),
	}

	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	rows, err := c.db.QueryxContext(dbCtx, "ItemsCatalog.get_check_items_resale", args...)
	if err != nil {
		return nil, fmt.Errorf("can't execute query: %w", err)
	}
	defer rows.Close()

	commodityGroupAvailabilities := make(map[basket_item.ItemId]*CommodityGroupAvailability, 0)
	for rows.Next() {
		var raw struct {
			ItemId             string `db:"item_id"`
			CommodityGroupName string `db:"product_group"`
			IsAvailability     bool   `db:"is_allowed"`
		}

		err := rows.StructScan(&raw)
		if err != nil {
			logger.Error("can't scan commodity group availabilities to struct from db", zap.Error(err))
			continue
		}

		commodityGroupAvailabilities[basket_item.ItemId(raw.ItemId)] = &CommodityGroupAvailability{
			IsAvailabilityForResale: raw.IsAvailability,
			Name:                    raw.CommodityGroupName,
		}
	}
	if rows.Err() != nil {
		logger.Error("database error on rows iteration", zap.Error(rows.Err()))
		return nil, errors.New("database error, see log for more info")
	}

	return commodityGroupAvailabilities, nil
}
