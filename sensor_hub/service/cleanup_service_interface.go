package service

import "context"

type CleanupServiceInterface interface {
	StartPeriodicCleanup(ctx context.Context)
}
