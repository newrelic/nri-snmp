package main

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

func populateScalarMetrics(device string, metricSet metricSet, entity *integration.Entity) error {
	var oids []string
	oidToMetricMap := make(map[string]*metricDef)
	for _, metric := range metricSet.Metrics {
		oid := strings.TrimSpace(metric.oid)
		oids = append(oids, oid)
		oidToMetricMap[oid] = metric
		//All scalar OIDs must end with a .0 suffix by convention.
		//But they are not always specified with their .0 suffix in MIBs and elsewhere
		//So be nice and treat an OID and and its variant with .0 suffix as equivalent
		if !strings.HasSuffix(oid, ".0") {
			oidToMetricMap[oid+".0"] = metric
		}
	}
	if len(oids) == 0 {
		return nil
	}
	if len(oids) > 200 {
		return fmt.Errorf("Metric Set %s has %d metrics, the current limit is 200. This metric set will not be reported", metricSet.Name, len(oids))
	}

	ms := entity.NewMetricSet(metricSet.EventType,
		attribute.Attr("device", device),
		attribute.Attr("name", metricSet.Name))

	snmpGetResult, err := theSNMP.Get(oids)
	if err != nil {
		return err
	}

	// Response received with errors
	if snmpGetResult.Error != gosnmp.NoError {
		err = ms.SetMetric("errorCode", getErrorCode(snmpGetResult.Error), metric.ATTRIBUTE)
		if err != nil {
			log.Error(err.Error())
		}
		err = ms.SetMetric("errorMessage", getErrorMessage(snmpGetResult.Error), metric.ATTRIBUTE)
		if err != nil {
			log.Error(err.Error())
		}
		return nil
	}

	for _, pdu := range snmpGetResult.Variables {
		if pdu.Type == gosnmp.NoSuchObject || pdu.Type == gosnmp.NoSuchInstance {
			log.Warn("OID %s not supported by target %s", pdu.Name, targetHost)
			continue
		}
		oid := strings.TrimSpace(pdu.Name)
		metric, ok := oidToMetricMap[oid]
		if ok {
			metricName := metric.metricName
			if metricName == "" {
				metricName = metric.oid
			}
			err := createMetric(metricName, metric.metricType, pdu, ms)
			if err != nil {
				log.Error(err.Error())
			}
		} else {
			errorMessage, ok := knownErrorOids[oid]
			if ok {
				log.Error(errorMessage)
			} else {
				log.Debug("unexpected OID %s received")
			}
		}
	}
	return nil
}
