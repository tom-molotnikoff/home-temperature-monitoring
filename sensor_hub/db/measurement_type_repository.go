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
	query := fmt.Sprintf(`
		SELECT mt.id, mt.name, mt.display_name, mt.category, mt.default_unit,
			COALESCE(mta.function, 'avg') AS default_aggregation_function
		FROM %s mt
		LEFT JOIN measurement_type_aggregations mta ON mta.measurement_type_id = mt.id AND mta.is_default = 1
		ORDER BY mt.name
	`, types.TableMeasurementTypes)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying measurement types: %w", err)
	}
	defer rows.Close()

	var mts []types.MeasurementType
	for rows.Next() {
		var mt types.MeasurementType
		if err := rows.Scan(&mt.Id, &mt.Name, &mt.DisplayName, &mt.Category, &mt.Unit, &mt.DefaultAggregationFunction); err != nil {
			return nil, fmt.Errorf("error scanning measurement type row: %w", err)
		}
		mts = append(mts, mt)
	}
	return mts, rows.Err()
}

func (r *MeasurementTypeRepositoryImpl) GetAllWithReadings(ctx context.Context) ([]types.MeasurementType, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT mt.id, mt.name, mt.display_name, mt.category, mt.default_unit,
			COALESCE(mta.function, 'avg') AS default_aggregation_function
		FROM %s mt
		INNER JOIN %s r ON r.measurement_type_id = mt.id
		LEFT JOIN measurement_type_aggregations mta ON mta.measurement_type_id = mt.id AND mta.is_default = 1
		ORDER BY mt.name
	`, types.TableMeasurementTypes, types.TableReadings)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying measurement types with readings: %w", err)
	}
	defer rows.Close()

	var mts []types.MeasurementType
	for rows.Next() {
		var mt types.MeasurementType
		if err := rows.Scan(&mt.Id, &mt.Name, &mt.DisplayName, &mt.Category, &mt.Unit, &mt.DefaultAggregationFunction); err != nil {
			return nil, fmt.Errorf("error scanning measurement type row: %w", err)
		}
		mts = append(mts, mt)
	}
	return mts, rows.Err()
}

func (r *MeasurementTypeRepositoryImpl) GetByName(ctx context.Context, name string) (*types.MeasurementType, error) {
	query := fmt.Sprintf(`
		SELECT mt.id, mt.name, mt.display_name, mt.category, mt.default_unit,
			COALESCE(mta.function, 'avg') AS default_aggregation_function
		FROM %s mt
		LEFT JOIN measurement_type_aggregations mta ON mta.measurement_type_id = mt.id AND mta.is_default = 1
		WHERE LOWER(mt.name) = LOWER(?)
	`, types.TableMeasurementTypes)
	var mt types.MeasurementType
	err := r.db.QueryRowContext(ctx, query, name).Scan(&mt.Id, &mt.Name, &mt.DisplayName, &mt.Category, &mt.Unit, &mt.DefaultAggregationFunction)
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

func (r *MeasurementTypeRepositoryImpl) GetMeasurementTypesWithReadings(ctx context.Context, sensorId int) ([]types.MeasurementType, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT mt.id, mt.name, mt.display_name, mt.category,
			COALESCE(NULLIF(smt.unit, ''), mt.default_unit) AS unit,
			COALESCE(mta.function, 'avg') AS default_aggregation_function
		FROM %s r
		JOIN %s mt ON r.measurement_type_id = mt.id
		LEFT JOIN %s smt ON smt.sensor_id = r.sensor_id AND smt.measurement_type_id = mt.id
		LEFT JOIN measurement_type_aggregations mta ON mta.measurement_type_id = mt.id AND mta.is_default = 1
		WHERE r.sensor_id = ?
		ORDER BY mt.name
	`, types.TableReadings, types.TableMeasurementTypes, types.TableSensorMeasurementTypes)

	rows, err := r.db.QueryContext(ctx, query, sensorId)
	if err != nil {
		return nil, fmt.Errorf("error querying measurement types with readings: %w", err)
	}
	defer rows.Close()

	var mts []types.MeasurementType
	for rows.Next() {
		var mt types.MeasurementType
		if err := rows.Scan(&mt.Id, &mt.Name, &mt.DisplayName, &mt.Category, &mt.Unit, &mt.DefaultAggregationFunction); err != nil {
			return nil, fmt.Errorf("error scanning measurement type row: %w", err)
		}
		mts = append(mts, mt)
	}
	return mts, rows.Err()
}

func (r *MeasurementTypeRepositoryImpl) GetAggregationsForMeasurementType(ctx context.Context, name string) (*types.MeasurementTypeAggregation, error) {
	query := `
		SELECT mta.function, mta.is_default
		FROM measurement_type_aggregations mta
		JOIN measurement_types mt ON mta.measurement_type_id = mt.id
		WHERE LOWER(mt.name) = LOWER(?)
		ORDER BY mta.is_default DESC, mta.function ASC
	`
	rows, err := r.db.QueryContext(ctx, query, name)
	if err != nil {
		return nil, fmt.Errorf("error fetching aggregations for measurement type %q: %w", name, err)
	}
	defer func() { _ = rows.Close() }()

	result := &types.MeasurementTypeAggregation{
		MeasurementType: name,
	}
	for rows.Next() {
		var fn string
		var isDefault int
		if err := rows.Scan(&fn, &isDefault); err != nil {
			return nil, fmt.Errorf("error scanning aggregation row: %w", err)
		}
		result.SupportedFunctions = append(result.SupportedFunctions, fn)
		if isDefault == 1 {
			result.DefaultFunction = fn
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating aggregation rows: %w", err)
	}
	if len(result.SupportedFunctions) == 0 {
		result.DefaultFunction = "avg"
		result.SupportedFunctions = []string{"avg"}
	}
	return result, nil
}
