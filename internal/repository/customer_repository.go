package repository

import (
	"xquant-default-management/internal/core"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerRepository interface {
	Create(customer *core.Customer) error
	GetByName(name string) (*core.Customer, error)
	GetByID(id uuid.UUID) (*core.Customer, error)
}

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(customer *core.Customer) error {
	return r.db.Create(customer).Error
}

func (r *customerRepository) GetByName(name string) (*core.Customer, error) {
	var customer core.Customer
	err := r.db.Where("name = ?", name).First(&customer).Error
	return &customer, err
}

func (r *customerRepository) GetByID(id uuid.UUID) (*core.Customer, error) {
	var customer core.Customer
	err := r.db.First(&customer, id).Error
	return &customer, err
}
