package main

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

func populateScalarMetrics(name string, eventType string, metricDefinitions []*metricDefinition, i *integration.Integration) error {
	// Create an entity for the host
	e, err := i.Entity(targetHost, "host")
	if err != nil {
		return err
	}
	ms := e.NewMetricSet(eventType)
	err = ms.SetMetric("name", name, metric.ATTRIBUTE)
	if err != nil {
		log.Error(err.Error())
	}
	var oids []string
	metricDefinitionMap := make(map[string]*metricDefinition)
	for _, metricDefinition := range metricDefinitions {
		oid := strings.TrimSpace(metricDefinition.oid)
		oids = append(oids, oid)
		metricDefinitionMap[oid] = metricDefinition
		//All scalar OIDs must end with a .0 suffix by convention.
		//But they are not always specified with their .0 suffix in MIBs and elsewhere
		//So be nice and treat an OID and and its variant with .0 suffix as equivalent
		if !strings.HasSuffix(oid, ".0") {
			metricDefinitionMap[oid+".0"] = metricDefinition
		}
	}

	if len(oids) == 0 {
		return nil
	}

	snmpGetResult, err := theSNMP.Get(oids)
	if err != nil {
		return fmt.Errorf("SNMPGet Error %v", err)
	}
	for _, variable := range snmpGetResult.Variables {
		err = processScalarPDU(variable, metricDefinitionMap, ms)
		if err != nil {
			return fmt.Errorf("Error processing %s. %v", variable.Name, err)
		}
	}
	return nil
}

func processScalarPDU(pdu gosnmp.SnmpPDU, metricDefinitionMap map[string]*metricDefinition, ms *metric.Set) error {
	var metricName string
	oid := strings.TrimSpace(pdu.Name)
	metricDefinition, ok := metricDefinitionMap[oid]
	if !ok {
		errorMessage, ok := allerrors[oid]
		if ok {
			return fmt.Errorf("Error Message: %s", errorMessage)
		}
		log.Warn("OID not configured in metricDefinitions and will not be reported[" + oid + "]")
		return nil
	}
	metricName = metricDefinition.metricName
	if metricName == "" {
		metricName = metricDefinition.oid
	}
	return createMetric(metricName, metricDefinition.metricType, pdu, ms)
}
