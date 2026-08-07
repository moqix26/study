package repo

import (
	"errors"
	"strings"

	"shortlink/internal/model"

	"gorm.io/gorm"
)

type LinkRepo struct {
	db *gorm.DB
}

func NewLinkRepo(db *gorm.DB) *LinkRepo {
	return &LinkRepo{db: db}
}

func (r *LinkRepo) AutoMigrate() error {
	return r.db.AutoMigrate(&model.Link{})
}

func (r *LinkRepo) Create(link *model.Link) error {
	return r.db.Create(link).Error
}

func (r *LinkRepo) FindByCode(code string) (*model.Link, error) {
	var link model.Link
	err := r.db.Where("code = ?", code).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *LinkRepo) IncrClick(code string) error {
	return r.db.Model(&model.Link{}).Where("code = ?", code).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}

func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique") || strings.Contains(msg, "1062")
}
