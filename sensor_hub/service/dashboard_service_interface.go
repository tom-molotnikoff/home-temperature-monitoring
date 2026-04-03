package service

import (
	"context"
	"example/sensorHub/types"
)

type DashboardServiceInterface interface {
	ServiceListDashboards(ctx context.Context, userId int) ([]types.Dashboard, error)
	ServiceGetDashboard(ctx context.Context, id int) (*types.Dashboard, error)
	ServiceGetDefaultDashboard(ctx context.Context, userId int) (*types.Dashboard, error)
	ServiceCreateDashboard(ctx context.Context, userId int, req types.CreateDashboardRequest) (int, error)
	ServiceUpdateDashboard(ctx context.Context, userId int, id int, req types.UpdateDashboardRequest) error
	ServiceDeleteDashboard(ctx context.Context, userId int, id int) error
	ServiceShareDashboard(ctx context.Context, userId int, dashboardId int, targetUserId int) error
	ServiceSetDefaultDashboard(ctx context.Context, userId int, dashboardId int) error
}
