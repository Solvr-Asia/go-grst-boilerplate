// Package user_repository provides data access for users.
package user_repository

import (
	"context"

	"veemon/entity"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindAll(ctx context.Context, params ListParams) ([]entity.User, int64, error)
	// UpdateFields applies a partial update to only the given columns and
	// returns the refreshed row. It returns gorm.ErrRecordNotFound if no live
	// row matches. Using column-scoped updates (instead of Save on a
	// previously-read struct) avoids clobbering columns changed concurrently.
	UpdateFields(ctx context.Context, id string, fields map[string]interface{}) (*entity.User, error)
	Delete(ctx context.Context, id string) error
}

type ListParams struct {
	Page      int
	Size      int
	Search    string
	SortBy    string
	SortOrder string
}

// allowedSortColumns whitelists the columns that may appear in ORDER BY, since
// the column name is concatenated into raw SQL and cannot be parameterized.
var allowedSortColumns = map[string]bool{
	"created_at": true,
	"updated_at": true,
	"name":       true,
	"email":      true,
	"status":     true,
}

type repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindAll(ctx context.Context, params ListParams) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.User{})

	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Whitelist sort column and direction. These are concatenated into the SQL
	// ORDER BY clause (GORM cannot parameterize identifiers), so they must never
	// come straight from user input.
	sortColumn := "created_at"
	if allowedSortColumns[params.SortBy] {
		sortColumn = params.SortBy
	}
	sortOrder := "desc"
	if params.SortOrder == "asc" {
		sortOrder = "asc"
	}

	offset := (params.Page - 1) * params.Size
	err := query.
		Order(sortColumn + " " + sortOrder).
		Offset(offset).
		Limit(params.Size).
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *repository) UpdateFields(ctx context.Context, id string, fields map[string]interface{}) (*entity.User, error) {
	result := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var user entity.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.User{}).Error
}
