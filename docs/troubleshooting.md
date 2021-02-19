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

## Debugging SNMP values

If the integration is connecting correctly but not producing the correct metrics then there are two options:

* Connect to the device and play with the config till it works
* Take a sample of the output and replay it locally

### Capturing production data to use locally

You must be able to connect to device and run `snmpwalk` against it.
To get a backup you need to run the `snmpwalk` command as follows:

```shell
$ snmpwalk -One <connection arguments>
```

Example, if you spin up the Docker Compose file in `build/troubleshooting` folder either by:

* `docker-compose -f ./build/troubleshooting/docker-compose.yml up`
* `make docker-snmp-example`

This spins up mock SNMP server with some canned data.
To connect to it and make a copy of the data just run:

```shell
$ snmpwalk -One  -v2c -c public 127.0.0.1:1024 > toubleshooting.snmpwalk
```

This creates a file called `toubleshooting.snmpwalk` which we can replay back locally.

### Replaying the data

Once you have the file place it in the route of the project and run:

```shell
$ make docker-snmp-troubleshooting 
```

This spins up a Docker container with the file and creates a stub SNMP server that will serve you back the data in `toubleshooting.snmpwalk`.
You can see this by running snmp work against it as follows:

```shell
$ snmpwalk -Of -v2c -c toubleshooting 127.0.0.1:1024
```

You are now ready to run `nri-snmp` against it and check why the config does not produce the data you expect.
To do this you need to make sure you have the collections file you are debugging (e.g. `toubleshooting-collections.yml`).
Then run:

```shell
$ nri-snmp -verbose -snmp_host localhost -collection_files <path to toubleshooting-collections.yml> -snmp_port 1024 -community toubleshooting | jq .
```

If you're just built this image from the example SNMP server you can run the following to get some data:

```shell
$ nri-snmp -verbose -snmp_host localhost -collection_files $PWD/build/troubleshooting/toubleshooting-collections-example.yml  -snmp_port 1024 -community toubleshooting | jq .
```

This will give you the same data as if you have run it against the community you made the capture from:

```shell
$ nri-snmp -verbose -snmp_host localhost -collection_files $PWD/build/troubleshooting/toubleshooting-collections-example.yml  -snmp_port 1024 -community public | jq .
```
