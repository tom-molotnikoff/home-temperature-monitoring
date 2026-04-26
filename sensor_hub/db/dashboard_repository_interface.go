package database

import (
	"context"
	gen "example/sensorHub/gen"
)

type DashboardRepository interface {
	Create(ctx context.Context, dashboard *gen.Dashboard) (int, error)
	GetById(ctx context.Context, id int) (*gen.Dashboard, error)
	GetByUserId(ctx context.Context, userId int) ([]gen.Dashboard, error)
	Update(ctx context.Context, dashboard *gen.Dashboard) error
	Delete(ctx context.Context, id int) error
	SetDefault(ctx context.Context, userId int, dashboardId int) error
	GetDefaultForUser(ctx context.Context, userId int) (*gen.Dashboard, error)
	CreateCopy(ctx context.Context, sourceDashboardId int, targetUserId int) (int, error)
}
