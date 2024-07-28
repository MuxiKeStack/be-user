package ioc

import (
	"github.com/MuxiKeStack/be-user/pkg/logger"
	"github.com/MuxiKeStack/be-user/repository"
	"github.com/MuxiKeStack/be-user/repository/cache"
	"github.com/MuxiKeStack/be-user/repository/dao"
	"github.com/go-redsync/redsync/v4"
	"gorm.io/gorm"
)

func InitUserRepository(db *gorm.DB, cache cache.UserCache, l logger.Logger, rs *redsync.Redsync) repository.UserRepository {
	d := dao.GORMUserDAO{}
	d.SetDB(db)
	return repository.NewCacheConsistencyUserRepository(d, cache, l, rs)
}
