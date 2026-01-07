package repository

import (
	"gorm.io/gorm"
)

// Repository is the base repository interface
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB returns the underlying database connection
func (r *Repository) DB() *gorm.DB {
	return r.db
}

