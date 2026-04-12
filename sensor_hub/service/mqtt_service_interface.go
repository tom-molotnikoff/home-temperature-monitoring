package service

import (
	"context"
	"example/sensorHub/types"
)

type MQTTServiceInterface interface {
	// Broker operations
	AddBroker(ctx context.Context, broker types.MQTTBroker) (int, error)
	GetBrokerByID(ctx context.Context, id int) (*types.MQTTBroker, error)
	GetBrokerByName(ctx context.Context, name string) (*types.MQTTBroker, error)
	GetAllBrokers(ctx context.Context) ([]types.MQTTBroker, error)
	GetEnabledBrokers(ctx context.Context) ([]types.MQTTBroker, error)
	UpdateBroker(ctx context.Context, broker types.MQTTBroker) error
	DeleteBroker(ctx context.Context, id int) error

	// Subscription operations
	AddSubscription(ctx context.Context, sub types.MQTTSubscription) (int, error)
	GetSubscriptionByID(ctx context.Context, id int) (*types.MQTTSubscription, error)
	GetAllSubscriptions(ctx context.Context) ([]types.MQTTSubscription, error)
	GetSubscriptionsByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error)
	GetEnabledSubscriptionsByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error)
	UpdateSubscription(ctx context.Context, sub types.MQTTSubscription) error
	DeleteSubscription(ctx context.Context, id int) error
}
