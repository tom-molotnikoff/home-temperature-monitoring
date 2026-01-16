package service

import (
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/ws"
	"log"
)

type PropertiesService struct{}

func NewPropertiesService() *PropertiesService {
	return &PropertiesService{}
}

func (ps *PropertiesService) ServiceUpdateProperties(properties map[string]string) error {
	appProperties, smtpProperties, dbProperties := appProps.ConvertConfigurationToMaps(appProps.AppConfig)

	for key, value := range properties {
		if value == "*****" {
			sensitiveKeys := appProps.SensitivePropertiesKeys
			isSensitive := false
			for _, sensitiveKey := range sensitiveKeys {
				if key == sensitiveKey {
					isSensitive = true
					break
				}
			}
			if isSensitive {
				continue
				// unchanged sensitive property, skip updating
			}
		}

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

	go func() {
		properties, err := ps.ServiceGetProperties()
		if err != nil {
			log.Printf("Error fetching updated properties for WebSocket broadcast: %v", err)
			return
		}
		ws.BroadcastToTopic("properties", properties)
	}()

	return nil
}

func (ps *PropertiesService) ServiceGetProperties() (map[string]interface{}, error) {
	propertiesMap := make(map[string]interface{})

	appProperties, smtpProperties, dbProperties := appProps.ConvertConfigurationToMaps(appProps.AppConfig)

	for key, value := range appProperties {
		propertiesMap[key] = value
	}
	for key, value := range dbProperties {
		propertiesMap[key] = value
	}
	for key, value := range smtpProperties {
		propertiesMap[key] = value
	}

	for key := range propertiesMap {
		for _, sensitiveKey := range appProps.SensitivePropertiesKeys {
			if key == sensitiveKey {
				propertiesMap[key] = "*****"
			}
		}
	}

	return propertiesMap, nil
}
