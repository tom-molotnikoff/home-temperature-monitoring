package service

import (
	"context"

	gen "example/sensorHub/gen"
)

type CommandServiceInterface interface {
	GetHistory(ctx context.Context, sensorID int) ([]gen.CommandHistoryEntry, error)
	Send(ctx context.Context, sensorID int, actor *gen.User, property string, value string) (SentCommandResult, error)
}
