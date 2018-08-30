package main

import (
	"io"
	"strings"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

func connect() error {
	target := strings.TrimSpace(args.Hostname)

	if args.V3 {
		theSNMP = &gosnmp.GoSNMP{
			Target:        target,
			Port:          uint16(args.Port),
			Version:       gosnmp.Version3,
			Timeout:       time.Duration(30) * time.Second,
			SecurityModel: gosnmp.UserSecurityModel,
			MsgFlags:      gosnmp.AuthPriv,
			SecurityParameters: &gosnmp.UsmSecurityParameters{UserName: args.V3Username,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: args.V3Passphrase,
				PrivacyProtocol:          gosnmp.DES,
				PrivacyPassphrase:        args.V3PrivPassphrase,
			},
		}
	} else {
		community := strings.TrimSpace(args.Community)
		theSNMP = &gosnmp.GoSNMP{
			Target:    target,
			Port:      uint16(args.Port),
			Version:   gosnmp.Version2c,
			Community: community,
			Timeout:   time.Duration(10 * time.Second), // Timeout better suited to walking
			MaxOids:   8900,
		}
	}

	err := theSNMP.Connect()
	if err != nil {
		log.Error("Connect error")
		return err
	}
	log.Info("Connect established to " + target)
	return nil
}

func disconnect() {
	target := strings.TrimSpace(args.Hostname)
	err := theSNMP.Conn.Close()
	if err != nil {
		log.Error("Error disconnecting from SNMP server")
		log.Fatal(err)
	} else {
		log.Debug("Disconnected from " + target)
	}
}

func close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}
