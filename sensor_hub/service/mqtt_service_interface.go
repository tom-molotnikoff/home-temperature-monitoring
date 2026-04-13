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

	// SetSubscriptionNotifier registers a notifier that is called when
	// subscriptions are added or removed at runtime.
	SetSubscriptionNotifier(notifier SubscriptionNotifier)
}

// SubscriptionNotifier is called by the service layer when subscriptions
// change at runtime. Defined here to avoid circular imports (mqtt→service).
type SubscriptionNotifier interface {
	OnSubscriptionAdded(sub types.MQTTSubscription)
	OnSubscriptionRemoved(sub types.MQTTSubscription)
}
