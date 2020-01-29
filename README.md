# New Relic Infrastructure Integration for SNMP

New Relic Infrastructure Integration for SNMP captures critical performance
metrics and inventory reported by an SNMP server.

Metrics data is obtained by making SNMP GET requests for configured list of OIDs and SNMP walk requests for the configured list of SNMP tables.

See our [documentation web site](https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/snmp-monitoring-integration) for more details.

## Usage

This is the description about how to run the SNMP Integration with New Relic
Infrastructure agent, so it is required to have the agent installed
(see
[agent installation](https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/installation/install-infrastructure-linux)).

In order to use the SNMP Integration it is required to configure
`snmp-config.yml` file. Depending on your needs, specify all instances that
you want to monitor with correct arguments.

## Integration development usage

Assuming that you have the source code and Go tool installed you can build and run the SNMP Integration locally.

* After cloning this repository, go to the directory of the SNMP Integration and build it

```bash
$ make
```

* The command above will execute the tests for the SNMP Integration and build an executable file called `nri-snmp` under `bin` directory. Run `nri-snmp`:

```bash
$ ./bin/nri-snmp
```

* If you want to know more about usage of `./bin/nri-snmp` check

```bash
$ ./bin/nri-snmp -help
```

For managing external dependencies [govendor tool](https://github.com/kardianos/govendor) is used. It is required to lock all external dependencies to specific version (if possible) into vendor directory.

## Contributing Code

We welcome code contributions (in the form of pull requests) from our user
community. Before submitting a pull request please review [these guidelines](https://github.com/newrelic/nri-SNMP/blob/master/CONTRIBUTING.md).

Following these helps us efficiently review and incorporate your contribution
and avoid breaking your code with future changes to the agent.

## Custom Integrations

To extend your monitoring solution with custom metrics, we offer the Integrations
Golang SDK which can be found on [github](https://github.com/newrelic/infra-integrations-sdk).

Refer to [our docs site](https://docs.newrelic.com/docs/infrastructure/integrations-sdk/get-started/intro-infrastructure-integrations-sdk)
to get help on how to build your custom integrations.

## Support

You can find more detailed documentation [on our website](http://newrelic.com/docs),
and specifically in the [Infrastructure category](https://docs.newrelic.com/docs/infrastructure).

If you can't find what you're looking for there, reach out to us on our [support
site](http://support.newrelic.com/) or our [community forum](http://forum.newrelic.com)
and we'll be happy to help you.

Find a bug? Contact us via [support.newrelic.com](http://support.newrelic.com/),
or email support@newrelic.com.

New Relic, Inc.
