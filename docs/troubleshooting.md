# Troubleshooting issues

Working with SNMP can be difficult which leads to issues with the config.
This document should help to debug config issues.

## What do you need?

The following is needed to debug your `nri-snmp` config and other issues.

1. `snmpwalk`
2. `docker`

Optional tools: 

1. [`jq`](https://stedolan.github.io/jq/)

## Connection issues

If you are getting errors connecting `nri-snmp` to your SNMP endpoint then it is recommended to try and use `snmpwalk`.
The trick here is to try and convert your connection config to the write params needed for `snmpwalk`.
For example if your config is as follows:

```yaml
---
integrations:
  - name: nri-snmp
    env:
      # Use the discovered IP as the host address
      SNMP_HOST: snmp.example
      SNMP_PORT: 161
      COMMUNITY: public
      V2: true      
      COLLECTION_FILES: "/etc/newrelic-infra/integrations.d/snmp-metrics.yml"
```
And your metrics collection file is:
```yaml
---
collect:
- device: NR-SNMP-MIB
  metric_sets:
  - name: scalar metrics
    type: scalar
    event_type: SNMPSample
    metrics:
    - metric_name: newrelicExampleInteger
      oid: .1.3.6.1.4.1.52032.1.1.1.0
    - metric_name: newrelicExampleMetric
      oid: .1.3.6.1.4.1.52032.1.1.2.0
    - metric_name: newrelicExampleString
      oid: .1.3.6.1.4.1.52032.1.1.3.0
```

Then to connect with `snmpwalk` you'll need to:

```shell script
snmpwalk \
    -One \ # Helps with printing the OIDs
    -v2c \ # SNMP v2
    -c public \ # Sets the COMMUNITY to "public"
    snmp.example:161 \ # The location of the endpoint
    1.3.6.1.4.1.52032 \ # OID to start walking through
```

The `OID` comes from the metrics collection file so chose an index from anywhere you want to walk from.
As this is a connection test it isn't important.

A more complicated V3 connection configuration like:

```yaml
---
integrations:
  - name: nri-snmp
    env:
      # Use the discovered IP as the host address
      SNMP_HOST: snmp.example
      SNMP_PORT: 161
      V3: true
      SECURITY_LEVEL: authPriv
      AUTH_PASSPHRASE: AUTH_XXXXXXXXXX
      AUTH_PROTOCOL: MD5
      PRIV_PASSPHRASE: PRIV_XXXXXXXXXX
      PRIV_PROTOCOL: DES
      USERNAME: snmpv3user
      COLLECTION_FILES: "/etc/newrelic-infra/integrations.d/snmp-metrics.yml"
```

Would become:

```shell script
snmpwalk \
  -v3 \ # V3 has been set to true
  -l authPriv \ # SECURITY_LEVEL
  -u snmpv3user_test \ # USERNAME
  -a MD5 \ # AUTH_PROTOCOL
  -A AUTH_XXXXXXXXXX \ # AUTH_PASSPHRASE
  -x DES \ # PRIV_PROTOCOL
  -X PRIV_XXXXXXXXXX \ # PRIV_PASSPHRASE
  snmp.example:161 \ # The location of the endpoint
  1.3.6.1.4.1.52032 \ # OID to start walking through
```
