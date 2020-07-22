[![Community Project header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Community_Project.png)](https://opensource.newrelic.com/oss-category/#community-project)

# New Relic Infrastructure integration for SNMP

The New Relic integration for SNMP captures critical performance metrics and inventory reported by an SNMP server.

Metric data is obtained by making GET requests for configured list of OIDs and SNMP walk requests for the configured list of SNMP tables.

## Install

For installation and usage instructions, see our [official documentation](https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/snmp-monitoring-integration).

## Build

After cloning this repository, go to the directory of the SNMP Integration and build it:

```bash
$ make
```

The command above executes the tests for the SNMP Integration and builds an executable file named `nri-snmp` under `bin` directory. 

To start the integration, run `nri-snmp`:

```bash
$ ./bin/nri-snmp
```

If you want to know more about the usage of `./bin/nri-snmp`, pass the `-help` parameter:

```bash
$ ./bin/nri-snmp -help
```

External dependencies are managed through the [govendor tool](https://github.com/kardianos/govendor). Locking all external dependencies to a specific version (if possible) into the vendor directory is required.

## Testing

To run the tests execute:

```bash
$ make test
```

## Support

New Relic hosts and moderates an online forum where customers can interact with New Relic employees as well as other customers to get help and share best practices. Like all official New Relic open source projects, there's a related Community topic in the New Relic Explorers Hub. You can find this project's topic/threads here:

https://discuss.newrelic.com/c/support-products-agents/new-relic-infrastructure

# Contributing

We encourage contributions to improve New Relic infrastructure agent Chef cookbook! Keep in mind when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.
If you have any questions, or to execute our corporate CLA, required if your contribution is on behalf of a company,  please drop us an email at opensource@newrelic.com.

## License
New Relic Infrastructure Integration for SNMP is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.

The New Relic Infrastructure Integration for SNMP also uses source code from third-party libraries. You can find full details on which libraries are used and the terms under which they are licensed in the [third-party notices](./THIRD_PARTY_NOTICES.md) document.
