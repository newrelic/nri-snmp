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
		Description   string                   `yaml:"description"`
		EventType     string                   `yaml:"event_type"`
		ScalarMetrics []metricDefinitionParser `yaml:"scalar_metrics"`
		TableMetrics  tableDefinitionParser    `yaml:"table_metrics"`
	}
}

// metricDefinitionParser is a struct to aid the automatic
// parsing of a collection yaml file
type metricDefinitionParser struct {
	Oid        string `yaml:"oid"`
	MetricType string `yaml:"metric_type"`
	MetricName string `yaml:"metric_name"`
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

// descDefinition is a validated and simplified
// representation of the requested collection parameters
// from a single domain
type descDefinition struct {
	description     string
	eventType       string
	scalarMetrics   []*attributeRequest
	tableDefinition tableDefinition
}

// tableDefinition is a validated and simplified
// representation of the requested collection parameters
// from a single domain
type tableDefinition struct {
	rootOid string
	index   []*index
	metrics []*attributeRequest
}

// attributeRequest is a storage struct containing
// the information necessary to turn a OID
// into a metric
type attributeRequest struct {
	oid        string
	metricName string
	metricType metric.SourceType
}

// index is a storage struct containing
// the information necessary to turn a OID
// into a index
type index struct {
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
// an array of domains containing the validated configuration
func parseCollectionDefinition(c *collectionDefinitionParser) ([]*descDefinition, error) {

	// For each metricSet in the collection
	var collections []*descDefinition
	for _, metricSetDef := range c.Collect {
		var scalarMetrics []*attributeRequest
		var newMetric *attributeRequest
		// For each scalar metric in the metricSet
		for _, oidDef := range metricSetDef.ScalarMetrics {
			// Parse the metric and add it to the set
			newMetric = &attributeRequest{metricName: oidDef.MetricName, oid: oidDef.Oid}
			metricTypeString := oidDef.MetricType
			mt, ok := metricTypes[metricTypeString]
			if !ok {
				return nil, fmt.Errorf("invalid metric type %s", metricTypeString)
			}
			newMetric.metricType = mt
			scalarMetrics = append(scalarMetrics, newMetric)
		}

		tableDef := metricSetDef.TableMetrics
		var tableMetrics []*attributeRequest
		// For each table metric in the table column set
		for _, oidDef := range tableDef.Metrics {
			// Parse the metric and add it to the set
			newMetric = &attributeRequest{metricName: oidDef.MetricName, oid: oidDef.Oid}
			metricTypeString := oidDef.MetricType
			mt, ok := metricTypes[metricTypeString]
			if !ok {
				return nil, fmt.Errorf("invalid metric type %s", metricTypeString)
			}
			newMetric.metricType = mt
			tableMetrics = append(tableMetrics, newMetric)
		}

		var tableIndices []*index
		var newIndex *index
		for _, indexOidDef := range tableDef.Index {
			newIndex = &index{name: indexOidDef.Name, oid: indexOidDef.Oid}
			tableIndices = append(tableIndices, newIndex)
		}
		tableDefinition := tableDefinition{rootOid: tableDef.RootOid, index: tableIndices, metrics: tableMetrics}

		var eventType string
		if metricSetDef.EventType == "" {
			eventType = "SNMPSample"
		} else {
			eventType = metricSetDef.EventType
		}
		collections = append(collections,
			&descDefinition{eventType: eventType, scalarMetrics: scalarMetrics},
			&descDefinition{eventType: eventType, tableDefinition: tableDefinition})
	}

	return collections, nil
}
