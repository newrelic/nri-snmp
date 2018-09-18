package main

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

func connect(targetHost string, targetPort int) error {
	if args.V3 {
		theSNMP = &gosnmp.GoSNMP{
			Target:        targetHost,
			Port:          uint16(targetPort),
			Version:       gosnmp.Version3,
			Timeout:       time.Duration(30) * time.Second,
			SecurityModel: gosnmp.UserSecurityModel,
			MsgFlags:      gosnmp.AuthPriv,
			SecurityParameters: &gosnmp.UsmSecurityParameters{UserName: args.Username,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: args.AuthPassphrase,
				PrivacyProtocol:          gosnmp.DES,
				PrivacyPassphrase:        args.PrivPassphrase,
			},
		}
	} else {
		community := strings.TrimSpace(args.Community)
		theSNMP = &gosnmp.GoSNMP{
			Target:    targetHost,
			Port:      uint16(targetPort),
			Version:   gosnmp.Version2c,
			Community: community,
			Timeout:   time.Duration(10 * time.Second), // Timeout better suited to walking
			MaxOids:   8900,
		}
	}

	err := theSNMP.Connect()
	if err != nil {
		log.Error("Connect error")
		os.Exit(1)
		return err
	}
	log.Info("SNMP target: " + targetHost)
	return nil
}

func disconnect() {
	err := theSNMP.Conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}
