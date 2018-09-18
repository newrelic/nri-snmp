package main

import (
	"regexp"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

func runCollection(metricSetDefinitions []*metricSetDefinition, inventoryDefinitions []*inventoryItemDefinition, i *integration.Integration) error {
	for _, metricSetDefinition := range metricSetDefinitions {
		eventType := metricSetDefinition.EventType
		metricSetType := metricSetDefinition.Type
		switch metricSetType {
		case "scalar":
			populateScalarMetrics(eventType, metricSetDefinition.Metrics, i)
		case "table":
			rootOid := metricSetDefinition.RootOid
			indexDefinitions := metricSetDefinition.Index
			populateTableMetrics(eventType, rootOid, indexDefinitions, metricSetDefinition.Metrics, i)
		default:
			log.Error("Invalid type for metric_set: %s", metricSetType)
		}
	}
	populateInventory(inventoryDefinitions, i)

	return nil
}

func populateScalarMetrics(eventType string, metricDefinitions []*metricDefinition, i *integration.Integration) error {
	// Create an entity for the host
	e, err := i.Entity(targetHost, "host")
	if err != nil {
		return err
	}
	ms := e.NewMetricSet(eventType)
	var oids []string
	metricDefinitionMap := make(map[string]*metricDefinition)
	for _, metricDefinition := range metricDefinitions {
		oid := strings.TrimSpace(metricDefinition.oid)
		oids = append(oids, oid)
		metricDefinitionMap[oid] = metricDefinition
	}
	snmpGetResult, err := theSNMP.Get(oids)
	if err != nil {
		log.Error("SNMP Get Error %s", err)
		return err
	}
	for _, variable := range snmpGetResult.Variables {
		err = processSNMPValue(variable, metricDefinitionMap, ms)
		if err != nil {
			log.Error("SNMP Error processing %s. %s", variable.Name, err)
		}
	}
	return nil
}

func populateTableMetrics(eventType string, rootOid string, indexDefinitions []*indexDefinition, metricDefinitions []*metricDefinition, i *integration.Integration) error {
	var err error
	// Create an entity for the host
	e, err := i.Entity(targetHost, "host")
	if err != nil {
		return err
	}

	indexKeys := make(map[string]struct{}) // "Set" datastructure
	var exists = struct{}{}

	indexAttributeMaps := make(map[string]map[string]string)
	metrics := make(map[string]gosnmp.SnmpPDU)

	snmpWalkCallback := func(pdu gosnmp.SnmpPDU) error {
		oid := strings.TrimSpace(pdu.Name)
		for _, indexDefinition := range indexDefinitions {
			indexKeyPattern := indexDefinition.oid + "\\.(.*)"
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
				indexMap[indexDefinition.name] = indexValue
				return nil
			}
		}
		metrics[oid] = pdu
		return nil
	}
	err = theSNMP.BulkWalk(rootOid, snmpWalkCallback)
	if err != nil {
		log.Error("SNMP Walk Error")
		return err
	}

	for indexKey := range indexKeys {

		indexMap, ok := indexAttributeMaps[indexKey]
		if !ok {
			continue
		}
		ms := e.NewMetricSet(eventType)
		for indexName, indexValue := range indexMap {
			err = ms.SetMetric(indexName, indexValue, metric.ATTRIBUTE)
		}
		if err != nil {
			log.Error(err.Error())
		}
		for _, metricDefinition := range metricDefinitions {
			baseOid := strings.TrimSpace(metricDefinition.oid)
			metricName := metricDefinition.metricName
			sourceType := metricDefinition.metricType
			oid := baseOid + "." + indexKey
			pdu := metrics[oid]
			if metricName == "" {
				metricName = oid
			}
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

func processSNMPValue(pdu gosnmp.SnmpPDU, metricDefinitionMap map[string]*metricDefinition, ms *metric.Set) error {
	var name string
	var sourceType metric.SourceType
	var value interface{}

	oid := strings.TrimSpace(pdu.Name)
	metricDefinition, ok := metricDefinitionMap[oid]
	if ok {
		name = metricDefinition.metricName
		if name == "" {
			name = metricDefinition.oid
		}
		sourceType = metricDefinition.metricType
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
	} else {
		log.Info("Null value for OID[" + oid + "]")
	}

	return nil
}

func populateInventory(inventoryItems []*inventoryItemDefinition, i *integration.Integration) error {
	// Create an entity for the host
	e, err := i.Entity(targetHost, "host")
	if err != nil {
		return err
	}
	var oids []string
	inventoryOidMap := make(map[string]*inventoryItemDefinition)
	for _, inventoryItem := range inventoryItems {
		oid := strings.TrimSpace(inventoryItem.oid)
		oids = append(oids, oid)
		inventoryOidMap[oid] = inventoryItem
	}

	snmpGetResult, err := theSNMP.Get(oids)
	if err != nil {
		log.Error("SNMP Get error fetching inventory items", err)
		return err
	}
	for _, variable := range snmpGetResult.Variables {
		var name string
		var category string
		var value interface{}

		oid := strings.TrimSpace(variable.Name)
		itemDefinition, ok := inventoryOidMap[oid]
		if ok {
			name = itemDefinition.name
			category = itemDefinition.category
		} else {
			log.Error("OID not configured in inventoryDefinitions and will not be reported[" + oid + "]")
			continue
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
			err = e.SetInventoryItem(category, name, value)
			if err != nil {
				log.Error(err.Error())
			}
			if err != nil {
				log.Error(err.Error())
			}
		} else {
			log.Info("Null value for OID[" + oid + "]")
		}
		if err != nil {
			log.Error("SNMP Error processing inventory variable "+variable.Name, err)
		}
	}
	return nil
}
