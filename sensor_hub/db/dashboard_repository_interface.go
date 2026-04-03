package database

import (
	"context"
	"example/sensorHub/types"
)

type DashboardRepository interface {
	Create(ctx context.Context, dashboard *types.Dashboard) (int, error)
	GetById(ctx context.Context, id int) (*types.Dashboard, error)
	GetByUserId(ctx context.Context, userId int) ([]types.Dashboard, error)
	Update(ctx context.Context, dashboard *types.Dashboard) error
	Delete(ctx context.Context, id int) error
	SetDefault(ctx context.Context, userId int, dashboardId int) error
	GetDefaultForUser(ctx context.Context, userId int) (*types.Dashboard, error)
	CreateCopy(ctx context.Context, sourceDashboardId int, targetUserId int) (int, error)
}
