package repo

import (
	"fmt"

	"github.com/improwised/datautil/pkg/db"
	"github.com/improwised/datautil/pkg/models"
)

func CreateRoute(route *models.RouteConfig) error {
	query := `
		INSERT INTO custom_routes (
			path, source_type, source, filter_expr, select_cols,
			transform_add, transform_remove, transform_rename,
			auth_required, is_dynamic, dynamic_params,
			pagination_enabled, default_limit, max_limit,
			default_format, cache_enabled, cache_ttl, active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.DB.Exec(query,
		route.Path, route.SourceType, route.Source, route.FilterExpr, route.SelectCols,
		route.TransformAdd, route.TransformRemove, route.TransformRename,
		route.AuthRequired, route.IsDynamic, route.DynamicParams,
		route.PaginationEnabled, route.DefaultLimit, route.MaxLimit,
		route.DefaultFormat, route.CacheEnabled, route.CacheTTL, route.Active,
	)
	if err != nil {
		return fmt.Errorf("failed to create route: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	route.ID = uint(id)
	return nil
}

func GetRouteByPath(path string) (*models.RouteConfig, error) {
	query := `SELECT * FROM custom_routes WHERE path = ? AND active = true`

	route := &models.RouteConfig{}
	err := db.DB.QueryRow(query, path).Scan(
		&route.ID, &route.Path, &route.SourceType, &route.Source,
		&route.FilterExpr, &route.SelectCols,
		&route.TransformAdd, &route.TransformRemove, &route.TransformRename,
		&route.AuthRequired, &route.IsDynamic, &route.DynamicParams,
		&route.PaginationEnabled, &route.DefaultLimit, &route.MaxLimit,
		&route.DefaultFormat, &route.CacheEnabled, &route.CacheTTL,
		&route.CreatedAt, &route.UpdatedAt, &route.Active,
	)
	if err != nil {
		return nil, fmt.Errorf("route not found: %w", err)
	}
	return route, nil
}

func GetAllRoutes() ([]models.RouteConfig, error) {
	query := `SELECT * FROM custom_routes WHERE active = true ORDER BY path`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query routes: %w", err)
	}
	defer rows.Close()

	var routes []models.RouteConfig
	for rows.Next() {
		var route models.RouteConfig
		err := rows.Scan(
			&route.ID, &route.Path, &route.SourceType, &route.Source,
			&route.FilterExpr, &route.SelectCols,
			&route.TransformAdd, &route.TransformRemove, &route.TransformRename,
			&route.AuthRequired, &route.IsDynamic, &route.DynamicParams,
			&route.PaginationEnabled, &route.DefaultLimit, &route.MaxLimit,
			&route.DefaultFormat, &route.CacheEnabled, &route.CacheTTL,
			&route.CreatedAt, &route.UpdatedAt, &route.Active,
		)
		if err != nil {
			continue
		}
		routes = append(routes, route)
	}
	return routes, nil
}

func UpdateRoute(route *models.RouteConfig) error {
	query := `
		UPDATE custom_routes SET
			path = ?, source_type = ?, source = ?, filter_expr = ?, select_cols = ?,
			transform_add = ?, transform_remove = ?, transform_rename = ?,
			auth_required = ?, is_dynamic = ?, dynamic_params = ?,
			pagination_enabled = ?, default_limit = ?, max_limit = ?,
			default_format = ?, cache_enabled = ?, cache_ttl = ?, active = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.DB.Exec(query,
		route.Path, route.SourceType, route.Source, route.FilterExpr, route.SelectCols,
		route.TransformAdd, route.TransformRemove, route.TransformRename,
		route.AuthRequired, route.IsDynamic, route.DynamicParams,
		route.PaginationEnabled, route.DefaultLimit, route.MaxLimit,
		route.DefaultFormat, route.CacheEnabled, route.CacheTTL, route.Active,
		route.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update route: %w", err)
	}
	return nil
}

func DeleteRoute(id uint) error {
	query := `UPDATE custom_routes SET active = false, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}
	return nil
}

func GetRouteByID(id uint) (*models.RouteConfig, error) {
	query := `SELECT * FROM custom_routes WHERE id = ?`

	route := &models.RouteConfig{}
	err := db.DB.QueryRow(query, id).Scan(
		&route.ID, &route.Path, &route.SourceType, &route.Source,
		&route.FilterExpr, &route.SelectCols,
		&route.TransformAdd, &route.TransformRemove, &route.TransformRename,
		&route.AuthRequired, &route.IsDynamic, &route.DynamicParams,
		&route.PaginationEnabled, &route.DefaultLimit, &route.MaxLimit,
		&route.DefaultFormat, &route.CacheEnabled, &route.CacheTTL,
		&route.CreatedAt, &route.UpdatedAt, &route.Active,
	)
	if err != nil {
		return nil, fmt.Errorf("route not found: %w", err)
	}
	return route, nil
}
