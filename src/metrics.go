package main

import (
	"fmt"

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
			name := metricSetDefinition.Name
			err := populateScalarMetrics(name, eventType, metricSetDefinition.Metrics, i)
			if err != nil {
				log.Error("Error populating scalar metrics. %v", err)
			}
		case "table":
			name := metricSetDefinition.Name
			rootOid := metricSetDefinition.RootOid
			indexDefinitions := metricSetDefinition.Index
			err := populateTableMetrics(name, eventType, rootOid, indexDefinitions, metricSetDefinition.Metrics, i)
			if err != nil {
				log.Error("Error populating table metrics. %v", err)
			}
		default:
			log.Error("Invalid type for metric_set: %s", metricSetType)
		}
	}
	err := populateInventory(inventoryDefinitions, i)
	if err != nil {
		log.Error("Error populating inventory. %s", err)
	}
	return nil
}

func createMetric(metricName string, metricType metricSourceType, pdu gosnmp.SnmpPDU, ms *metric.Set) error {
	var sourceType metric.SourceType
	var value interface{}
	switch pdu.Type {
	case gosnmp.OctetString:
		if v, ok := pdu.Value.([]byte); ok {
			value = string(v)
			return ms.SetMetric(metricName, value, metric.ATTRIBUTE)
		}
	case gosnmp.Gauge32, gosnmp.Counter32, gosnmp.Counter64, gosnmp.Integer, gosnmp.Uinteger32:
		switch metricType {
		case auto, gauge:
			value = gosnmp.ToBigInt(pdu.Value)
			sourceType = metric.GAUGE
		case delta:
			value = gosnmp.ToBigInt(pdu.Value)
			sourceType = metric.DELTA
		case rate:
			value = gosnmp.ToBigInt(pdu.Value)
			sourceType = metric.RATE
		case attribute:
			value = gosnmp.ToBigInt(pdu.Value).String()
			sourceType = metric.ATTRIBUTE
		}
		return ms.SetMetric(metricName, value, sourceType)
	case gosnmp.ObjectIdentifier, gosnmp.IPAddress:
		if v, ok := pdu.Value.(string); ok {
			value = v
			sourceType = metric.ATTRIBUTE
			return ms.SetMetric(metricName, value, sourceType)
		}
		return fmt.Errorf("Unable to assert ObjectIdentifier or IPAddress as string")
	case gosnmp.Boolean:
		return fmt.Errorf("Unsupported PDU type[Boolean]. %v", pdu.Type)
	case gosnmp.BitString:
		return fmt.Errorf("Unsupported PDU type[BitString]. %v", pdu.Type)
	case gosnmp.TimeTicks:
		return fmt.Errorf("Unsupported PDU type[TimeTicks]. %v", pdu.Type)
	case gosnmp.OpaqueFloat, gosnmp.OpaqueDouble:
		return fmt.Errorf("Unsupported PDU type[OpaqueFloat/Double]. %v", pdu.Type)
	case gosnmp.Null:
		return fmt.Errorf("Null value[" + metricName + "].")
	case gosnmp.NoSuchObject, gosnmp.NoSuchInstance:
		return fmt.Errorf("No such object or instance[" + metricName + "].")
	default:
		return fmt.Errorf("Unsupported PDU type[%x] for %v", pdu.Type, metricName)
	}
	return nil
}
