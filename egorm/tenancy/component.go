package tenancy

import (
	"context"
	"github.com/soedev/soego-component/egorm"
	"github.com/soedev/soego/core/elog"
	"sync"
)

const PackageName = "component.tenancy"

// Tenancy 多数据源管理组件
type Tenancy struct {
	storage  Storage
	logger   *elog.Component
	dbs      map[string]*egorm.Component
	dbsMu    sync.RWMutex
	compName string
}

//NewTenancy 初始化租户池子
func NewTenancy(storage Storage, name string) *Tenancy {
	tenancy := defaultTenancy()
	tenancy.storage = storage
	tenancy.compName = name
	return tenancy
}

func defaultTenancy() *Tenancy {
	return &Tenancy{
		dbs:    make(map[string]*egorm.Component),
		logger: elog.EgoLogger.With(elog.FieldComponent(PackageName)),
	}
}

//Get 获取数据源
func (cmp *Tenancy) Get(ctx context.Context, tenancyKey string) (*egorm.Component, error) {
	//先重入锁
	cmp.dbsMu.RLock()
	if sqldb, ok := cmp.dbs[tenancyKey]; ok {
		if db, err := sqldb.DB(); err == nil {
			if err = db.Ping(); err == nil {
				cmp.dbsMu.RUnlock()
				return sqldb, nil
			}
		} else {
			db.Close()
			delete(cmp.dbs, tenancyKey)
			cmp.logger.Warn("数据源已从连接池中移除!", elog.FieldCustomKeyValue("tenantId", tenancyKey), elog.FieldErr(err))
			cmp.dbsMu.RUnlock()
		}
	}

	cmp.dbsMu.RUnlock()
	//在用原始锁 跨进程锁
	cmp.dbsMu.Lock()
	if sqldb, ok := cmp.dbs[tenancyKey]; ok {
		if db, err := sqldb.DB(); err == nil {
			if err = db.Ping(); err == nil {
				cmp.dbsMu.Unlock()
				return sqldb, nil
			}
		} else {
			db.Close()
			delete(cmp.dbs, tenancyKey)
			cmp.logger.Warn("数据源已从连接池中移除!", elog.FieldCustomKeyValue("tenantId", tenancyKey), elog.FieldErr(err))
		}
	}

	return cmp.getTenantDB(ctx, tenancyKey)
}

//连接租户
func (cmp *Tenancy) getTenantDB(ctx context.Context, tenancyKey string) (newDB *egorm.Component, err error) {
	defer func() {
		//解锁 并做recover
		cmp.dbsMu.Unlock()
		if rec := recover(); rec != nil {
			err = rec.(error)
			cmp.logger.Error("tenancy_egorm_recover", elog.FieldCustomKeyValue("tenantId", tenancyKey), elog.FieldErr(rec.(error)))
		}
	}()
	config, err := cmp.storage.Config(ctx, tenancyKey)
	if err != nil {
		return nil, err
	}
	//连接 会 Panic  一定要 recover 处理否则会整个系统崩溃
	newDB = egorm.DefaultContainer().Build(
		egorm.WithDialect(config.Dialect),
		egorm.WithDSN(config.DSN),
		egorm.WithConnMaxLifetime(config.ConnMaxLifetime),
		egorm.WithMaxIdleConns(config.MaxIdleConns),
		egorm.WithMaxOpenConns(config.MaxOpenConns),
		egorm.WithDebug(config.Debug),
	)
	cmp.dbs[tenancyKey] = newDB
	return cmp.dbs[tenancyKey], nil
}
