package repository

import (
	"context"
	"fmt"
	"github.com/MuxiKeStack/be-user/domain"
	"github.com/MuxiKeStack/be-user/pkg/logger"
	"github.com/MuxiKeStack/be-user/repository/cache"
	"github.com/MuxiKeStack/be-user/repository/dao"
	"github.com/go-redsync/redsync/v4"
	"time"
)

type CacheConsistencyUserRepository struct {
	dao   dao.GORMUserDAO
	cache cache.UserCache
	l     logger.Logger
	rs    *redsync.Redsync
}

func NewCacheConsistencyUserRepository(dao dao.GORMUserDAO, cache cache.UserCache, l logger.Logger, rs *redsync.Redsync) UserRepository {
	return &CacheConsistencyUserRepository{
		dao:   dao,
		cache: cache,
		l:     l,
		rs:    rs,
	}
}

// FindById 个人信息的变更需要保持强一致性
// 不一致性问题的原因：1. 部分失败 2. 并发更新
// 解决：使用 分布式锁+MySQL事务 来无限接近于强一致性
func (repo *CacheConsistencyUserRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	res, err := repo.cache.Get(ctx, uid)
	if err == nil {
		return res, nil
	}
	if err != cache.ErrKeyNotExists {
		// redis崩溃或者网络错误，用户量不大，MySQL撑得住，所以不降级处理
		repo.l.Error("访问Redis失败，查询用户缓存", logger.Error(err), logger.Int64("uid", uid))
	}

	mutexName := fmt.Sprintf("kstack:user:mutex:%d", uid)
	mutex := repo.rs.NewMutex(mutexName)
	if err = mutex.Lock(); err != nil {
		return domain.User{}, err
	}
	defer func() {
		if _, er := mutex.Unlock(); er != nil {
			repo.l.Error("解锁异常", logger.Error(err), logger.Int64("uid", uid))
		}
	}()
	u, err := repo.dao.FindById(ctx, uid) // 查
	if err != nil {
		return domain.User{}, err
	}
	res = repo.toDomain(u)
	// 回写
	if err = repo.cache.Set(ctx, res); err != nil {
		repo.l.Error("回写user失败", logger.Error(err), logger.Int64("uid", uid))
	}
	return res, nil
}

func (repo *CacheConsistencyUserRepository) UpdateSensitiveInfo(ctx context.Context, user domain.User) error {
	now := time.Now().UnixMilli()
	// 开启事务
	tx := repo.dao.DB().WithContext(ctx).Begin()
	// 更新数据库
	err := tx.Model(&dao.User{}).
		Where("id = ?", user.Id).
		Updates(map[string]any{
			"avatar":   user.Avatar,
			"nickname": user.Nickname,
			"utime":    now,
		}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	mutexName := fmt.Sprintf("kstack:user:mutex:%d", user.Id)
	mutex := repo.rs.NewMutex(mutexName)
	// 加分布式锁
	err = mutex.Lock()
	if err != nil {
		tx.Rollback()
		return err
	}
	// 删缓存
	err = repo.cache.Del(ctx, user.Id)
	if err != nil {
		tx.Rollback()
		return err
	}
	// 提交事务
	tx.Commit()
	// 释放分布式锁
	_, err = mutex.Unlock()
	repo.l.Error("释放分布式锁失败", logger.Error(err))
	return nil
}

func (repo *CacheConsistencyUserRepository) FindByStudentId(ctx context.Context, studentId string) (domain.User, error) {
	u, err := repo.dao.FindByStudentId(ctx, studentId)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}

func (repo *CacheConsistencyUserRepository) Create(ctx context.Context, u domain.User) error {
	return repo.dao.Insert(ctx, repo.toEntity(u))
}

func (repo *CacheConsistencyUserRepository) toDomain(u dao.User) domain.User {
	return domain.User{
		Id:        u.Id,
		StudentId: u.Sid,
		Avatar:    u.Avatar,
		Nickname:  u.Nickname,
		New:       u.Utime == u.Ctime, // 更新时间为创建时间说明是未更新过信息的新用户
		Utime:     time.UnixMilli(u.Utime),
		Ctime:     time.UnixMilli(u.Ctime),
	}
}

func (repo *CacheConsistencyUserRepository) toEntity(u domain.User) dao.User {
	return dao.User{
		Id:       u.Id,
		Sid:      u.StudentId,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
	}
}
