// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/soniah/gosnmp"
)

func connect(targetHost string, targetPort int) error {
	if args.V3 {
		// Ensure a collection file is specified
		if args.SecurityLevel == "" {
			return fmt.Errorf("Must specify valid security_level for SNMP v3 (valid values are noAuthnoPriv, authNoPriv and authPriv")
		}

		secLevel := strings.ToLower(strings.TrimSpace(args.SecurityLevel))
		switch secLevel {
		case "noauthnopriv":
			msgFlags := gosnmp.NoAuthNoPriv
			theSNMP = &gosnmp.GoSNMP{
				Target:             targetHost,
				Port:               uint16(targetPort),
				Version:            gosnmp.Version3,
				Timeout:            time.Duration(10) * time.Second,
				SecurityModel:      gosnmp.UserSecurityModel,
				MsgFlags:           msgFlags,
				SecurityParameters: &gosnmp.UsmSecurityParameters{UserName: args.Username},
			}
		case "authnopriv":
			msgFlags := gosnmp.AuthNoPriv
			authProtocolArg := strings.ToUpper(strings.TrimSpace(args.AuthProtocol))

			authProtocol := gosnmp.SHA
			if authProtocolArg == "MD5" {
				authProtocol = gosnmp.MD5
				log.Info("Setting auth_protocol=MD5")
			} else if authProtocolArg == "SHA" {
				authProtocol = gosnmp.SHA
				log.Info("Setting auth_protocol=SHA")
			} else {
				return fmt.Errorf("Must specify valid auth_protocol for SNMP v3 (valid values are SHA or MD5)")
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
				},
			}
		case "authpriv":
			msgFlags := gosnmp.AuthPriv

			authProtocolArg := strings.ToUpper(strings.TrimSpace(args.AuthProtocol))
			authProtocol := gosnmp.SHA
			if authProtocolArg == "MD5" {
				authProtocol = gosnmp.MD5
			} else if authProtocolArg == "SHA" {
				authProtocol = gosnmp.SHA
			} else {
				return fmt.Errorf("Must specify valid auth_protocol for SNMP v3 (valid values are SHA or MD5)")
			}

			privProtocolArg := strings.ToUpper(strings.TrimSpace(args.PrivProtocol))
			privProtocol := gosnmp.AES
			if privProtocolArg == "AES" {
				privProtocol = gosnmp.AES
			} else if privProtocolArg == "DES" {
				privProtocol = gosnmp.DES
			} else {
				return fmt.Errorf("Must specify valid priv_protocol for SNMP v3 (valid values are AES or DES)")
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
		default:
			return fmt.Errorf("Must specify valid security_level for SNMP v3 (valid values are noAuthnoPriv, authNoPriv and authPriv)")
		}

	} else {
		community := strings.TrimSpace(args.Community)
		theSNMP = &gosnmp.GoSNMP{
			Target:    targetHost,
			Port:      uint16(targetPort),
			Version:   gosnmp.Version2c,
			Community: community,
			Timeout:   10 * time.Second, // Timeout better suited to walking
			MaxOids:   8900,
		}
	}

	err := theSNMP.Connect()
	if err != nil {
		log.Error(err.Error())
		return fmt.Errorf("Error connecting to target %s: %s", targetHost, err)
	}
	log.Info("Connecting to target: %v:%p", targetHost, targetPort)
	return nil
}

func disconnect() {
	err := theSNMP.Conn.Close()
	if err != nil {
		log.Warn("Error disconnecting from target %s: %s", targetHost, err)
	}
}
