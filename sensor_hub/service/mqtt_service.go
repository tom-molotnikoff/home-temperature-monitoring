package service

import (
	"context"
	database "example/sensorHub/db"
	"example/sensorHub/drivers"
	"example/sensorHub/types"
	"fmt"
	"log/slog"
	"strings"
)

type MQTTService struct {
	brokerRepo database.MQTTBrokerRepositoryInterface
	subRepo    database.MQTTSubscriptionRepositoryInterface
	logger     *slog.Logger
}

func NewMQTTService(
	brokerRepo database.MQTTBrokerRepositoryInterface,
	subRepo database.MQTTSubscriptionRepositoryInterface,
	logger *slog.Logger,
) *MQTTService {
	return &MQTTService{
		brokerRepo: brokerRepo,
		subRepo:    subRepo,
		logger:     logger.With("component", "mqtt_service"),
	}
}

// ============================================================================
// Broker operations
// ============================================================================

func (s *MQTTService) AddBroker(ctx context.Context, broker types.MQTTBroker) (int, error) {
	normaliseEmbeddedBroker(&broker)
	if err := validateBroker(broker); err != nil {
		return 0, err
	}
	return s.brokerRepo.Add(ctx, broker)
}

func (s *MQTTService) GetBrokerByID(ctx context.Context, id int) (*types.MQTTBroker, error) {
	return s.brokerRepo.GetByID(ctx, id)
}

func (s *MQTTService) GetBrokerByName(ctx context.Context, name string) (*types.MQTTBroker, error) {
	return s.brokerRepo.GetByName(ctx, name)
}

func (s *MQTTService) GetAllBrokers(ctx context.Context) ([]types.MQTTBroker, error) {
	return s.brokerRepo.GetAll(ctx)
}

func (s *MQTTService) GetEnabledBrokers(ctx context.Context) ([]types.MQTTBroker, error) {
	return s.brokerRepo.GetEnabled(ctx)
}

func (s *MQTTService) UpdateBroker(ctx context.Context, broker types.MQTTBroker) error {
	if broker.Id <= 0 {
		return fmt.Errorf("broker id must be positive")
	}
	normaliseEmbeddedBroker(&broker)
	if err := validateBroker(broker); err != nil {
		return err
	}
	return s.brokerRepo.Update(ctx, broker)
}

func (s *MQTTService) DeleteBroker(ctx context.Context, id int) error {
	return s.brokerRepo.Delete(ctx, id)
}

// ============================================================================
// Subscription operations
// ============================================================================

func (s *MQTTService) AddSubscription(ctx context.Context, sub types.MQTTSubscription) (int, error) {
	if err := s.validateSubscription(ctx, sub); err != nil {
		return 0, err
	}
	return s.subRepo.Add(ctx, sub)
}

func (s *MQTTService) GetSubscriptionByID(ctx context.Context, id int) (*types.MQTTSubscription, error) {
	return s.subRepo.GetByID(ctx, id)
}

func (s *MQTTService) GetAllSubscriptions(ctx context.Context) ([]types.MQTTSubscription, error) {
	return s.subRepo.GetAll(ctx)
}

func (s *MQTTService) GetSubscriptionsByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error) {
	return s.subRepo.GetByBrokerID(ctx, brokerID)
}

func (s *MQTTService) GetEnabledSubscriptionsByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error) {
	return s.subRepo.GetEnabledByBrokerID(ctx, brokerID)
}

func (s *MQTTService) UpdateSubscription(ctx context.Context, sub types.MQTTSubscription) error {
	if sub.Id <= 0 {
		return fmt.Errorf("subscription id must be positive")
	}
	if err := s.validateSubscription(ctx, sub); err != nil {
		return err
	}
	return s.subRepo.Update(ctx, sub)
}

func (s *MQTTService) DeleteSubscription(ctx context.Context, id int) error {
	return s.subRepo.Delete(ctx, id)
}

// ============================================================================
// Validation
// ============================================================================

// normaliseEmbeddedBroker sets host to "localhost" for embedded brokers so
// callers don't need to supply it — the embedded broker always runs locally.
func normaliseEmbeddedBroker(broker *types.MQTTBroker) {
	if broker.Type == "embedded" {
		broker.Host = "localhost"
	}
}

func validateBroker(broker types.MQTTBroker) error {
	if broker.Name == "" {
		return fmt.Errorf("broker name cannot be empty")
	}
	if broker.Host == "" {
		return fmt.Errorf("broker host cannot be empty")
	}
	if broker.Port <= 0 || broker.Port > 65535 {
		return fmt.Errorf("broker port must be between 1 and 65535")
	}
	if broker.Type != "embedded" && broker.Type != "external" {
		return fmt.Errorf("broker type must be 'embedded' or 'external'")
	}
	return nil
}

func (s *MQTTService) validateSubscription(ctx context.Context, sub types.MQTTSubscription) error {
	if sub.TopicPattern == "" {
		return fmt.Errorf("topic pattern cannot be empty")
	}
	if sub.DriverType == "" {
		return fmt.Errorf("driver type cannot be empty")
	}
	if sub.BrokerId <= 0 {
		return fmt.Errorf("broker id must be positive")
	}

	// Validate the driver exists and is a PushDriver
	driver, ok := drivers.Get(sub.DriverType)
	if !ok {
		return fmt.Errorf("unknown driver type: %s", sub.DriverType)
	}
	if _, isPush := driver.(drivers.PushDriver); !isPush {
		return fmt.Errorf("driver %s is not an MQTT push driver", sub.DriverType)
	}

	// Validate the broker exists
	broker, err := s.brokerRepo.GetByID(ctx, sub.BrokerId)
	if err != nil {
		return fmt.Errorf("broker not found: %w", err)
	}
	if broker == nil {
		return fmt.Errorf("broker not found: no broker with id %d", sub.BrokerId)
	}

	// Basic MQTT topic pattern validation
	if err := validateTopicPattern(sub.TopicPattern); err != nil {
		return err
	}

	return nil
}

func validateTopicPattern(pattern string) error {
	if strings.Contains(pattern, " ") {
		return fmt.Errorf("topic pattern must not contain spaces")
	}
	parts := strings.Split(pattern, "/")
	for i, part := range parts {
		if part == "#" && i != len(parts)-1 {
			return fmt.Errorf("multi-level wildcard (#) must be the last segment in topic pattern")
		}
	}
	return nil
}
