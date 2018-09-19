package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/log"
	yaml "gopkg.in/yaml.v2"
)

// collectionParser is a struct to aid the automatic
// parsing of a collection yaml file
type collectionParser struct {
	Collect []struct {
		DataSet    string            `yaml:"data_set"`
		MetricSets []metricSetParser `yaml:"metric_sets"`
		Inventory  []inventoryParser `yaml:"inventory"`
	}
}

// metricSetParser is a struct to aid the automatic
// parsing of a collection yaml file
type metricSetParser struct {
	Type      string         `yaml:"type"`
	EventType string         `yaml:"event_type"`
	Metrics   []metricParser `yaml:"metrics"`
	RootOid   string         `yaml:"root_oid"`
	Index     []indexParser  `yaml:"index"`
}

// metricParser is a struct to aid the automatic
// parsing of a collection yaml file
type metricParser struct {
	Oid        string `yaml:"oid"`
	MetricType string `yaml:"metric_type"`
	MetricName string `yaml:"metric_name"`
}

// indexParser is a struct to aid the automatic
// parsing of a collection yaml file
type indexParser struct {
	Oid  string `yaml:"oid"`
	Name string `yaml:"name"`
}

// inventoryParser is a struct to aid the automatic
// parsing of a collection yaml file
type inventoryParser struct {
	Oid      string `yaml:"oid"`
	Category string `yaml:"category"`
	Name     string `yaml:"name"`
}

// End of parser defs

// metricSetDefinition is a validated and simplified
// representation of the requested dataset
type metricSetDefinition struct {
	Type      string
	EventType string
	Metrics   []*metricDefinition
	RootOid   string
	Index     []*indexDefinition
}

// metricDefinition is a storage struct containing
// the information of a single metric. It can represent
// a scalar metric or a table metric
type metricDefinition struct {
	oid        string
	metricName string
	metricType metric.SourceType
}

// indexDefinition is a storage struct containing
// the information representing a table index
type indexDefinition struct {
	oid  string
	name string
}

// inventoryItemDefinition is a storage struct containing
// the information of a single inventory item
type inventoryItemDefinition struct {
	oid      string
	category string
	name     string
}

var (
	// metricTypes maps the string used in yaml to a metric type
	metricTypes = map[string]metric.SourceType{
		"gauge":     metric.GAUGE,
		"delta":     metric.DELTA,
		"attribute": metric.ATTRIBUTE,
		"rate":      metric.RATE,
	}
)

// parseYaml reads a yaml file and parses it into a collectionParser.
// It validates syntax only and not content
func parseYaml(filename string) (*collectionParser, error) {
	// Read the file
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error("failed to open %s: %s", filename, err)
		return nil, err
	}
	// Parse the file
	var c collectionParser
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		log.Error("failed to parse collection: %s", err)
		return nil, err
	}
	return &c, nil
}

// parseCollection takes a raw collectionParser and returns
// an slice of metricSetDefinition objects containing the validated configuration
func parseCollection(c *collectionParser) ([]*metricSetDefinition, []*inventoryItemDefinition, error) {
	var metricSetDefinitions []*metricSetDefinition
	var inventoryDefinitions []*inventoryItemDefinition
	for _, dataSet := range c.Collect {
		var newMetricSetDefinition *metricSetDefinition
		for _, metricSetParser := range dataSet.MetricSets {
			eventType := strings.TrimSpace(metricSetParser.EventType)
			metricSetType := strings.TrimSpace(metricSetParser.Type)
			metricParsers := metricSetParser.Metrics
			var metricDefinitions []*metricDefinition
			for _, metricParser := range metricParsers {
				var newMetricDefinition *metricDefinition
				newMetricDefinition = &metricDefinition{
					metricName: metricParser.MetricName,
					oid:        metricParser.Oid,
				}
				metricTypeString := metricParser.MetricType
				if metricTypeString == "" {
					newMetricDefinition.metricType = metric.GAUGE //default metric type
				} else {
					mt, ok := metricTypes[metricTypeString]
					if !ok {
						return nil, nil, fmt.Errorf("invalid metric type %s", metricTypeString)
					}
					newMetricDefinition.metricType = mt
				}
				metricDefinitions = append(metricDefinitions, newMetricDefinition)
			}
			var indexDefinitions []*indexDefinition
			indexParsers := metricSetParser.Index
			for _, indexParser := range indexParsers {
				var newIndexDefinition *indexDefinition
				newIndexDefinition = &indexDefinition{
					name: indexParser.Name,
					oid:  indexParser.Oid,
				}
				indexDefinitions = append(indexDefinitions, newIndexDefinition)
			}
			rootOID := strings.TrimSpace(metricSetParser.RootOid)
			newMetricSetDefinition = &metricSetDefinition{
				Type:      metricSetType,
				EventType: eventType,
				Metrics:   metricDefinitions,
				RootOid:   rootOID,
				Index:     indexDefinitions,
			}
			metricSetDefinitions = append(metricSetDefinitions, newMetricSetDefinition)
		}

		var newInventoryDefinition *inventoryItemDefinition
		for _, inventoryParser := range dataSet.Inventory {
			newInventoryDefinition = &inventoryItemDefinition{
				oid:      inventoryParser.Oid,
				category: inventoryParser.Category,
				name:     inventoryParser.Name,
			}
			inventoryDefinitions = append(inventoryDefinitions, newInventoryDefinition)
		}
	}
	return metricSetDefinitions, inventoryDefinitions, nil
}
