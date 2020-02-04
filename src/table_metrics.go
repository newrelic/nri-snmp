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

func populateTableMetrics(device string, metricSet metricSet, entity *integration.Entity) error {
	var err error

	tableRootOid := metricSet.RootOid
	if len(metricSet.Index) == 0 {
		return fmt.Errorf("Table index not specified for table OID `" + tableRootOid + "`")
	}

	metrics := make(map[string]gosnmp.SnmpPDU)
	snmpWalkCallback := func(pdu gosnmp.SnmpPDU) error {
		oid := strings.TrimSpace(pdu.Name)
		errorMessage, ok := knownErrorOids[oid]
		if ok {
			return fmt.Errorf("Error Message: %s", errorMessage)
		}
		metrics[oid] = pdu
		return nil
	}

	err = theSNMP.BulkWalk(tableRootOid, snmpWalkCallback)
	if err != nil {
		return err
	}

	//an `index` uniquely identifies a row in an SNMP table.
	//an `index key` is my term for the OID portion that is appended to the index OID and metric OID to produce SNMP table column data
	//an `index key map` holds column data (as name-value pairs) for a certain row (aka index key)
	//The `index key maps` map the row identifier (aka index key) to its column data (aka index key map)
	indexKeyMaps := make(map[string]map[string]string)
	for _, index := range metricSet.Index {
		//Index OID + "." + Index Key = Index Value
		indexKeyPattern := index.oid + "\\.(.*)"
		re, err := regexp.Compile(indexKeyPattern)
		if err != nil {
			log.Error("unable to compile index key search pattern", err)
			continue
		}
		for oid, pdu := range metrics {
			matches := re.FindStringSubmatch(oid)
			if len(matches) > 1 {
				indexKey := matches[1]
				indexValue, err := extractIndexValue(pdu)
				if err != nil {
					log.Error("unable to extract index value for ", indexKey, err)
					continue
				}
				indexMap, ok := indexKeyMaps[indexKey]
				if !ok {
					indexMap = make(map[string]string)
					indexKeyMaps[indexKey] = indexMap
				}
				indexMap[index.name] = indexValue
			}
		}
	}

	for indexKey, indexNVPairs := range indexKeyMaps {
		ms := entity.NewMetricSet(metricSet.EventType, metric.Attr("IntegrationVersion", integrationVersion))
		err = ms.SetMetric("device", device, metric.ATTRIBUTE)
		if err != nil {
			log.Error(err.Error())
		}
		err = ms.SetMetric("name", metricSet.Name, metric.ATTRIBUTE)
		if err != nil {
			log.Error(err.Error())
		}
		err = ms.SetMetric("index", indexKey, metric.ATTRIBUTE)
		if err != nil {
			log.Error(err.Error())
		}
		for n, v := range indexNVPairs {
			err = ms.SetMetric(n, v, metric.ATTRIBUTE)
			if err != nil {
				log.Error(err.Error())
			}
		}
		for _, metric := range metricSet.Metrics {
			baseOid := strings.TrimSpace(metric.oid)
			metricName := metric.metricName
			oid := baseOid + "." + indexKey
			if pdu, ok := metrics[oid]; ok {
				if metricName == "" {
					metricName = oid
				}
				err = createMetric(metricName, metric.metricType, pdu, ms)
				if err != nil {
					log.Error(err.Error())
				}
			} else {
				log.Warn("No data for " + oid)
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
