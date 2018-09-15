package main

import (
	"regexp"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

type metricDef struct {
	name       string
	sourcetype metric.SourceType
}

func runCollection(collection []*descDefinition, i *integration.Integration) error {
	for _, description := range collection {
		eventType := description.eventType
		scalarMetrics := description.scalarMetrics
		if len(scalarMetrics) > 0 {
			populateScalarMetrics(eventType, scalarMetrics, i)
		}
		tableDefinition := description.tableDefinition
		if len(tableDefinition.metrics) > 0 {
			populateTableMetrics(eventType, tableDefinition, i)
		}
	}
	return nil
}

func populateScalarMetrics(eventType string, metricDefinitions []*attributeRequest, i *integration.Integration) error {
	// Create an entity for the host
	e, err := i.Entity(args.Hostname, "host")
	if err != nil {
		return err
	}
	ms := e.NewMetricSet(eventType)
	var oids []string
	oidDefMap := make(map[string]metricDef)
	for _, metricDefinition := range metricDefinitions {
		oid := strings.TrimSpace(metricDefinition.oid)
		oids = append(oids, oid)
		oidDefMap[oid] = metricDef{name: metricDefinition.metricName, sourcetype: metricDefinition.metricType}
	}
	snmpGetResult, err := theSNMP.Get(oids)
	if err != nil {
		log.Error("SNMP Get Error %s", err)
		return err
	}
	for _, variable := range snmpGetResult.Variables {
		err = processSNMPValue(variable, oidDefMap, ms)
		if err != nil {
			log.Error("SNMP Error processing %s. %s", variable.Name, err)
		}
	}
	return nil
}

func populateTableMetrics(eventType string, tableDefinition tableDefinition, i *integration.Integration) error {
	var err error
	tableOid := tableDefinition.rootOid
	indices := tableDefinition.index
	metricDefinition := tableDefinition.metrics

	indexKeys := make(map[string]struct{}) // "Set" datastructure
	var exists = struct{}{}

	indexAttributeMaps := make(map[string]map[string]string)
	metrics := make(map[string]gosnmp.SnmpPDU)

	snmpWalkCallback := func(pdu gosnmp.SnmpPDU) error {
		oid := strings.TrimSpace(pdu.Name)
		for _, index := range indices {
			indexKeyPattern := index.oid + "\\.(.*)"
			re, err := regexp.Compile(indexKeyPattern)
			if err != nil {
				return err
			}
			matches := re.FindStringSubmatch(oid)
			if len(matches) > 1 {
				indexKey := matches[1]
				indexKeys[indexKey] = exists
				indexValue := "unknown"
				switch pdu.Type {
				case gosnmp.OctetString:
					indexValue = string(pdu.Value.([]byte))
				case gosnmp.Gauge32, gosnmp.Counter32:
					indexValue = gosnmp.ToBigInt(pdu.Value).String()
				default:
					log.Error("Unsupported index value type")
				}
				indexMap, ok := indexAttributeMaps[indexKey]
				if !ok {
					indexMap = make(map[string]string)
					indexAttributeMaps[indexKey] = indexMap
				}
				indexMap[index.name] = indexValue
				return nil
			}
		}
		metrics[oid] = pdu
		return nil
	}
	err = theSNMP.BulkWalk(tableOid, snmpWalkCallback)
	if err != nil {
		log.Error("SNMP Walk Error")
		return err
	}

	for indexKey := range indexKeys {

		indexMap, ok := indexAttributeMaps[indexKey]
		if !ok {
			continue
		}
		// Create an entity for the host
		e, err := i.Entity(args.Hostname, "host")
		if err != nil {
			return err
		}
		ms := e.NewMetricSet(eventType)
		for indexName, indexValue := range indexMap {
			err = ms.SetMetric(indexName, indexValue, metric.ATTRIBUTE)
		}
		if err != nil {
			log.Error(err.Error())
		}
		for _, metricDefinition := range metricDefinition {
			baseOid := strings.TrimSpace(metricDefinition.oid)
			metricName := metricDefinition.metricName
			sourceType := metricDefinition.metricType
			oid := baseOid + "." + indexKey
			pdu := metrics[oid]
			var value interface{}

			switch pdu.Type {
			case gosnmp.OctetString:
				value = string(pdu.Value.([]byte))
				sourceType = metric.ATTRIBUTE
				log.Error("This plugin will always report OctetString values as ATTRIBUTE source type [" + metricName + "]")
			case gosnmp.Gauge32, gosnmp.Counter32:
				value = gosnmp.ToBigInt(pdu.Value)
				if sourceType == metric.ATTRIBUTE {
					value = gosnmp.ToBigInt(pdu.Value).String()
				}
			default:
				value = pdu.Value
				if sourceType == metric.ATTRIBUTE {
					value = gosnmp.ToBigInt(pdu.Value).String()
				}
			}
			err = ms.SetMetric(metricName, value, sourceType)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}
	return nil
}

func processSNMPValue(pdu gosnmp.SnmpPDU, oidDefMap map[string]metricDef, ms *metric.Set) error {
	var name string
	var sourceType metric.SourceType
	var value interface{}

	oid := strings.TrimSpace(pdu.Name)
	oidDef, ok := oidDefMap[oid]
	if ok {
		name = oidDef.name
		sourceType = oidDef.sourcetype
	} else {
		log.Error("OID not configured in metricDefinitions and will not be reported[" + oid + "]")
		return nil
	}

	switch pdu.Type {
	case gosnmp.OctetString:
		value = string(pdu.Value.([]byte))
		sourceType = metric.ATTRIBUTE
	case gosnmp.Gauge32, gosnmp.Counter32:
		value = gosnmp.ToBigInt(pdu.Value)
		if sourceType == metric.ATTRIBUTE {
			value = gosnmp.ToBigInt(pdu.Value).String()
		}
	default:
		value = pdu.Value
		if sourceType == metric.ATTRIBUTE {
			value = gosnmp.ToBigInt(pdu.Value).String()
		}
	}

	if value != nil {
		err := ms.SetMetric(name, value, sourceType)
		if err != nil {
			log.Error(err.Error())
		}
	}

	return nil
}

func getSourceType(srctype string) metric.SourceType {
	var sourceType metric.SourceType
	switch srctype {
	case "gauge", "GAUGE", "Gauge":
		sourceType = metric.GAUGE
	case "attribute", "ATTRIBUTE", "Attribute":
		sourceType = metric.ATTRIBUTE
	case "rate", "RATE", "Rate":
		sourceType = metric.RATE
	case "delta", "DELTA", "Delta":
		sourceType = metric.DELTA
	default:
		sourceType = metric.GAUGE
	}
	return sourceType
}

/*
func populateInventory(entity *integration.Entity, msDefinition metricSetDefinition) error {
	tableOid := msDefinition.TableRoot
	if tableOid == "" {
		categoryDefinitions := msDefinition.InventoryDefinitions
		for category, categoryDefinition := range categoryDefinitions {
			var oids []string
			oidMap := make(map[string]string)
			for metricName, inventoryDefinition := range categoryDefinition {
				oid := strings.TrimSpace(inventoryDefinition[0])
				oids = append(oids, oid)
				oidMap[oid] = metricName
			}
			snmpGetResult, err := theSNMP.Get(oids)
			if err != nil {
				log.Error("SNMP Get Error", err)
				return err
			}
			for _, variable := range snmpGetResult.Variables {
				var name string
				var value interface{}

				oid := strings.TrimSpace(variable.Name)
				metricName, ok := oidMap[oid]
				if ok {
					name = metricName
				} else {
					log.Error("OID not configured in inventoryDefinitions and will not be reported[" + oid + "]")
					return nil
				}

				switch variable.Type {
				case gosnmp.OctetString:
					value = string(variable.Value.([]byte))
				case gosnmp.Gauge32, gosnmp.Counter32:
					value = gosnmp.ToBigInt(variable.Value)
				default:
					value = variable.Value
				}

				if value != nil {
					err = entity.SetInventoryItem(category, name, value)
					if err != nil {
						log.Error(err.Error())
					}
					if err != nil {
						log.Error(err.Error())
					}
				}
				if err != nil {
					log.Error("SNMP Error processing inventory variable "+variable.Name, err)
				}
			}
		}
	}
	return nil
}
*/
