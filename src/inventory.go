package main

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

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

	if len(oids) == 0 {
		return nil
	}

	snmpGetResult, err := theSNMP.Get(oids)
	if err != nil {
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
			errorMessage, ok := allerrors[oid]
			if ok {
				return fmt.Errorf("Error Message: %s", errorMessage)
			}
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
