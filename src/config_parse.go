// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

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
		Device     string            `yaml:"device"`
		MetricSets []metricSetParser `yaml:"metric_sets"`
		Inventory  []inventoryParser `yaml:"inventory"`
	}
}

// metricSetParser is a struct to aid the automatic
// parsing of a collection yaml file
type metricSetParser struct {
	Name      string         `yaml:"name"`
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
	Name string `yaml:"metric_name"`
}

// inventoryParser is a struct to aid the automatic
// parsing of a collection yaml file
type inventoryParser struct {
	Oid      string `yaml:"oid"`
	Category string `yaml:"category"`
	Name     string `yaml:"name"`
}

// End of parser defs

// fully parsed and validated collection
type collection struct {
	Device     string
	MetricSets []metricSet
	Inventory  []inventoryItem
}

// metricSet is a validated and simplified
// representation of the requested dataset
type metricSet struct {
	Name      string
	Type      string
	EventType string
	Metrics   []*metricDef
	RootOid   string
	Index     []*index
}

// metricDef is a storage struct containing
// the information of a single metric. It can represent
// a scalar metric or a table metric
type metricDef struct {
	oid        string
	metricName string
	metricType metric.SourceType
}

// index is a storage struct containing
// the information representing a table index
type index struct {
	oid  string
	name string
}

// inventoryItem is a storage struct containing
// the information of a single inventory item
type inventoryItem struct {
	oid      string
	category string
	name     string
}

var (
	// SourcesNameToType maps the string used in yaml to a metric type
	SourcesNameToType = map[string]metric.SourceType{
		"auto":      -1,
		"gauge":     metric.GAUGE,
		"delta":     metric.DELTA,
		"pdelta":    metric.PDELTA,
		"rate":      metric.RATE,
		"prate":     metric.PRATE,
		"attribute": metric.ATTRIBUTE,
	}
)

// parseYaml reads a yaml file and parses it into a collectionParser.
// It validates syntax only and not content
func parseYaml(filename string) (*collectionParser, error) {
	// Read the file
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error("Failed to open %s: %s", filename, err)
		return nil, err
	}
	// Parse the file
	var c collectionParser
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		log.Error("Failed to parse collection: %s", err)
		return nil, err
	}
	return &c, nil
}

// parseCollection takes a raw collectionParser and returns
// an slice of metricSetDefinition objects containing the validated configuration
func parseCollection(c *collectionParser) ([]*collection, error) {
	var cols []*collection
	var metricSets []metricSet
	var inventory []inventoryItem
	for _, dataSet := range c.Collect {
		var newMetricSet metricSet
		for _, metricSetParser := range dataSet.MetricSets {
			name := strings.TrimSpace(metricSetParser.Name)
			eventType := strings.TrimSpace(metricSetParser.EventType)
			metricSetType := strings.TrimSpace(metricSetParser.Type)
			metricParsers := metricSetParser.Metrics
			var metrics []*metricDef
			for _, metricParser := range metricParsers {
				metricOid := strings.TrimSpace(metricParser.Oid)
				//force all oids to start with a leading dot indicating abolute oids as required by gosnmp
				if !strings.HasPrefix(metricOid, ".") {
					metricOid = "." + metricOid
				}
				newMetric := &metricDef{
					metricName: metricParser.MetricName,
					oid:        metricOid,
				}
				metricTypeString := metricParser.MetricType
				if metricTypeString == "" {
					newMetric.metricType = -1
				} else {
					mt, ok := SourcesNameToType[metricTypeString]
					if !ok {
						return nil, fmt.Errorf("invalid metric type %s", metricTypeString)
					}
					newMetric.metricType = mt
				}
				metrics = append(metrics, newMetric)
			}
			var indexes []*index
			indexParsers := metricSetParser.Index
			for _, indexParser := range indexParsers {
				indexOid := strings.TrimSpace(indexParser.Oid)
				//force all oids to start with a leading dot indicating abolute oids as required by gosnmp
				if !strings.HasPrefix(indexOid, ".") {
					indexOid = "." + indexOid
				}
				newIndex := &index{
					name: indexParser.Name,
					oid:  indexOid,
				}
				indexes = append(indexes, newIndex)
			}
			rootOID := strings.TrimSpace(metricSetParser.RootOid)
			newMetricSet = metricSet{
				Name:      name,
				Type:      metricSetType,
				EventType: eventType,
				Metrics:   metrics,
				RootOid:   rootOID,
				Index:     indexes,
			}
			metricSets = append(metricSets, newMetricSet)
		}

		for _, inventoryParser := range dataSet.Inventory {
			newInventoryItem := inventoryItem{
				oid:      inventoryParser.Oid,
				category: inventoryParser.Category,
				name:     inventoryParser.Name,
			}
			inventory = append(inventory, newInventoryItem)
		}
		col := collection{Device: dataSet.Device, MetricSets: metricSets, Inventory: inventory}
		cols = append(cols, &col)
	}
	return cols, nil
}
