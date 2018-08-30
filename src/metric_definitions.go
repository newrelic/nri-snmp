package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/newrelic/infra-integrations-sdk/log"
)

type metricSetDefinition struct {
	Name                 string                         `json:"name"`
	EventType            string                         `json:"eventType"`
	TableRoot            string                         `json:"tableRoot"`
	Index                map[string]string              `json:"index"`
	MetricDefinitions    map[string][]string            `json:"metricDefinitions"`
	InventoryDefinitions map[string]map[string][]string `json:"inventoryDefinitions"`
}

func loadConfiguration(file string) ([]metricSetDefinition, error) {
	log.Info("Loading configuration from file, " + file)
	var msDefinitions []metricSetDefinition
	configFile, err := os.Open(file)
	defer close(configFile)
	if err != nil {
		log.Error("Error reading configuration file ", file, err)
		return msDefinitions, err
	}
	byteValue, _ := ioutil.ReadAll(configFile)
	err = json.Unmarshal(byteValue, &msDefinitions)
	if err != nil {
		log.Error("Error unmarshalling configuration data ", file, err)
		return msDefinitions, err
	}
	return msDefinitions, nil
}
