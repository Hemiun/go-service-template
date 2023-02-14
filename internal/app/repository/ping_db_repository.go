package repository

import (
	"context"

	"go-service-template/internal/app/infrastructure"
	"go-service-template/internal/app/repository/basedbhandler"
)

// PingDBRepository repository for checking db availability
type PingDBRepository struct {
	infrastructure.SugarLogger
	h basedbhandler.DBHandler
}

// NewPingRepository returns new PingDBRepository
func NewPingRepository(dbHandler basedbhandler.DBHandler) *PingDBRepository {
	var target PingDBRepository
	target.h = dbHandler
	return &target
}

// Ping - check db availability
func (r *PingDBRepository) Ping(ctx context.Context, value int) (int, error) {
	row, err := r.h.QueryRow(ctx, "select $1::int", value)
	if err != nil {
		return 0, err
	}
	var res int
	err = row.Scan(&res)
	if err != nil {
		r.LogError(ctx, "Can't execute ping query", err)
		return 0, err
	}
	return res, nil
}
