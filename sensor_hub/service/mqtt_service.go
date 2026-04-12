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

// maxTopicLength is the MQTT specification limit for topic filters (UTF-8 encoded).
const maxTopicLength = 65535

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
	if broker.Type == "embedded" {
		existing, err := s.brokerRepo.GetAll(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to check existing brokers: %w", err)
		}
		for _, b := range existing {
			if b.Type == "embedded" {
				return 0, fmt.Errorf("an embedded broker already exists (id=%d, name=%q)", b.Id, b.Name)
			}
		}
	}
	if err := validateBroker(broker); err != nil {
		return 0, err
	}
	if err := s.checkBrokerNameUnique(ctx, broker.Name, 0); err != nil {
		return 0, err
	}
	if err := s.checkBrokerHostPortUnique(ctx, broker.Host, broker.Port, 0); err != nil {
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
	if err := s.checkBrokerNameUnique(ctx, broker.Name, broker.Id); err != nil {
		return err
	}
	if err := s.checkBrokerHostPortUnique(ctx, broker.Host, broker.Port, broker.Id); err != nil {
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
	if strings.TrimSpace(broker.Name) == "" {
		return fmt.Errorf("broker name cannot be empty")
	}
	if strings.TrimSpace(broker.Host) == "" {
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

// checkBrokerNameUnique ensures no other broker has the same name (case-insensitive).
// excludeID is the broker being updated (0 for new brokers).
func (s *MQTTService) checkBrokerNameUnique(ctx context.Context, name string, excludeID int) error {
	existing, err := s.brokerRepo.GetByName(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check broker name uniqueness: %w", err)
	}
	if existing != nil && existing.Id != excludeID {
		return fmt.Errorf("broker name %q is already in use (id=%d)", existing.Name, existing.Id)
	}
	return nil
}

// checkBrokerHostPortUnique ensures no other broker targets the same host:port.
// excludeID is the broker being updated (0 for new brokers).
func (s *MQTTService) checkBrokerHostPortUnique(ctx context.Context, host string, port int, excludeID int) error {
	all, err := s.brokerRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to check broker host:port uniqueness: %w", err)
	}
	for _, b := range all {
		if b.Id != excludeID && strings.EqualFold(b.Host, host) && b.Port == port {
			return fmt.Errorf("broker host:port %s:%d is already in use by broker %q (id=%d)", host, port, b.Name, b.Id)
		}
	}
	return nil
}

func (s *MQTTService) validateSubscription(ctx context.Context, sub types.MQTTSubscription) error {
	if strings.TrimSpace(sub.TopicPattern) == "" {
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

	// Check for overlapping subscriptions on the same broker
	if err := s.checkTopicOverlap(ctx, sub.BrokerId, sub.TopicPattern, sub.Id); err != nil {
		return err
	}

	return nil
}

func validateTopicPattern(pattern string) error {
	if strings.Contains(pattern, " ") {
		return fmt.Errorf("topic pattern must not contain spaces")
	}
	if len(pattern) > maxTopicLength {
		return fmt.Errorf("topic pattern exceeds maximum length of %d bytes", maxTopicLength)
	}
	parts := strings.Split(pattern, "/")
	for i, part := range parts {
		if part == "#" && i != len(parts)-1 {
			return fmt.Errorf("multi-level wildcard (#) must be the last segment in topic pattern")
		}
	}
	return nil
}

// checkTopicOverlap verifies that a new or updated subscription does not overlap
// with existing subscriptions on the same broker, which would cause duplicate
// message processing. excludeID is the subscription being updated (0 for new).
func (s *MQTTService) checkTopicOverlap(ctx context.Context, brokerID int, newTopic string, excludeID int) error {
	existing, err := s.subRepo.GetByBrokerID(ctx, brokerID)
	if err != nil {
		return fmt.Errorf("failed to check topic overlap: %w", err)
	}
	for _, sub := range existing {
		if sub.Id == excludeID {
			continue
		}
		if topicsOverlap(sub.TopicPattern, newTopic) {
			return fmt.Errorf("topic pattern %q overlaps with existing subscription %q (id=%d) on this broker; overlapping topics cause duplicate message processing", newTopic, sub.TopicPattern, sub.Id)
		}
	}
	return nil
}

// topicsOverlap returns true if two MQTT topic filters could match any of the
// same concrete topics. It checks both directions: whether pattern A could
// match topics that B also matches, and vice versa.
func topicsOverlap(a, b string) bool {
	return topicCouldMatch(a, b) || topicCouldMatch(b, a)
}

// topicCouldMatch returns true if a message matching concrete segments of
// pattern `sub` could also be delivered to `filter`. This handles MQTT's `+`
// (single-level) and `#` (multi-level) wildcards.
func topicCouldMatch(filter, sub string) bool {
	filterParts := strings.Split(filter, "/")
	subParts := strings.Split(sub, "/")

	for i := 0; i < len(filterParts); i++ {
		if filterParts[i] == "#" {
			return true // # matches everything from here on
		}
		if i >= len(subParts) {
			return false // filter has more segments than sub, no overlap
		}
		if filterParts[i] == "+" || subParts[i] == "+" || subParts[i] == "#" {
			if subParts[i] == "#" {
				return true
			}
			continue // single-level wildcard matches any single segment
		}
		if filterParts[i] != subParts[i] {
			return false // literal segments don't match
		}
	}

	return len(filterParts) == len(subParts)
}
