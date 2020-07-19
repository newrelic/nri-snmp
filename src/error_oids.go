// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

var knownErrorOids = map[string]string{
	".1.3.6.1.6.3.15.1.1.3.0": "oidUsmStatsUnknownUserNames",
	".1.3.6.1.6.3.15.1.1.4.0": "oidUsmStatsUnknownEngineIDs",
	".1.3.6.1.6.3.15.1.1.5.0": "oidUsmStatsWrongDigests",
	".1.3.6.1.6.3.15.1.1.6.0": "oidUsmStatsDecryptionErrors",
}
