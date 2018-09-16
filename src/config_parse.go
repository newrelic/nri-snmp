package main

import (
	"fmt"
	"io/ioutil"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/log"
	yaml "gopkg.in/yaml.v2"
)

// collectionDefinitionParser is a struct to aid the automatic
// parsing of a collection yaml file
type collectionDefinitionParser struct {
	Collect []struct {
		Description   string                      `yaml:"description"`
		EventType     string                      `yaml:"event_type"`
		ScalarMetrics []metricDefinitionParser    `yaml:"scalar_metrics"`
		TableMetrics  tableDefinitionParser       `yaml:"table_metrics"`
		Inventory     []inventoryDefinitionParser `yaml:"inventory"`
	}
}

// metricDefinitionParser is a struct to aid the automatic
// parsing of a collection yaml file
type metricDefinitionParser struct {
	Oid        string `yaml:"oid"`
	MetricType string `yaml:"metric_type"`
	MetricName string `yaml:"metric_name"`
}

// inventoryDefinitionParser is a struct to aid the automatic
// parsing of a collection yaml file
type inventoryDefinitionParser struct {
	Oid      string `yaml:"oid"`
	Category string `yaml:"category"`
	Name     string `yaml:"name"`
}

// tableDefinitionParser is a struct to aid the automatic
// parsing of a collection yaml file
type tableDefinitionParser struct {
	RootOid string                   `yaml:"root_oid"`
	Index   []indexDefinitionParser  `yaml:"index"`
	Metrics []metricDefinitionParser `yaml:"metrics"`
}

// indexDefinitionParser is a struct to aid the automatic
// parsing of a collection yaml file
type indexDefinitionParser struct {
	Oid  string `yaml:"oid"`
	Name string `yaml:"name"`
}

// collectionDefinition is a validated and simplified
// representation of the requested collection parameters
// from a single collection
type collectionDefinition struct {
	description             string
	eventType               string
	scalarMetrics           []*metricDefinition
	tableDefinition         tableDefinition
	inventoryItemDefinition []*inventoryItemDefinition
}

// tableDefinition is a validated and simplified
// representation of the requested collection parameters
// from a single table
type tableDefinition struct {
	rootOid           string
	indexDefinitions  []*indexDefinition
	columnDefinitions []*metricDefinition
}

// metricDefinition is a storage struct containing
// the information of a single metric. It can represent
// a scalar metric or a table metric
type metricDefinition struct {
	oid        string
	metricName string
	metricType metric.SourceType
}

// inventoryItemDefinition is a storage struct containing
// the information of a single inventory item
type inventoryItemDefinition struct {
	oid      string
	category string
	name     string
}

// indexDefinition is a storage struct containing
// the information representing a table index
type indexDefinition struct {
	oid  string
	name string
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

// parseYaml reads a yaml file and parses it into a collectionDefinitionParser.
// It validates syntax only and not content
func parseYaml(filename string) (*collectionDefinitionParser, error) {
	// Read the file
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error("failed to open %s: %s", filename, err)
		return nil, err
	}

	// Parse the file
	var c collectionDefinitionParser
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		log.Error("failed to parse collection: %s", err)
		return nil, err
	}

	return &c, nil
}

// parseCollection takes a raw collectionDefinitionParser and returns
// an array of collections containing the validated configuration
func parseCollectionDefinition(c *collectionDefinitionParser) ([]*collectionDefinition, error) {

	// For each collection
	var collections []*collectionDefinition
	for _, parsedCollectionDefinition := range c.Collect {
		//parse event_type
		var eventType string
		if parsedCollectionDefinition.EventType == "" {
			eventType = "SNMPSample"
		} else {
			eventType = parsedCollectionDefinition.EventType
		}

		//parse scalar_metrics if any
		var scalarMetricDefinitions []*metricDefinition
		var newMetricDefinition *metricDefinition
		for _, parsedMetricDefinition := range parsedCollectionDefinition.ScalarMetrics {
			newMetricDefinition = &metricDefinition{
				metricName: parsedMetricDefinition.MetricName,
				oid:        parsedMetricDefinition.Oid,
			}
			metricTypeString := parsedMetricDefinition.MetricType
			if metricTypeString == "" {
				newMetricDefinition.metricType = metric.GAUGE //default metric type
			} else {
				mt, ok := metricTypes[metricTypeString]
				if !ok {
					return nil, fmt.Errorf("invalid metric type %s", metricTypeString)
				}
				newMetricDefinition.metricType = mt
			}
			scalarMetricDefinitions = append(scalarMetricDefinitions, newMetricDefinition)
		}

		//parse inventory metrics if any
		var inventoryItemDefinitions []*inventoryItemDefinition
		var newItem *inventoryItemDefinition
		for _, parsedInventoryItem := range parsedCollectionDefinition.Inventory {
			newItem = &inventoryItemDefinition{
				oid:      parsedInventoryItem.Oid,
				category: parsedInventoryItem.Category,
				name:     parsedInventoryItem.Name,
			}
			inventoryItemDefinitions = append(inventoryItemDefinitions, newItem)
		}

		// parse table_metrics if any
		parsedTableDefinition := parsedCollectionDefinition.TableMetrics
		var indexDefinitions []*indexDefinition
		var newIndex *indexDefinition
		var columnDefinitions []*metricDefinition
		for _, indexOidDef := range parsedTableDefinition.Index {
			newIndex = &indexDefinition{
				name: indexOidDef.Name,
				oid:  indexOidDef.Oid,
			}
			indexDefinitions = append(indexDefinitions, newIndex)
		}
		for _, parsedTableMetricDefinition := range parsedTableDefinition.Metrics {
			newMetricDefinition = &metricDefinition{
				metricName: parsedTableMetricDefinition.MetricName,
				oid:        parsedTableMetricDefinition.Oid,
			}
			metricTypeString := parsedTableMetricDefinition.MetricType
			if metricTypeString == "" {
				newMetricDefinition.metricType = metric.GAUGE
			} else {
				mt, ok := metricTypes[metricTypeString]
				if !ok {
					return nil, fmt.Errorf("invalid metric type %s", metricTypeString)
				}
				newMetricDefinition.metricType = mt
			}
			columnDefinitions = append(columnDefinitions, newMetricDefinition)
		}
		tableDefinition := tableDefinition{
			rootOid:           parsedTableDefinition.RootOid,
			indexDefinitions:  indexDefinitions,
			columnDefinitions: columnDefinitions,
		}
		collections = append(collections,
			&collectionDefinition{
				eventType:               eventType,
				scalarMetrics:           scalarMetricDefinitions,
				tableDefinition:         tableDefinition,
				inventoryItemDefinition: inventoryItemDefinitions,
			})
	}

	return collections, nil
}
