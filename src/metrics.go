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
	sourcetype string
}

func populateMetrics(entity *integration.Entity, msDefinition metricSetDefinition) error {
	tableOid := msDefinition.TableRoot
	if tableOid == "" {
		err := populateScalarMetrics(entity, msDefinition)
		if err != nil {
			return err
		}
	} else {
		err := populateTableMetrics(entity, msDefinition)
		if err != nil {
			return err
		}
	}

	return nil
}

func populateScalarMetrics(entity *integration.Entity, msDefinition metricSetDefinition) error {
	eventType := msDefinition.EventType
	metricDefinitions := msDefinition.MetricDefinitions
	ms := entity.NewMetricSet(eventType)
	var oids []string
	oidDefMap := make(map[string]metricDef)
	for metricName, metricDefinition := range metricDefinitions {
		oid := strings.TrimSpace(metricDefinition[0])
		metricSourceType := strings.TrimSpace(metricDefinition[1])
		oids = append(oids, oid)
		oidDefMap[oid] = metricDef{name: metricName, sourcetype: metricSourceType}
	}
	snmpGetResult, err := theSNMP.Get(oids)
	if err != nil {
		log.Error("SNMP Get Error", err)
		return err
	}
	for _, variable := range snmpGetResult.Variables {
		err = processSNMPValue(variable, oidDefMap, ms)
		if err != nil {
			log.Error("SNMP Error processing ", variable.Name, err)
		}
	}
	return nil
}

func populateTableMetrics(entity *integration.Entity, msDefinition metricSetDefinition) error {
	tableOid := msDefinition.TableRoot
	index := msDefinition.Index
	var indexName string
	var indexOid string
	for indexK, indexV := range index {
		indexName = indexK
		indexOid = indexV
		break
	}
	metricDefs := msDefinition.MetricDefinitions

	indexKeys := make([]string, 0, 0)
	indexKeyValueMapper := make(map[string]string)
	metrics := make(map[string]gosnmp.SnmpPDU)

	snmpWalkCallback := func(pdu gosnmp.SnmpPDU) error {
		oid := strings.TrimSpace(pdu.Name)
		indexKeyPattern := indexOid + "\\.(.*)"
		re, err := regexp.Compile(indexKeyPattern)
		if err != nil {
			return err
		}
		matches := re.FindStringSubmatch(oid)
		if len(matches) > 1 {
			indexKey := matches[1]
			indexKeys = append(indexKeys, indexKey)
			indexValue := "unknown"
			switch pdu.Type {
			case gosnmp.OctetString:
				indexValue = string(pdu.Value.([]byte))
			case gosnmp.Gauge32, gosnmp.Counter32:
				indexValue = gosnmp.ToBigInt(pdu.Value).String()
			default:
				log.Error("Unsupported index value type")
			}
			indexKeyValueMapper[indexKey] = indexValue
			return nil
		}
		metrics[oid] = pdu
		return nil
	}
	err := theSNMP.BulkWalk(tableOid, snmpWalkCallback)
	if err != nil {
		log.Error("SNMP Walk Error")
		return err
	}

	for _, indexKey := range indexKeys {
		ms := entity.NewMetricSet(msDefinition.EventType)

		indexValue, ok := indexKeyValueMapper[indexKey]
		if !ok {
			continue
		}
		err = ms.SetMetric(indexName, indexValue, metric.ATTRIBUTE)
		if err != nil {
			log.Error(err.Error())
		}
		for name, metricDef := range metricDefs {
			baseOid := strings.TrimSpace(metricDef[0])
			srctype := strings.TrimSpace(metricDef[1])
			oid := baseOid + "." + indexKey
			pdu := metrics[oid]
			var sourceType metric.SourceType
			var value interface{}

			switch pdu.Type {
			case gosnmp.OctetString:
				value = string(pdu.Value.([]byte))
				sourceType = metric.ATTRIBUTE
				log.Error("This plugin will always report OctetString values as ATTRIBUTE source type [" + name + "]")
			case gosnmp.Gauge32, gosnmp.Counter32:
				value = gosnmp.ToBigInt(pdu.Value)
				sourceType = getSourceType(srctype)
				if sourceType == metric.ATTRIBUTE {
					value = gosnmp.ToBigInt(pdu.Value).String()
				}
			default:
				value = pdu.Value
				sourceType = getSourceType(srctype)
				if sourceType == metric.ATTRIBUTE {
					value = gosnmp.ToBigInt(pdu.Value).String()
				}
			}
			err = ms.SetMetric(name, value, sourceType)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}
	return nil
}

func processSNMPValue(pdu gosnmp.SnmpPDU, oidDefMap map[string]metricDef, ms *metric.Set) error {
	var name string
	var srctype string
	var sourceType metric.SourceType
	var value interface{}

	oid := strings.TrimSpace(pdu.Name)
	oidDef, ok := oidDefMap[oid]
	if ok {
		name = oidDef.name
		srctype = oidDef.sourcetype
	} else {
		log.Error("OID not configured in metricDefinitions and will not be reported[" + oid + "]")
		return nil
	}

	switch pdu.Type {
	case gosnmp.OctetString:
		value = string(pdu.Value.([]byte))
		sourceType = metric.ATTRIBUTE
		log.Warn("Plugin will report OctetString values as ATTRIBUTE source type only [" + name + "]")
	case gosnmp.Gauge32, gosnmp.Counter32:
		value = gosnmp.ToBigInt(pdu.Value)
		sourceType = getSourceType(srctype)
		if sourceType == metric.ATTRIBUTE {
			value = gosnmp.ToBigInt(pdu.Value).String()
		}
	default:
		value = pdu.Value
		sourceType = getSourceType(srctype)
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
