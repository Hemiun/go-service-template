package service

import (
	"context"

	"go-service-template/internal/app/infrastructure/postgres"
)

// InFunc - func type for passing into transaction
type InFunc = postgres.InFunc

// TxFunc - func type for
type TxFunc = postgres.TxFunc

// TxHelper - interface for constructing transactional methods in the service layer
//
//go:generate mockgen -destination=mocks/mock_tx_helper.go -package=mocks . TxHelper
type TxHelper interface {
	// WithTx - factory for transactional methods
	WithTx(ctx context.Context) TxFunc
}
