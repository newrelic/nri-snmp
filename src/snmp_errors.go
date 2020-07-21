// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import "github.com/soniah/gosnmp"

func getErrorMessage(snmpErr gosnmp.SNMPError) string {
	switch snmpErr {
	case gosnmp.TooBig:
		return "The size of the Response-PDU would be too large to transport"
	case gosnmp.NoSuchName:
		return "The name of a requested object was not found"
	case gosnmp.BadValue:
		return "A value in the request didn't match the structure that the recipient of the request had for the object. For example, an object in the request was specified with an incorrect length or type"
	case gosnmp.ReadOnly:
		return "An attempt was made to set a variable that has an Access value indicating that it is read-only"
	case gosnmp.GenErr:
		return "An error occurred other than one indicated by a more specific error code in this table"
	case gosnmp.NoAccess:
		return "Access was denied to the object for security reasons"
	case gosnmp.WrongType:
		return "The object type in a variable binding is incorrect for the object"
	case gosnmp.WrongLength:
		return "A variable binding specifies a length incorrect for the object"
	case gosnmp.WrongEncoding:
		return "A variable binding specifies an encoding incorrect for the object"
	case gosnmp.WrongValue:
		return "The value given in a variable binding is not possible for the object"
	case gosnmp.NoCreation:
		return "A specified variable does not exist and cannot be created"
	case gosnmp.InconsistentValue:
		return "A variable binding specifies a value that could be held by the variable but cannot be assigned to it at this time"
	case gosnmp.ResourceUnavailable:
		return "An attempt to set a variable required a resource that is not available"
	case gosnmp.CommitFailed:
		return "An attempt to set a particular variable failed"
	case gosnmp.UndoFailed:
		return "An attempt to set a particular variable as part of a group of variables failed, and the attempt to then undo the setting of other variables was not successful"
	case gosnmp.AuthorizationError:
		return "A problem occurred in authorization"
	case gosnmp.NotWritable:
		return "The variable cannot be written or created"
	case gosnmp.InconsistentName:
		return "The name in a variable binding specifies a variable that does not exist"
	}
	return ""
}

func getErrorCode(snmpErr gosnmp.SNMPError) string {
	switch snmpErr {
	case gosnmp.TooBig:
		return "ERR_TooBig"
	case gosnmp.NoSuchName:
		return "ERR_NoSuchName"
	case gosnmp.BadValue:
		return "ERR_BadValue"
	case gosnmp.ReadOnly:
		return "ERR_ReadOnly"
	case gosnmp.GenErr:
		return "ERR_GenErr"
	case gosnmp.NoAccess:
		return "ERR_NoAccess"
	case gosnmp.WrongType:
		return "ERR_WrongType"
	case gosnmp.WrongLength:
		return "ERR_WrongLength"
	case gosnmp.WrongEncoding:
		return "ERR_WrongEncoding"
	case gosnmp.WrongValue:
		return "ERR_WrongValue"
	case gosnmp.NoCreation:
		return "ERR_NoCreation"
	case gosnmp.InconsistentValue:
		return "ERR_InconsistentValue"
	case gosnmp.ResourceUnavailable:
		return "ERR_ResourceUnavailable"
	case gosnmp.CommitFailed:
		return "ERR_CommitFailed"
	case gosnmp.UndoFailed:
		return "ERR_UndoFailed"
	case gosnmp.AuthorizationError:
		return "ERR_AuthorizationError"
	case gosnmp.NotWritable:
		return "ERR_NotWritable"
	case gosnmp.InconsistentName:
		return "ERR_InconsistentName"
	}
	return ""
}
