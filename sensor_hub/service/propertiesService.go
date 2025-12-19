package service

import (
	appProps "example/sensorHub/application_properties"
	"log"
)

type PropertiesService struct{}

func NewPropertiesService() *PropertiesService {
	return &PropertiesService{}
}

func (ps *PropertiesService) ServiceUpdateProperties(properties map[string]string) error {
	appProperties, smtpProperties, dbProperties := appProps.ConvertConfigurationToMaps(appProps.AppConfig)

	for key, value := range properties {
		if _, ok := appProperties[key]; ok {
			appProperties[key] = value
		} else if _, ok := dbProperties[key]; ok {
			dbProperties[key] = value
		} else if _, ok := smtpProperties[key]; ok {
			smtpProperties[key] = value
		}
	}

	_, err := appProps.LoadConfigurationFromMaps(appProperties, smtpProperties, dbProperties)
	if err != nil {
		return err
	}

	appProps.ReloadConfig(appProperties, smtpProperties, dbProperties)

	go func() {
		err = appProps.SaveConfigurationToFiles()
		if err != nil {
			log.Printf("Error saving configuration to files, configuration will not be saved on restart: %v", err)
		}
	}()

	return nil
}

func (ps *PropertiesService) ServiceGetProperties() (map[string]interface{}, error) {
	propertiesMap := make(map[string]interface{})

	appProperties, dbProperties, smtpProperties := appProps.ConvertConfigurationToMaps(appProps.AppConfig)

	for key, value := range appProperties {
		propertiesMap[key] = value
	}
	for key, value := range dbProperties {
		propertiesMap[key] = value
	}
	for key, value := range smtpProperties {
		propertiesMap[key] = value
	}

	return propertiesMap, nil
}
