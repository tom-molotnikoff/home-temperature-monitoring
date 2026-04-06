package database

import (
	"context"
	"database/sql"
	"example/sensorHub/types"
	"fmt"
	"log/slog"
)

type MeasurementTypeRepositoryImpl struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewMeasurementTypeRepository(db *sql.DB, logger *slog.Logger) MeasurementTypeRepository {
	return &MeasurementTypeRepositoryImpl{db: db, logger: logger.With("component", "measurement_type_repository")}
}

func (r *MeasurementTypeRepositoryImpl) GetAll(ctx context.Context) ([]types.MeasurementType, error) {
	query := fmt.Sprintf("SELECT id, name, display_name, category, default_unit FROM %s ORDER BY name", types.TableMeasurementTypes)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying measurement types: %w", err)
	}
	defer rows.Close()

	var mts []types.MeasurementType
	for rows.Next() {
		var mt types.MeasurementType
		if err := rows.Scan(&mt.Id, &mt.Name, &mt.DisplayName, &mt.Category, &mt.Unit); err != nil {
			return nil, fmt.Errorf("error scanning measurement type row: %w", err)
		}
		mts = append(mts, mt)
	}
	return mts, rows.Err()
}

func (r *MeasurementTypeRepositoryImpl) GetByName(ctx context.Context, name string) (*types.MeasurementType, error) {
	query := fmt.Sprintf("SELECT id, name, display_name, category, default_unit FROM %s WHERE LOWER(name) = LOWER(?)", types.TableMeasurementTypes)
	var mt types.MeasurementType
	err := r.db.QueryRowContext(ctx, query, name).Scan(&mt.Id, &mt.Name, &mt.DisplayName, &mt.Category, &mt.Unit)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error querying measurement type by name: %w", err)
	}
	return &mt, nil
}

func (r *MeasurementTypeRepositoryImpl) GetBySensorId(ctx context.Context, sensorId int) ([]types.SensorMeasurementType, error) {
	query := fmt.Sprintf(`
		SELECT smt.sensor_id, smt.measurement_type_id, mt.name, COALESCE(NULLIF(smt.unit, ''), mt.default_unit)
		FROM %s smt
		JOIN %s mt ON smt.measurement_type_id = mt.id
		WHERE smt.sensor_id = ?
		ORDER BY mt.name
	`, types.TableSensorMeasurementTypes, types.TableMeasurementTypes)

	rows, err := r.db.QueryContext(ctx, query, sensorId)
	if err != nil {
		return nil, fmt.Errorf("error querying sensor measurement types: %w", err)
	}
	defer rows.Close()

	var smts []types.SensorMeasurementType
	for rows.Next() {
		var smt types.SensorMeasurementType
		if err := rows.Scan(&smt.SensorId, &smt.MeasurementTypeId, &smt.MeasurementType, &smt.Unit); err != nil {
			return nil, fmt.Errorf("error scanning sensor measurement type row: %w", err)
		}
		smts = append(smts, smt)
	}
	return smts, rows.Err()
}

func (r *MeasurementTypeRepositoryImpl) EnsureExists(ctx context.Context, mt types.MeasurementType) error {
	query := fmt.Sprintf("INSERT OR IGNORE INTO %s (name, display_name, category, default_unit) VALUES (?, ?, ?, ?)", types.TableMeasurementTypes)
	_, err := r.db.ExecContext(ctx, query, mt.Name, mt.DisplayName, mt.Category, mt.Unit)
	if err != nil {
		return fmt.Errorf("error ensuring measurement type exists: %w", err)
	}
	return nil
}

func (r *MeasurementTypeRepositoryImpl) AssignToSensor(ctx context.Context, sensorId, measurementTypeId int, unit string) error {
	query := fmt.Sprintf("INSERT OR IGNORE INTO %s (sensor_id, measurement_type_id, unit) VALUES (?, ?, ?)", types.TableSensorMeasurementTypes)
	_, err := r.db.ExecContext(ctx, query, sensorId, measurementTypeId, unit)
	if err != nil {
		return fmt.Errorf("error assigning measurement type to sensor: %w", err)
	}
	return nil
}

func (r *MeasurementTypeRepositoryImpl) RemoveFromSensor(ctx context.Context, sensorId, measurementTypeId int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE sensor_id = ? AND measurement_type_id = ?", types.TableSensorMeasurementTypes)
	_, err := r.db.ExecContext(ctx, query, sensorId, measurementTypeId)
	if err != nil {
		return fmt.Errorf("error removing measurement type from sensor: %w", err)
	}
	return nil
}
