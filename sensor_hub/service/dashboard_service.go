package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	database "example/sensorHub/db"
	"example/sensorHub/types"
	gen "example/sensorHub/gen"
)

type DashboardService struct {
	repo   database.DashboardRepository
	logger *slog.Logger
}

func NewDashboardService(repo database.DashboardRepository, logger *slog.Logger) *DashboardService {
	return &DashboardService{
		repo:   repo,
		logger: logger.With("component", "dashboard_service"),
	}
}

func (s *DashboardService) ServiceListDashboards(ctx context.Context, userId int) ([]gen.Dashboard, error) {
	dashboards, err := s.repo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error listing dashboards: %w", err)
	}
	return dashboards, nil
}

func (s *DashboardService) ServiceGetDashboard(ctx context.Context, id int) (*gen.Dashboard, error) {
	dashboard, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting dashboard: %w", err)
	}
	return dashboard, nil
}

func (s *DashboardService) ServiceGetDefaultDashboard(ctx context.Context, userId int) (*gen.Dashboard, error) {
	dashboard, err := s.repo.GetDefaultForUser(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting default dashboard: %w", err)
	}
	return dashboard, nil
}

func (s *DashboardService) ServiceCreateDashboard(ctx context.Context, userId int, req types.CreateDashboardRequest) (int, error) {
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return 0, fmt.Errorf("error marshalling dashboard config: %w", err)
	}

	dashboard := &gen.Dashboard{
		UserId: userId,
		Name:   req.Name,
		Config: string(configJSON),
	}

	id, err := s.repo.Create(ctx, dashboard)
	if err != nil {
		return 0, fmt.Errorf("error creating dashboard: %w", err)
	}
	s.logger.Info("dashboard created", "id", id, "user_id", userId, "name", req.Name)
	return id, nil
}

func (s *DashboardService) ServiceUpdateDashboard(ctx context.Context, userId int, id int, req types.UpdateDashboardRequest) error {
	existing, err := s.repo.GetById(ctx, id)
	if err != nil {
		return fmt.Errorf("error fetching dashboard for update: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("dashboard not found")
	}
	if existing.UserId != userId {
		return fmt.Errorf("not authorized to update this dashboard")
	}

	if req.Name != "" {
		existing.Name = req.Name
	}

	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return fmt.Errorf("error marshalling dashboard config: %w", err)
	}
	existing.Config = string(configJSON)

	if err := s.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("error updating dashboard: %w", err)
	}
	s.logger.Info("dashboard updated", "id", id, "user_id", userId)
	return nil
}

func (s *DashboardService) ServiceDeleteDashboard(ctx context.Context, userId int, id int) error {
	existing, err := s.repo.GetById(ctx, id)
	if err != nil {
		return fmt.Errorf("error fetching dashboard for delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("dashboard not found")
	}
	if existing.UserId != userId {
		return fmt.Errorf("not authorized to delete this dashboard")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting dashboard: %w", err)
	}
	s.logger.Info("dashboard deleted", "id", id, "user_id", userId)
	return nil
}

func (s *DashboardService) ServiceShareDashboard(ctx context.Context, userId int, dashboardId int, targetUserId int) error {
	existing, err := s.repo.GetById(ctx, dashboardId)
	if err != nil {
		return fmt.Errorf("error fetching dashboard for share: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("dashboard not found")
	}
	if existing.UserId != userId {
		return fmt.Errorf("not authorized to share this dashboard")
	}

	_, err = s.repo.CreateCopy(ctx, dashboardId, targetUserId)
	if err != nil {
		return fmt.Errorf("error sharing dashboard: %w", err)
	}
	s.logger.Info("dashboard shared", "id", dashboardId, "from_user", userId, "to_user", targetUserId)
	return nil
}

func (s *DashboardService) ServiceSetDefaultDashboard(ctx context.Context, userId int, dashboardId int) error {
	existing, err := s.repo.GetById(ctx, dashboardId)
	if err != nil {
		return fmt.Errorf("error fetching dashboard for set-default: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("dashboard not found")
	}
	if existing.UserId != userId {
		return fmt.Errorf("not authorized to set this dashboard as default")
	}

	if err := s.repo.SetDefault(ctx, userId, dashboardId); err != nil {
		return fmt.Errorf("error setting default dashboard: %w", err)
	}
	s.logger.Info("default dashboard set", "dashboard_id", dashboardId, "user_id", userId)
	return nil
}
