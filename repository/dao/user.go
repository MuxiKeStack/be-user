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
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewGORMUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{db: db}
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

type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Sid      string `gorm:"type=char(12);unique"`
	Nickname string `gorm:"type=varchar(20)"`
	Avatar   string
	Utime    int64 // 如果涉及跨国，整个系统统一使用UTC 0时区
	Ctime    int64
}
