package dao

import (
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"log"
	"myweb/app/user/internal/model/po"
)

type UserDao struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserDao 新建UserDao对象
// initialize the user data access object(Dao) level instance
func NewUserDao(db *gorm.DB, logger *zap.Logger) *UserDao {
	return &UserDao{
		db:     db,
		logger: logger,
	}
}

// ContainsName 查询是否存在该用户名
// query whether the given name was registered
func (d UserDao) ContainsName(name string) bool {
	return !errors.Is(d.db.Where("name = ?", name).Take(&po.User{}).Error, gorm.ErrRecordNotFound)
}

// ContainsEmail 查询是否存在用户邮箱
// query whether the given email was registered
func (d UserDao) ContainsEmail(email string) bool {
	return !errors.Is(d.db.Where("email = ?", email).Take(&po.User{}).Error, gorm.ErrRecordNotFound)
}

// CreateUser 创建用户信息,传递指针给函数所以会将id赋给user参数
// Creates the user information, passes the pointer to the function so the id is assigned to the user argument
func (d UserDao) CreateUser(user *po.User) {
	d.db.Create(&user)
}

// GetUser 查询符合条件的用户信息，符合条件的用户存在则返回true
// get user information,if user exists,return true
func (d UserDao) GetUser(user *po.User) bool {
	if err := d.db.Where(user).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
		log.Print("GetUser error : ", err)
		return false
	}
	return true
}
