package main

import (
	"fmt"
	"path/filepath"
	"strings"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
	SNMPHost        string `default:"127.0.0.1" help:"Hostname or IP where the SNMP server is running."`
	SNMPPort        int    `default:"161" help:"Port on which SNMP server is listening."`
	Community       string `default:"public" help:"SNMP Version 2 Community string "`
	V3              bool   `default:"false" help:"Use SNMP Version 3."`
	SecurityLevel   string `default:"" help:"Valid values are noAuthnoPriv, authNoPriv or authPriv"`
	Username        string `default:"" help:"The security name that identifies the SNMPv3 user."`
	AuthProtocol    string `default:"SHA" help:"The algorithm used for SNMPv3 authentication (SHA or MD5)."`
	AuthPassphrase  string `default:"" help:"The password used to generate the key used for SNMPv3 authentication."`
	PrivProtocol    string `default:"AES" help:"The algorithm used for SNMPv3 message integrity."`
	PrivPassphrase  string `default:"" help:"The password used to generate the key used to verify SNMPv3 message integrity."`
	CollectionFiles string `default:"" help:"A comma separated list of full paths to metrics configuration files"`
}

const (
	integrationName    = "com.newrelic.snmp"
	integrationVersion = "1.1.3"
)

var (
	args argumentList
)

var theSNMP *gosnmp.GoSNMP
var targetHost string
var targetPort int

func main() {
	// Create Integration
	snmpIntegration, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error(err.Error())
		return
	}

	targetHost = strings.TrimSpace(args.SNMPHost)
	targetPort = args.SNMPPort
	err = connect(targetHost, targetPort)
	if err != nil {
		log.Error("Error connecting to snmp server " + targetHost)
		log.Error(err.Error())
		return
	}
	defer disconnect()

	// Ensure a collection file is specified
	if args.CollectionFiles == "" {
		log.Error("Must specify at least one collection file")
		return
	}

	// For each collection definition file, parse and collect it
	collectionFiles := strings.Split(args.CollectionFiles, ",")
	for _, collectionFile := range collectionFiles {

		// Check that the filepath is an absolute path
		if !filepath.IsAbs(collectionFile) {
			log.Error("invalid metrics collection path %s. Metrics collection files must be specified as absolute paths.", collectionFile)
			return
		}

		// Parse the yaml file into a raw definition
		collectionParser, err := parseYaml(collectionFile)
		if err != nil {
			log.Error("failed to parse collection definition file: " + collectionFile)
			log.Error(err.Error())
			return
		}
		collections, err := parseCollection(collectionParser)
		if err != nil {
			log.Error("failed to parse collection definition: " + collectionFile)
			log.Error(err.Error())
			return
		}

		for _, collection := range collections {
			if err := runCollection(collection, snmpIntegration); err != nil {
				log.Error("failed to complete collection execution")
				log.Error(err.Error())
			}
		}
	}

	if err := snmpIntegration.Publish(); err != nil {
		log.Error(err.Error())
	}
}

func runCollection(collection *collection, i *integration.Integration) error {
	var err error
	// Create an entity for the host
	entity, err := i.Entity(fmt.Sprintf("%s:%d", targetHost, targetPort), "address")
	if err != nil {
		return err
	}

	device := collection.Device
	for _, metricSet := range collection.MetricSets {
		metricSetType := metricSet.Type
		switch metricSetType {
		case "scalar":
			err = populateScalarMetrics(device, metricSet, entity)
			if err != nil {
				log.Error("unable to populate metrics for scalar metric set [%s]. %v", metricSet.Name, err)
				reportError(device, metricSet, entity, err.Error())
			}
		case "table":
			err = populateTableMetrics(device, metricSet, entity)
			if err != nil {
				log.Error("unable to populate metrics for table [%v] %v", metricSet.RootOid, err)
				reportError(device, metricSet, entity, err.Error())
			}
		default:
			log.Error("invalid `metric_set` type: %s. check collection file", metricSetType)
		}
	}
	err = populateInventory(collection.Inventory, entity)
	if err != nil {
		log.Error("unable to populate inventory. %s", err)
	}
	return nil
}

func reportError(device string, metricSet metricSet, entity *integration.Entity, errorMessage string) {
	ms := entity.NewMetricSet(metricSet.EventType)
	err := ms.SetMetric("device", device, metric.ATTRIBUTE)
	if err != nil {
		log.Error(err.Error())
	}
	err = ms.SetMetric("name", metricSet.Name, metric.ATTRIBUTE)
	if err != nil {
		log.Error(err.Error())
	}
	err = ms.SetMetric("errorCode", "SNMPError", metric.ATTRIBUTE)
	if err != nil {
		log.Error(err.Error())
	}
	err = ms.SetMetric("errorMessage", errorMessage, metric.ATTRIBUTE)
	if err != nil {
		log.Error(err.Error())
	}
}
