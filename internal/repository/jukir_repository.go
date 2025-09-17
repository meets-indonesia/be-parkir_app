package repository

import (
	"be-parkir/internal/domain/entities"

	"gorm.io/gorm"
)

type JukirRepository interface {
	Create(jukir *entities.Jukir) error
	GetByID(id uint) (*entities.Jukir, error)
	GetByUserID(userID uint) (*entities.Jukir, error)
	GetByQRToken(qrToken string) (*entities.Jukir, error)
	GetByJukirCode(jukirCode string) (*entities.Jukir, error)
	Update(jukir *entities.Jukir) error
	Delete(id uint) error
	List(limit, offset int) ([]entities.Jukir, int64, error)
	GetByAreaID(areaID uint) ([]entities.Jukir, error)
	GetPendingJukirs() ([]entities.Jukir, error)
}

type jukirRepository struct {
	db *gorm.DB
}

func NewJukirRepository(db *gorm.DB) JukirRepository {
	return &jukirRepository{db: db}
}

func (r *jukirRepository) Create(jukir *entities.Jukir) error {
	return r.db.Create(jukir).Error
}

func (r *jukirRepository) GetByID(id uint) (*entities.Jukir, error) {
	var jukir entities.Jukir
	err := r.db.Preload("User").Preload("Area").First(&jukir, id).Error
	if err != nil {
		return nil, err
	}
	return &jukir, nil
}

func (r *jukirRepository) GetByUserID(userID uint) (*entities.Jukir, error) {
	var jukir entities.Jukir
	err := r.db.Preload("User").Preload("Area").Where("user_id = ?", userID).First(&jukir).Error
	if err != nil {
		return nil, err
	}
	return &jukir, nil
}

func (r *jukirRepository) GetByQRToken(qrToken string) (*entities.Jukir, error) {
	var jukir entities.Jukir
	err := r.db.Preload("User").Preload("Area").Where("qr_token = ?", qrToken).First(&jukir).Error
	if err != nil {
		return nil, err
	}
	return &jukir, nil
}

func (r *jukirRepository) GetByJukirCode(jukirCode string) (*entities.Jukir, error) {
	var jukir entities.Jukir
	err := r.db.Preload("User").Preload("Area").Where("jukir_code = ?", jukirCode).First(&jukir).Error
	if err != nil {
		return nil, err
	}
	return &jukir, nil
}

func (r *jukirRepository) Update(jukir *entities.Jukir) error {
	return r.db.Save(jukir).Error
}

func (r *jukirRepository) Delete(id uint) error {
	return r.db.Delete(&entities.Jukir{}, id).Error
}

func (r *jukirRepository) List(limit, offset int) ([]entities.Jukir, int64, error) {
	var jukirs []entities.Jukir
	var count int64

	query := r.db.Model(&entities.Jukir{})
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("User").Preload("Area").Limit(limit).Offset(offset).Find(&jukirs).Error
	return jukirs, count, err
}

func (r *jukirRepository) GetByAreaID(areaID uint) ([]entities.Jukir, error) {
	var jukirs []entities.Jukir
	err := r.db.Preload("User").Preload("Area").Where("area_id = ?", areaID).Find(&jukirs).Error
	return jukirs, err
}

func (r *jukirRepository) GetPendingJukirs() ([]entities.Jukir, error) {
	var jukirs []entities.Jukir
	err := r.db.Preload("User").Preload("Area").Where("status = ?", entities.JukirStatusPending).Find(&jukirs).Error
	return jukirs, err
}
