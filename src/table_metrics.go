package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

func populateTableMetrics(tableName string, eventType string, rootOid string, indexDefinitions []*indexDefinition, metricDefinitions []*metricDefinition, i *integration.Integration) error {
	var err error
	// Create an entity for the host
	e, err := i.Entity(targetHost, "host")
	if err != nil {
		return err
	}
	indexAttributeMaps := make(map[string]map[string]string)
	metrics := make(map[string]gosnmp.SnmpPDU)

	snmpWalkCallback := func(pdu gosnmp.SnmpPDU) error {
		oid := strings.TrimSpace(pdu.Name)
		//fmt.Printf("DEBUG: %s = %v\n", oid, pdu.Value)
		errorMessage, ok := allerrors[oid]
		if ok {
			return fmt.Errorf("Error Message: %s", errorMessage)
		}
		if len(indexDefinitions) == 0 {
			return fmt.Errorf("Table index not specified for table OID `" + rootOid + "`")
		}
		for _, indexDefinition := range indexDefinitions {
			indexKeyPattern := indexDefinition.oid + "\\.(.*)"
			re, err := regexp.Compile(indexKeyPattern)
			if err != nil {
				return err
			}
			matches := re.FindStringSubmatch(oid)
			if len(matches) > 1 {
				indexKey := matches[1]
				indexValue, err := extractIndexValue(pdu)
				if err != nil {
					return err
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
		return err
	}

	for indexKey, indexMap := range indexAttributeMaps {
		ms := e.NewMetricSet(eventType)
		err = ms.SetMetric("table_name", tableName, metric.ATTRIBUTE)
		if err != nil {
			log.Error(err.Error())
		}
		for indexName, indexValue := range indexMap {
			err = ms.SetMetric(indexName, indexValue, metric.ATTRIBUTE)
			if err != nil {
				log.Error(err.Error())
			}
		}
		for _, metricDefinition := range metricDefinitions {
			baseOid := strings.TrimSpace(metricDefinition.oid)
			metricName := metricDefinition.metricName
			oid := baseOid + "." + indexKey
			pdu := metrics[oid]
			if metricName == "" {
				metricName = oid
			}
			err = createMetric(metricName, metricDefinition.metricType, pdu, ms)
			if err != nil {
				log.Error(err.Error())
			}
		}
	}
	return nil
}

func extractIndexValue(pdu gosnmp.SnmpPDU) (string, error) {
	var indexValue string
	switch pdu.Type {
	case gosnmp.OctetString:
		if v, ok := pdu.Value.([]byte); ok {
			indexValue = string(v)
			return indexValue, nil
		}
		return "", fmt.Errorf("unable to assert OctetString as []byte, Oid[" + pdu.Name + "]")
	case gosnmp.Gauge32, gosnmp.Counter32, gosnmp.Counter64, gosnmp.Integer, gosnmp.Uinteger32:
		indexValue = gosnmp.ToBigInt(pdu.Value).String()
		return indexValue, nil
	case gosnmp.ObjectIdentifier, gosnmp.IPAddress:
		if v, ok := pdu.Value.(string); ok {
			indexValue = v
			return indexValue, nil
		}
		return "", fmt.Errorf("unable to assert ObjectIdentifier or IPAddress as string, Oid[" + pdu.Name + "]")
	case gosnmp.Boolean:
		return "", fmt.Errorf("unsupported PDU type[Boolean] for index")
	case gosnmp.BitString:
		return "", fmt.Errorf("unsupported PDU type[Boolean] for index")
	case gosnmp.TimeTicks:
		return "", fmt.Errorf("unsupported PDU type[Boolean] for index")
	case gosnmp.OpaqueFloat:
		return fmt.Sprintf("%f", float64(pdu.Value.(float32))), nil
	case gosnmp.OpaqueDouble:
		return fmt.Sprintf("%f", pdu.Value.(float64)), nil
	case gosnmp.Null:
		return "", fmt.Errorf("null value for table index: [" + pdu.Name + "]")
	case gosnmp.NoSuchObject, gosnmp.NoSuchInstance:
		return "", fmt.Errorf("no such table index: [%v]", pdu.Name)
	default:
		return "", fmt.Errorf("unsupported table index type[%v] for OID[%v]", pdu.Type, pdu.Name)
	}
}
