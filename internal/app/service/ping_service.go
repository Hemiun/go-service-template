package service

import (
	ctx "context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"go-service-template/internal/app/dto"
	"go-service-template/internal/app/infrastructure"
	"golang.org/x/net/context"
)

// PingRepository interface for repository
//
//go:generate mockgen -destination=mocks/mock_ping_repository.go -package=mocks . PingRepository
type PingRepository interface {
	Ping(ctx context.Context, value int) (int, error)
}

// PingService service for resource availability checking
type PingService struct {
	infrastructure.SugarLogger
	pingDBRepository    PingRepository
	pingKafkaRepository PingRepository
	txHelper            TxHelper
}

const (
	statusFailed = "Failed"
	statusOK     = "OK"
)

// NewPingService returns new PingService
func NewPingService(repoDB PingRepository, repoKafka PingRepository, txHelper TxHelper) *PingService {
	var target PingService
	target.pingDBRepository = repoDB
	target.pingKafkaRepository = repoKafka
	target.txHelper = txHelper
	return &target
}

func (s *PingService) pingDB(ctx ctx.Context, res *dto.Ping) {
	defer func() {
		var err error
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x) //nolint:goerr113
			case error:
				err = x
			}
			s.LogWarn(ctx, "panic recovered").Err(err).Stack()
			res.StatusDB = statusFailed
			res.Message += fmt.Sprintf("%s\n", r)
		}
	}()
	pingValue := rand.Intn(50)
	resValue, err := s.pingDBRepository.Ping(ctx, pingValue)
	if err != nil {
		res.StatusDB = statusFailed
		res.Message += err.Error() + "\n"
	} else if pingValue != resValue {
		res.StatusDB = statusFailed
		res.Message += "bad result" + "\n"
	} else {
		res.StatusDB = statusOK
	}
}

func (s *PingService) pingKafka(ctx ctx.Context, res *dto.Ping) {
	defer func() {
		if r := recover(); r != nil {
			var err error
			switch x := r.(type) {
			case string:
				err = errors.New(x) //nolint:goerr113
			case error:
				err = x
			}
			s.LogWarn(ctx, "panic recovered").Err(err).Stack()
			res.StatusKafka = statusFailed
			res.Message += fmt.Sprintf("%s\n", r)
		}
	}()
	_, err := s.pingKafkaRepository.Ping(ctx, 0)
	if err != nil {
		res.StatusKafka = statusFailed
		res.Message += err.Error() + "\n"
	} else {
		res.StatusKafka = statusOK
	}
}

// Ping check resource availability. Return dto with status
func (s *PingService) Ping(ctx ctx.Context) dto.Ping {
	var res dto.Ping
	txFunc := s.txHelper.WithTx(ctx)
	_ = txFunc(ctx, func(ctx context.Context) error {
		var err error
		res, err = s.ping(ctx)
		return err
	})
	return res
}

func (s *PingService) ping(ctx ctx.Context) (dto.Ping, error) {
	var res dto.Ping
	s.pingDB(ctx, &res)
	s.pingKafka(ctx, &res)

	if (res.StatusDB != "" && res.StatusDB != statusOK) || (res.StatusKafka != "" && res.StatusKafka != statusOK) {
		res.Status = statusFailed
	} else {
		res.Status = statusOK
	}
	return res, nil
}

// PingWithDelay check resources with delay in 1 min
func (s *PingService) PingWithDelay(ctx ctx.Context) dto.Ping {
	var res dto.Ping
	time.Sleep(60 * time.Second)
	s.pingDB(ctx, &res)
	s.pingKafka(ctx, &res)

	if (res.StatusDB != "" && res.StatusDB != statusOK) || (res.StatusKafka != "" && res.StatusKafka != statusOK) {
		res.Status = statusFailed
	} else {
		res.Status = statusOK
	}
	return res
}
