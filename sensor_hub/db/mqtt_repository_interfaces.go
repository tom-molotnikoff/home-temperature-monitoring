package database

import (
	"context"
	gen "example/sensorHub/gen"
)

type MQTTBrokerRepositoryInterface interface {
	Add(ctx context.Context, broker gen.MQTTBroker) (int, error)
	GetByID(ctx context.Context, id int) (*gen.MQTTBroker, error)
	GetByName(ctx context.Context, name string) (*gen.MQTTBroker, error)
	GetAll(ctx context.Context) ([]gen.MQTTBroker, error)
	GetEnabled(ctx context.Context) ([]gen.MQTTBroker, error)
	Update(ctx context.Context, broker gen.MQTTBroker) error
	Delete(ctx context.Context, id int) error
}

type MQTTSubscriptionRepositoryInterface interface {
	Add(ctx context.Context, sub gen.MQTTSubscription) (int, error)
	GetByID(ctx context.Context, id int) (*gen.MQTTSubscription, error)
	GetAll(ctx context.Context) ([]gen.MQTTSubscription, error)
	GetByBrokerID(ctx context.Context, brokerID int) ([]gen.MQTTSubscription, error)
	GetEnabledByBrokerID(ctx context.Context, brokerID int) ([]gen.MQTTSubscription, error)
	Update(ctx context.Context, sub gen.MQTTSubscription) error
	Delete(ctx context.Context, id int) error
}
