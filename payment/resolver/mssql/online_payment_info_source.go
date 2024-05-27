package mssql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	citizap_factory "go.citilink.cloud/citizap/factory"
	database "go.citilink.cloud/libdatabase"
	simplestorageprometheusv2 "go.citilink.cloud/libsimple-storage-prometheus/v2"
	simplestoragev2 "go.citilink.cloud/libsimple-storage/v2"
	"go.citilink.cloud/store_types"
	"go.uber.org/zap"
	"hash/fnv"
	"strconv"
	"time"
)

type OnlinePaymentInfo struct {
	db            database.DB
	loggerFactory citizap_factory.Factory
	storage       simplestoragev2.Storage
}

func NewOnlinePaymentInfo(db database.DB, loggerFactory citizap_factory.Factory, storage simplestoragev2.Storage) *OnlinePaymentInfo {
	return &OnlinePaymentInfo{db: db, loggerFactory: loggerFactory, storage: storage}
}

type YandexPayShopData struct {
	SpaceId   string `db:"space_id"`
	IsEnabled bool   `db:"pay_flag"`
}

// GetRegions возвращает список всех регионов системы и информацию о доступности онлайн оплаты в этих регионах
func (o *OnlinePaymentInfo) GetRegions(ctx context.Context) (map[store_types.SpaceId]*YandexPayShopData, error) {
	regions := make(map[store_types.SpaceId]*YandexPayShopData)

	logger := o.loggerFactory.Create(ctx)
	unitName := "OnlinePaymentInfo:GetRegions"
	h := fnv.New64a()
	// ошибка глушится, т.к. не влияет на дальнейшую работу
	_, _ = h.Write([]byte(unitName))
	cacheKey := strconv.FormatUint(h.Sum64(), 10)
	cacheCtx := simplestorageprometheusv2.NewUnitContext(ctx, unitName)
	marshaledRaws, err := o.storage.Get(cacheCtx, cacheKey)
	if err != nil {
		notFoundErr := &simplestoragev2.NotFoundErr{}
		if !errors.As(err, &notFoundErr) {
			logger.Warn("can't get saved data from cache", zap.Error(err))
		}
	} else if len(marshaledRaws) > 0 {
		err := json.Unmarshal(marshaledRaws, &regions)
		if err != nil {
			logger.Warn("can't unmarshal saved data from cache",
				zap.Error(err),
				zap.String("data", string(marshaledRaws)))
		} else {
			return regions, nil
		}
	}

	dbCallCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := o.db.QueryxContext(dbCallCtx, "dbo.pr_store_shop_data", sql.Named("operation", "list"))

	if err != nil {
		return nil, fmt.Errorf("unable to get shop data from DB: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		raw := YandexPayShopData{}
		err := rows.StructScan(&raw)
		if err != nil {
			logger.Error("unable to struct scan", zap.Error(err))

			continue
		}

		regions[store_types.NewSpaceId(raw.SpaceId)] = &raw
	}

	if rows.Err() != nil {
		logger.Error("database error on rows iteration", zap.Error(rows.Err()))
		return nil, errors.New("database error, see log for more info")
	}

	// т.к. в рамках вызываемой процедуры мы всегда ожидаем не пустой список, а по утверждению старших программистов, база данных в редких случаях ошибочно может возвращать пустой список, то добавлена эта проверка
	if len(regions) != 0 {
		marshaledRaws, err = json.Marshal(regions)
		if err != nil {
			logger.Error("can't marshal regions", zap.Error(err))
		} else {
			err = o.storage.Set(cacheCtx, cacheKey, marshaledRaws, 600)
			if err != nil {
				logger.Warn("can't set regions in cache", zap.Error(err))
			}
		}
	}

	return regions, nil
}
