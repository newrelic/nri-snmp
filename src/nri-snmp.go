package main

import (
	"os"
	"strconv"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	Hostname         string `default:"localhost" help:"Hostname or IP where the SNMP server is running."`
	Port             int    `default:"161" help:"Port on which SNMP server is listening."`
	Community        string `default:"public" help:"SNMP Version 2 Community string "`
	V3               bool   `default:"false" help:"Use SNMP Version 3."`
	V3Username       string `default:"" help:"SNMP Version 3 Authentication Username."`
	V3Passphrase     string `default:"" help:"SNMP Version 3 Authentication Passphrase."`
	V3PrivPassphrase string `default:"" help:"SNMP Version 3 Privacy Passphrase."`
	ConfigFile       string `default:"snmp-queries.json" help:"Configuration file containing SNMP oids to query"`
}

const (
	integrationName    = "com.newrelic.nri-snmp"
	integrationVersion = "0.1.0"
)

var (
	args argumentList
)

var theSNMP *gosnmp.GoSNMP

func main() {
	// Create Integration
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	snmpHost := args.Hostname
	// Create Entity, entities name must be unique
	e1, err := i.Entity(snmpHost, "custom")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	msDefinitions, err := loadConfiguration(args.ConfigFile)
	if err != nil {
		log.Error("Error loading configuration from file " + args.ConfigFile)
		log.Fatal(err)
	}

	err = connect()
	if err != nil {
		log.Error("Error connecting to snmp server " + args.Hostname)
		log.Fatal(err)
	}
	defer disconnect()

	// Add Inventory item
	if args.All() || args.Inventory {
		for _, msDefinition := range msDefinitions {
			if len(msDefinition.InventoryDefinitions) == 0 {
				continue
			}
			err = populateInventory(e1, msDefinition)
			if err != nil {
				log.Error(err.Error())
				os.Exit(1)
			}
		}
	}

	// Add Metric
	if args.All() || args.Metrics {
		for k, msDefinition := range msDefinitions {
			if len(msDefinition.MetricDefinitions) == 0 {
				log.Warn("No metrics configured for " + msDefinition.Name + "[" + strconv.Itoa(k) + "]")
				continue
			}
			err = populateMetrics(e1, msDefinition)
			if err != nil {
				log.Error(err.Error())
				os.Exit(1)
			}
		}
	}

	if err = i.Publish(); err != nil {
		log.Error(err.Error())
	}
}
