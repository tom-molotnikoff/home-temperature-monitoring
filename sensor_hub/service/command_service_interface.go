package service

import (
	"context"

	gen "example/sensorHub/gen"
)

type CommandServiceInterface interface {
	Send(ctx context.Context, sensorID int, actor *gen.User, property string, value string) (SentCommandResult, error)
}
