package service

import (
	"context"
	gen "example/sensorHub/gen"
)

type MQTTServiceInterface interface {
	// Broker operations
	AddBroker(ctx context.Context, broker gen.MQTTBroker) (int, error)
	GetBrokerByID(ctx context.Context, id int) (*gen.MQTTBroker, error)
	GetBrokerByName(ctx context.Context, name string) (*gen.MQTTBroker, error)
	GetAllBrokers(ctx context.Context) ([]gen.MQTTBroker, error)
	GetEnabledBrokers(ctx context.Context) ([]gen.MQTTBroker, error)
	UpdateBroker(ctx context.Context, broker gen.MQTTBroker) error
	DeleteBroker(ctx context.Context, id int) error

	// Subscription operations
	AddSubscription(ctx context.Context, sub gen.MQTTSubscription) (int, error)
	GetSubscriptionByID(ctx context.Context, id int) (*gen.MQTTSubscription, error)
	GetAllSubscriptions(ctx context.Context) ([]gen.MQTTSubscription, error)
	GetSubscriptionsByBrokerID(ctx context.Context, brokerID int) ([]gen.MQTTSubscription, error)
	GetEnabledSubscriptionsByBrokerID(ctx context.Context, brokerID int) ([]gen.MQTTSubscription, error)
	UpdateSubscription(ctx context.Context, sub gen.MQTTSubscription) error
	DeleteSubscription(ctx context.Context, id int) error

	// SetSubscriptionNotifier registers a notifier that is called when
	// subscriptions are added or removed at runtime.
	SetSubscriptionNotifier(notifier SubscriptionNotifier)
}

// SubscriptionNotifier is called by the service layer when subscriptions
// change at runtime. Defined here to avoid circular imports (mqtt→service).
type SubscriptionNotifier interface {
	OnSubscriptionAdded(sub gen.MQTTSubscription)
	OnSubscriptionRemoved(sub gen.MQTTSubscription)
}
