package main

import (
	"fmt"
	"strings"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

func populateInventory(inventoryItems []inventoryItem, entity *integration.Entity) error {
	var oids []string
	inventoryOidMap := make(map[string]inventoryItem)
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

	// SNMPv1 will return packet error for unsupported OIDs.
	if snmpGetResult.Error == gosnmp.NoSuchName && theSNMP.Version == gosnmp.Version1 {
		log.Warn("At least one OID not supported by target %s", targetHost)
	}
	// Response received with errors.
	// TODO: "stringify" gosnmp errors instead of showing error code.
	if snmpGetResult.Error != gosnmp.NoError {
		return fmt.Errorf("Error reported by target %s: Error Status %d", targetHost, snmpGetResult.Error)
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
			errorMessage, ok := knownErrorOids[oid]
			if ok {
				return fmt.Errorf("Error Message: %s", errorMessage)
			}
			log.Warn("Unexpected OID %s received", oid)
			continue
		}

		switch variable.Type {
		case gosnmp.OctetString:
			value = string(variable.Value.([]byte))
		case gosnmp.Gauge32, gosnmp.Counter32, gosnmp.Counter64, gosnmp.Integer, gosnmp.Uinteger32:
			value = gosnmp.ToBigInt(variable.Value)
		case gosnmp.ObjectIdentifier, gosnmp.IPAddress:
			if v, ok := variable.Value.(string); ok {
				value = v
			}
			log.Warn("unable to assert type as string for OID ", variable.Name)
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
		} else {
			log.Info("Null value for OID[" + oid + "]")
		}
		if err != nil {
			log.Error("SNMP Error processing inventory variable "+variable.Name, err)
		}
	}
	return nil
}
