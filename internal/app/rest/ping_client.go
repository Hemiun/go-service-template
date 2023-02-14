package rest

import (
	"context"
	"encoding/json"
	"os"

	"go-service-template/internal/app/infrastructure"

	"go-service-template/internal/app/dto"
	//"github.com/go-resty/resty/v2"
	httpClient "go-service-template/internal/app/infrastructure/http"
)

// PingClient иллюстрирует реализацию вызовов других сервисов (http)
// Для каждого ресурса(в данном случае ping)  рекомендуется реализовать отдельный клиент
type PingClient struct {
	infrastructure.SugarLogger
	address string
}

// NewPingClient - конструктор для PingClient
func NewPingClient(address string) *PingClient {
	var target PingClient
	target.address = address
	return &target
}

// Ping вызов ping
func (c *PingClient) Ping(ctx context.Context) (*dto.Ping, error) {
	// Получаем настроенный экземпляр клиента resty
	client := httpClient.GetBaseHTTPClient()

	// Подготавливаем и исполняем запрос
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetContext(ctx).
		Get(c.address)
	if os.IsTimeout(err) {
		c.LogError(ctx, "can't execute http query. client timeout occurred", err)
		return nil, httpClient.ErrHTTPRequestTimeout
	}
	if err != nil {
		c.LogError(ctx, "can't execute request", err)
		return nil, err
	}

	// Проверяем, что запрос выполнен успешно
	if !resp.IsSuccess() {
		return nil, httpClient.ErrHTTPRequestFailed
	}

	// Пытаемся обработать ответ
	var res dto.Ping
	err = json.Unmarshal(resp.Body(), &res)

	if err != nil {
		c.LogError(ctx, "can't unmarshall result data. bad content ", err)
		return nil, httpClient.ErrHTTPBadContent
	}

	return &res, nil
}
