package service

import (
	"context"
	"example/sensorHub/types"
	gen "example/sensorHub/gen"
)

type DashboardServiceInterface interface {
	ServiceListDashboards(ctx context.Context, userId int) ([]gen.Dashboard, error)
	ServiceGetDashboard(ctx context.Context, id int) (*gen.Dashboard, error)
	ServiceGetDefaultDashboard(ctx context.Context, userId int) (*gen.Dashboard, error)
	ServiceCreateDashboard(ctx context.Context, userId int, req types.CreateDashboardRequest) (int, error)
	ServiceUpdateDashboard(ctx context.Context, userId int, id int, req types.UpdateDashboardRequest) error
	ServiceDeleteDashboard(ctx context.Context, userId int, id int) error
	ServiceShareDashboard(ctx context.Context, userId int, dashboardId int, targetUserId int) error
	ServiceSetDefaultDashboard(ctx context.Context, userId int, dashboardId int) error
}
