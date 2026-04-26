package service

import (
	"context"
	gen "example/sensorHub/gen"
)

type DashboardServiceInterface interface {
	ServiceListDashboards(ctx context.Context, userId int) ([]gen.Dashboard, error)
	ServiceGetDashboard(ctx context.Context, id int) (*gen.Dashboard, error)
	ServiceGetDefaultDashboard(ctx context.Context, userId int) (*gen.Dashboard, error)
	ServiceCreateDashboard(ctx context.Context, userId int, req gen.CreateDashboardRequest) (int, error)
	ServiceUpdateDashboard(ctx context.Context, userId int, id int, req gen.UpdateDashboardRequest) error
	ServiceDeleteDashboard(ctx context.Context, userId int, id int) error
	ServiceShareDashboard(ctx context.Context, userId int, dashboardId int, targetUserId int) error
	ServiceSetDefaultDashboard(ctx context.Context, userId int, dashboardId int) error
}
