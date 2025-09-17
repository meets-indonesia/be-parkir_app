package repository

import (
	"be-parkir/internal/domain/entities"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *entities.User) error
	GetByID(id uint) (*entities.User, error)
	GetByEmail(email string) (*entities.User, error)
	Update(user *entities.User) error
	Delete(id uint) error
	List(limit, offset int) ([]entities.User, int64, error)
	GetByRole(role entities.UserRole) ([]entities.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *entities.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*entities.User, error) {
	var user entities.User
	err := r.db.Preload("JukirProfile").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*entities.User, error) {
	var user entities.User
	err := r.db.Preload("JukirProfile").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *entities.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&entities.User{}, id).Error
}

func (r *userRepository) List(limit, offset int) ([]entities.User, int64, error) {
	var users []entities.User
	var count int64

	query := r.db.Model(&entities.User{})
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("JukirProfile").Limit(limit).Offset(offset).Find(&users).Error
	return users, count, err
}

func (r *userRepository) GetByRole(role entities.UserRole) ([]entities.User, error) {
	var users []entities.User
	err := r.db.Preload("JukirProfile").Where("role = ?", role).Find(&users).Error
	return users, err
}
