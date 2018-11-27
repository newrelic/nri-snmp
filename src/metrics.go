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
				log.Error("unable to populate metrics for scalar metric set [%s]. %v", name, err)
			}
		case "table":
			name := metricSetDefinition.Name
			rootOid := metricSetDefinition.RootOid
			indexDefinitions := metricSetDefinition.Index
			err := populateTableMetrics(name, eventType, rootOid, indexDefinitions, metricSetDefinition.Metrics, i)
			if err != nil {
				log.Error("unable to populate metrics for table [%v] %v", rootOid, err)
			}
		default:
			log.Error("invalid `metric_set` type: %s. check collection file", metricSetType)
		}
	}
	err := populateInventory(inventoryDefinitions, i)
	if err != nil {
		log.Error("unable to populate inventory. %s", err)
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
		return fmt.Errorf("unable to assert ObjectIdentifier or IPAddress as string")
	case gosnmp.OpaqueFloat:
		switch metricType {
		case auto, gauge:
			value = float64(pdu.Value.(float32))
			sourceType = metric.GAUGE
		case delta:
			value = float64(pdu.Value.(float32))
			sourceType = metric.DELTA
		case rate:
			value = float64(pdu.Value.(float32))
			sourceType = metric.RATE
		case attribute:
			value = fmt.Sprintf("%f", float64(pdu.Value.(float32)))
			sourceType = metric.ATTRIBUTE
		}
		return ms.SetMetric(metricName, value, sourceType)
	case gosnmp.OpaqueDouble:
		switch metricType {
		case auto, gauge:
			value = pdu.Value.(float64)
			sourceType = metric.GAUGE
		case delta:
			value = pdu.Value.(float64)
			sourceType = metric.DELTA
		case rate:
			value = pdu.Value.(float64)
			sourceType = metric.RATE
		case attribute:
			value = fmt.Sprintf("%f", pdu.Value.(float64))
			sourceType = metric.ATTRIBUTE
		}
		return ms.SetMetric(metricName, value, sourceType)
	case gosnmp.Boolean:
		return fmt.Errorf("unsupported PDU type[Boolean] for %v", metricName)
	case gosnmp.BitString:
		return fmt.Errorf("unsupported PDU type[BitString] for %v", metricName)
	case gosnmp.TimeTicks:
		return fmt.Errorf("unsupported PDU type[TimeTicks] for %v", metricName)
	case gosnmp.UnknownType:
		return fmt.Errorf("unsupported PDU type[UnknownType] for %v", metricName)
	case gosnmp.Null:
		return fmt.Errorf("null value[" + metricName + "].")
	case gosnmp.NoSuchObject, gosnmp.NoSuchInstance:
		return fmt.Errorf("no such object or instance[" + metricName + "].")
	default:
		return fmt.Errorf("unsupported PDU type[%x] for %v", pdu.Type, metricName)
	}
	return nil
}
