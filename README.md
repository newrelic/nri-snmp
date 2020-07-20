[![Community Project header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Community_Project.png)](https://opensource.newrelic.com/oss-category/#community-project)

# New Relic Infrastructure Integration for SNMP

New Relic infrastructure integration for SNMP captures critical performance
metrics and inventory reported by a SNMP server.

Metric data is obtained by making SNMP GET requests for configured list of OIDs and SNMP walk requests for the configured list of SNMP tables.

## Installing and using New Relic infrastructure agent Chef cookbook

* [Installation and usage instructions](https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/snmp-monitoring-integration)


## Building

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

For managing external dependencies [govendor tool](https://github.com/kardianos/govendor) is used. It is required to lock all external dependencies to a specific version (if possible) into the vendor directory.

## Testing

To run the tests execute

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
