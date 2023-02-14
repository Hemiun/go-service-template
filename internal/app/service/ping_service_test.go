package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go-service-template/internal/app/dto"
	"go-service-template/internal/app/service/mocks"
)

func TestPingService_Ping(t *testing.T) {
	type args struct {
		wantDBError    bool
		wantKafkaError bool
		badDBValue     bool
		wantKafkaPanic bool
		wantDBPanic    bool
	}

	tests := []struct {
		name string
		args args
		want dto.Ping
	}{
		{
			name: "PingService.Ping Case#1 Positive",
			args: args{
				wantDBError:    false,
				wantKafkaError: false,
				badDBValue:     false,
				wantKafkaPanic: false,
				wantDBPanic:    false,
			},
			want: dto.Ping{Status: "OK", StatusDB: "OK", StatusKafka: "OK", Message: ""},
		},
		{
			name: "PingService.Ping Case#2 DB error",
			args: args{
				wantDBError:    true,
				wantKafkaError: false,
				badDBValue:     false,
				wantKafkaPanic: false,
				wantDBPanic:    false,
			},
			want: dto.Ping{Status: "Failed", StatusDB: "Failed", StatusKafka: "OK", Message: "some db error\n"},
		},
		{
			name: "PingService.Ping Case#3 DB bad result",
			args: args{
				wantDBError:    false,
				wantKafkaError: false,
				badDBValue:     true,
				wantKafkaPanic: false,
				wantDBPanic:    false,
			},
			want: dto.Ping{Status: "Failed", StatusDB: "Failed", StatusKafka: "OK", Message: "bad result\n"},
		},
		{
			name: "PingService.Ping Case#4 Kafka error",
			args: args{
				wantDBError:    false,
				wantKafkaError: true,
				badDBValue:     false,
				wantKafkaPanic: false,
				wantDBPanic:    false,
			},
			want: dto.Ping{Status: "Failed", StatusDB: "OK", StatusKafka: "Failed", Message: "some kafka error\n"},
		},
		{
			name: "PingService.Ping Case#5 Kafka and DB error",
			args: args{
				wantDBError:    true,
				wantKafkaError: true,
				badDBValue:     false,
				wantKafkaPanic: false,
				wantDBPanic:    false,
			},
			want: dto.Ping{Status: "Failed", StatusDB: "Failed", StatusKafka: "Failed", Message: "some db error\nsome kafka error\n"},
		},
		{
			name: "PingService.Ping Case#6 Kafka panic from repository",
			args: args{
				wantDBError:    false,
				wantKafkaError: false,
				badDBValue:     false,
				wantKafkaPanic: true,
				wantDBPanic:    false,
			},
			want: dto.Ping{Status: "Failed", StatusDB: "OK", StatusKafka: "Failed", Message: "kafka panic\n"},
		},
		{
			name: "PingService.Ping Case#7 DB panic from repository",
			args: args{
				wantDBError:    false,
				wantKafkaError: false,
				badDBValue:     false,
				wantKafkaPanic: false,
				wantDBPanic:    true,
			},
			want: dto.Ping{Status: "Failed", StatusDB: "Failed", StatusKafka: "OK", Message: "db panic\n"},
		},
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	pingDBRepository := mocks.NewMockPingRepository(mockCtrl)
	pingKafkaRepository := mocks.NewMockPingRepository(mockCtrl)
	txHelper := mocks.NewMockTxHelper(mockCtrl)
	// successDTO := dto.Ping{Status: "OK", StatusDB: "OK", Message: ""}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pingDBRepository.EXPECT().Ping(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, inputVal int) (int, error) {
					var (
						err    error
						resVal int
					)
					if tt.args.wantDBError {
						err = errors.New("some db error")
					}
					if tt.args.badDBValue {
						resVal = inputVal + 10
					} else {
						resVal = inputVal
					}
					if tt.args.wantDBPanic {
						panic("db panic")
					}
					return resVal, err
				})
			pingKafkaRepository.EXPECT().Ping(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, inputVal int) (int, error) {
					var (
						err    error
						resVal int
					)
					if tt.args.wantKafkaError {
						err = errors.New("some kafka error")
					}
					if tt.args.wantKafkaPanic {
						panic("kafka panic")
					}
					resVal = 0
					return resVal, err
				})

			target := NewPingService(pingDBRepository, pingKafkaRepository, txHelper)
			resDTO, err := target.ping(context.Background())
			assert.NoError(t, err, "Unexpected error")
			assert.EqualValues(t, tt.want, resDTO)
		})
	}
}
