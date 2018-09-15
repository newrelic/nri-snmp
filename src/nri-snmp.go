package main

import (
	"os"
	"path/filepath"
	"strings"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	Hostname        string `default:"localhost" help:"Hostname or IP where the SNMP server is running."`
	Port            int    `default:"161" help:"Port on which SNMP server is listening."`
	Community       string `default:"public" help:"SNMP Version 2 Community string "`
	V3              bool   `default:"false" help:"Use SNMP Version 3."`
	Username        string `default:"" help:"The security name that identifies the SNMPv3 user."`
	AuthPassphrase  string `default:"" help:"The password used to generate the key used for SNMPv3 authentication."`
	PrivPassphrase  string `default:"" help:"The password used to generate the key used to verify SNMPv3 message integrity."`
	CollectionFiles string `default:"" help:"A comma separated list of full paths to metrics configuration files"`
}

const (
	integrationName    = "com.newrelic.nri-snmp"
	integrationVersion = "1.0.0"
)

var (
	args argumentList
)

var theSNMP *gosnmp.GoSNMP

func main() {
	// Create Integration
	snmpIntegration, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	err = connect()
	if err != nil {
		log.Error("Error connecting to snmp server " + args.Hostname)
		log.Fatal(err)
	}
	defer disconnect()

	// Ensure a collection file is specified
	if args.CollectionFiles == "" {
		log.Error("Must specify at least one collection file")
		os.Exit(1)
	}

	// For each collection definition file, parse and collect it
	collectionFiles := strings.Split(args.CollectionFiles, ",")
	for _, collectionFile := range collectionFiles {

		// Check that the filepath is an absolute path
		if !filepath.IsAbs(collectionFile) {
			log.Error("Invalid metrics collection path %s. Metrics collection files must be specified as absolute paths.", collectionFile)
			os.Exit(1)
		}

		// Parse the yaml file into a raw definition
		collectionDefinition, err := parseYaml(collectionFile)
		if err != nil {
			log.Error("Failed to parse collection definition file %s: %s", collectionFile, err)
			os.Exit(1)
		}

		// Validate the definition and create a collection object
		collection, err := parseCollectionDefinition(collectionDefinition)
		if err != nil {
			log.Error("Failed to parse collection definition %s: %s", collectionFile, err)
			os.Exit(1)
		}
		if err := runCollection(collection, snmpIntegration); err != nil {
			log.Error("Failed to complete collection: %s", err)
		}
	}

	if err := snmpIntegration.Publish(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
