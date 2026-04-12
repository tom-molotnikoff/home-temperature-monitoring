package database

import (
	"context"
	"example/sensorHub/types"
)

type MQTTBrokerRepositoryInterface interface {
	Add(ctx context.Context, broker types.MQTTBroker) (int, error)
	GetByID(ctx context.Context, id int) (*types.MQTTBroker, error)
	GetByName(ctx context.Context, name string) (*types.MQTTBroker, error)
	GetAll(ctx context.Context) ([]types.MQTTBroker, error)
	GetEnabled(ctx context.Context) ([]types.MQTTBroker, error)
	Update(ctx context.Context, broker types.MQTTBroker) error
	Delete(ctx context.Context, id int) error
}

type MQTTSubscriptionRepositoryInterface interface {
	Add(ctx context.Context, sub types.MQTTSubscription) (int, error)
	GetByID(ctx context.Context, id int) (*types.MQTTSubscription, error)
	GetAll(ctx context.Context) ([]types.MQTTSubscription, error)
	GetByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error)
	GetEnabledByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error)
	Update(ctx context.Context, sub types.MQTTSubscription) error
	Delete(ctx context.Context, id int) error
}
