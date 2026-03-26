package service

import "context"

type PropertiesServiceInterface interface {
	ServiceUpdateProperties(ctx context.Context, properties map[string]string) error
	ServiceGetProperties(ctx context.Context) (map[string]interface{}, error)
}
