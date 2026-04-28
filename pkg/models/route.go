package models

import (
	"encoding/json"
	"time"
)

type RouteConfig struct {
	ID                uint            `json:"id" db:"id"`
	Path              string          `json:"path" db:"path"`
	SourceType        string          `json:"source_type" db:"source_type"`
	Source            string          `json:"source" db:"source"`
	FilterExpr        string          `json:"filter_expr" db:"filter_expr"`
	SelectCols        string          `json:"select_cols" db:"select_cols"`
	TransformAdd      string          `json:"transform_add" db:"transform_add"`
	TransformRemove   string          `json:"transform_remove" db:"transform_remove"`
	TransformRename   string          `json:"transform_rename" db:"transform_rename"`
	AuthRequired      bool            `json:"auth_required" db:"auth_required"`
	IsDynamic         bool            `json:"is_dynamic" db:"is_dynamic"`
	DynamicParams     json.RawMessage `json:"dynamic_params" db:"dynamic_params"`
	PaginationEnabled bool            `json:"pagination_enabled" db:"pagination_enabled"`
	DefaultLimit      int             `json:"default_limit" db:"default_limit"`
	MaxLimit          int             `json:"max_limit" db:"max_limit"`
	DefaultFormat     string          `json:"default_format" db:"default_format"`
	CacheEnabled      bool            `json:"cache_enabled" db:"cache_enabled"`
	CacheTTL          int             `json:"cache_ttl" db:"cache_ttl"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" db:"updated_at"`
	Active            bool            `json:"active" db:"active"`
}

type RouteRepository interface {
	Create(route *RouteConfig) (*RouteConfig, error)
	GetByID(id uint) (*RouteConfig, error)
	GetByPath(path string) (*RouteConfig, error)
	Update(route *RouteConfig) (*RouteConfig, error)
	Delete(id uint) error
	List() ([]RouteConfig, error)
}

type DynamicParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Default  string `json:"default"`
	Required bool   `json:"required"`
}

func (r *RouteConfig) GetDynamicParams() []DynamicParam {
	if r.DynamicParams == nil {
		return nil
	}
	var params []DynamicParam
	if err := json.Unmarshal(r.DynamicParams, &params); err != nil {
		return nil
	}
	return params
}

func (r *RouteConfig) SetDynamicParams(params []DynamicParam) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	r.DynamicParams = data
	return nil
}

type RouteResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	Message    string      `json:"message,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	TotalRows  int  `json:"total_rows"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

type RouteManager struct {
	Routes map[string]*RouteConfig
	mu     any
}
