package mssql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go.citilink.cloud/catalog_types"
	citizap_factory "go.citilink.cloud/citizap/factory"
	database "go.citilink.cloud/libdatabase"
	simplestorageprometheus "go.citilink.cloud/libsimple-storage-prometheus/v2"
	simplestorage "go.citilink.cloud/libsimple-storage/v2"
	"go.citilink.cloud/store_types"
	"go.uber.org/zap"
	"time"
)

func NewCityByService(
	db database.DB,
	cache simplestorage.Storage,
	loggerFactory citizap_factory.Factory,
) *CityByService {
	return &CityByService{
		cache:         simplestorage.WrapPrefixed("city-by-job", cache),
		loggerFactory: loggerFactory,
		db:            db,
	}
}

type CityByService struct {
	loggerFactory citizap_factory.Factory
	cache         simplestorage.Storage
	db            database.DB
}

func (r *CityByService) GetAvailableCitiesKladrByServiceIds(
	ctx context.Context,
	spaceId store_types.SpaceId,
	serviceId string,
	priceCol catalog_types.PriceColumn,
) (map[store_types.KladrId]store_types.KladrId, error) {
	if spaceId == "" || serviceId == "" {
		return nil, nil
	}

	logger := r.loggerFactory.Create(ctx)
	key := "AvailCitiesByService:" + string(spaceId) + serviceId
	cacheCtx := simplestorageprometheus.NewUnitContext(ctx, "AvailCitiesByService")
	marshaledCityServices, err := r.cache.Get(cacheCtx, key)
	if err != nil {
		notFoundErr := &simplestorage.NotFoundErr{}
		if !errors.As(err, &notFoundErr) {
			logger.Warn("can't get marshaled avail cities by service from cache", zap.Error(err))
		}
	} else if marshaledCityServices != nil {
		var res map[store_types.KladrId]store_types.KladrId
		err := json.Unmarshal(marshaledCityServices, &res)
		if err != nil {
			logger.Warn("can't unmarshal marshaled avail cities by service from cache", zap.Error(err))
		}

		return res, nil
	}

	args := []interface{}{
		sql.Named("operation", "get_city_by_job"),
		sql.Named("space_id", spaceId),
		sql.Named("client_status", priceCol.ToOldFormat()),
		sql.Named("data", "<id>"+serviceId+"</id>"),
	}

	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	rows, err := r.db.QueryxContext(dbCtx, "dbo.pr_store_getCities", args...)
	if err != nil {
		return nil, fmt.Errorf("can't execute query: %w", err)
	}
	defer rows.Close()

	kladrIds := make(map[store_types.KladrId]store_types.KladrId, 0)
	for rows.Next() {
		var raw struct {
			Id store_types.KladrId `db:"id"`
		}

		err := rows.StructScan(&raw)
		if err != nil {
			logger.Error("can't scan avail cities by service to struct from db", zap.Error(err))
			continue
		}
		kladrIds[raw.Id] = raw.Id
	}

	if rows.Err() != nil {
		logger.Error("database error on rows iteration", zap.Error(rows.Err()))
		return nil, errors.New("database error, see log for more info")
	}

	marshaledCityServices, err = json.Marshal(kladrIds)
	if err != nil {
		logger.Error("can't marshal avail cities service", zap.Error(err))
	} else {
		err = r.cache.Set(cacheCtx, key, marshaledCityServices, 600)
		if err != nil {
			logger.Warn("can't save marshaled avail cities by service in cache", zap.Error(err))
		}
	}

	return kladrIds, nil
}

type availCitiesSource interface {
	GetAvailableCitiesKladrByServiceIds(
		ctx context.Context,
		spaceId store_types.SpaceId,
		serviceId string,
		priceCol catalog_types.PriceColumn,
	) (map[store_types.KladrId]store_types.KladrId, error)
}

func NewServiceChecker(
	source availCitiesSource,
	loggerFactory citizap_factory.Factory,
) *ServiceChecker {
	return &ServiceChecker{
		source:        source,
		loggerFactory: loggerFactory,
	}
}

type ServiceChecker struct {
	loggerFactory citizap_factory.Factory
	source        availCitiesSource
}

func (r *ServiceChecker) IsAvailableInCity(
	ctx context.Context,
	spaceId store_types.SpaceId,
	kladrId store_types.KladrId,
	serviceId string,
	priceCol catalog_types.PriceColumn,
) (bool, error) {
	if kladrId == "" {
		return false, nil
	}
	availKladrIds, err := r.source.GetAvailableCitiesKladrByServiceIds(ctx, spaceId, serviceId, priceCol)
	if err != nil {
		return false, fmt.Errorf("can't get avail city: %w", err)
	}
	_, kladrIdExists := availKladrIds[kladrId]
	return kladrIdExists, nil
}
