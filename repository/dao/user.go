package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDuplicateUser = errors.New("用户冲突")
	ErrRecordNotFind = gorm.ErrRecordNotFound
)

type UserDAO interface {
	FindByStudentId(ctx context.Context, sid string) (User, error)
	Insert(ctx context.Context, u User) error
	UpdateSensitiveInfoById(ctx context.Context, user User) error
	FindById(ctx context.Context, uid int64) (User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func (dao *GORMUserDAO) FindById(ctx context.Context, uid int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", uid).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) UpdateSensitiveInfoById(ctx context.Context, user User) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).
		Model(&User{}).
		Where("id = ?", user.Id).
		Updates(map[string]any{
			"avatar":   user.Avatar,
			"nickname": user.Nickname,
			"utime":    now,
		}).Error
}

func (dao *GORMUserDAO) FindByStudentId(ctx context.Context, sid string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("sid = ?", sid).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		const duplicateErr uint16 = 1062 // 常量在编译期确定其值，每次函数调用不需要重新计算或分配内存。
		if me.Number == duplicateErr {
			return ErrDuplicateUser
		}
	}
	return err
}

func (dao *GORMUserDAO) DB() *gorm.DB {
	return dao.db
}

func (dao *GORMUserDAO) SetDB(db *gorm.DB) {
	dao.db = db
}

func NewGORMUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{db: db}
}

type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Sid      string `gorm:"unique; type:char(12)"`
	Nickname string `gorm:"type:varchar(20)"`
	Avatar   string
	Utime    int64 // 如果涉及跨国，整个系统统一使用UTC 0时区
	Ctime    int64
}
