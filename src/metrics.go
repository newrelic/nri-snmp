package main

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/soniah/gosnmp"
)

func createMetric(metricName string, metricType metric.SourceType, pdu gosnmp.SnmpPDU, ms *metric.Set) error {
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
		case -1:
			value = gosnmp.ToBigInt(pdu.Value)
			sourceType = metric.GAUGE
		case metric.ATTRIBUTE:
			value = gosnmp.ToBigInt(pdu.Value)
			sourceType = metric.ATTRIBUTE
		default:
			value = gosnmp.ToBigInt(pdu.Value)
			sourceType = metricType
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
		case -1:
			value = float64(pdu.Value.(float32))
			sourceType = metric.GAUGE
		case metric.ATTRIBUTE:
			value = fmt.Sprintf("%f", float64(pdu.Value.(float32)))
			sourceType = metric.ATTRIBUTE
		default:
			value = float64(pdu.Value.(float32))
			sourceType = metricType
		}
		return ms.SetMetric(metricName, value, sourceType)
	case gosnmp.OpaqueDouble:
		switch metricType {
		case -1:
			value = pdu.Value.(float64)
			sourceType = metric.GAUGE
		case metric.ATTRIBUTE:
			value = fmt.Sprintf("%f", pdu.Value.(float64))
			sourceType = metric.ATTRIBUTE
		default:
			value = pdu.Value.(float64)
			sourceType = metricType
		}
		return ms.SetMetric(metricName, value, sourceType)
	case gosnmp.Boolean:
		boolValue := pdu.Value.(bool)
		if boolValue {
			return ms.SetMetric(metricName, 1, sourceType)
		} else {
			return ms.SetMetric(metricName, 0, sourceType)
		}
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
