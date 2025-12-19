package service

type PropertiesServiceInterface interface {
	ServiceUpdateProperties(properties map[string]string) error
	ServiceGetProperties() (map[string]interface{}, error)
}
