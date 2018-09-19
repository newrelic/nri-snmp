package main

import (
	"os"
	"strings"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

func connect(targetHost string, targetPort int) error {
	if args.V3 {
		msgFlags := gosnmp.AuthPriv
		authProtocol := gosnmp.MD5
		if args.AuthProtocol == "MD5" {
			authProtocol = gosnmp.MD5
		} else if args.AuthProtocol == "SHA" {
			authProtocol = gosnmp.SHA
		} else {
			log.Error("Invalid auth_protocol. Using MD5")
		}
		privProtocol := gosnmp.AES
		if args.AuthProtocol == "AES" {
			privProtocol = gosnmp.AES
		} else if args.AuthProtocol == "DES" {
			privProtocol = gosnmp.DES
		} else {
			log.Error("Invalid priv_protocol. Using AES")
		}
		if (args.AuthPassphrase != "") && (args.PrivPassphrase != "") {
			msgFlags = gosnmp.AuthPriv
		} else if (args.AuthPassphrase != "") && (args.PrivPassphrase == "") {
			msgFlags = gosnmp.AuthNoPriv
		} else if (args.AuthPassphrase == "") && (args.PrivPassphrase == "") {
			msgFlags = gosnmp.NoAuthNoPriv
		}
		theSNMP = &gosnmp.GoSNMP{
			Target:        targetHost,
			Port:          uint16(targetPort),
			Version:       gosnmp.Version3,
			Timeout:       time.Duration(10) * time.Second,
			SecurityModel: gosnmp.UserSecurityModel,
			MsgFlags:      msgFlags,
			SecurityParameters: &gosnmp.UsmSecurityParameters{UserName: args.Username,
				AuthenticationProtocol:   authProtocol,
				AuthenticationPassphrase: args.AuthPassphrase,
				PrivacyProtocol:          privProtocol,
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
		log.Error("Connect error %v", err)
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
