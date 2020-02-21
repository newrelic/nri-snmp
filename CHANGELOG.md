# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## 1.1.0 (2019-11-18)
### Changed
- Better support for SNMP v3
- Added id attributes to metric sets to work with delta, rate source type metrics

## 1.1.0 (2019-11-18)
### Changed
- Renamed the integration executable from nr-snmp to nri-snmp in order to be consistent with the package naming. **Important Note:** if you have any security module rules (eg. SELinux), alerts or automation that depends on the name of this binary, these will have to be updated.
## [1.0.4] - 2019-07-23
- Removed unneeded nrjmx dependency

## [1.0.3] - 2019-03-12

### Added

- Added connection and SNMP errors to be reported as an error event

## [1.0.2] - 2019-01-10

### Changed

- Renamed sample.metrics.yml to snmp-metrics.yml.sample

## [1.0.1] - 2018-10-27

### Changed

- Better error handling.
- Handle (gracefully) scalar OIDs that do not end in a .0
- Added support for OID types ObjectIdentifier, IPAddress, OpaqueFloat and OpaqueDouble

## [1.0.0] - 2018-08-30

### Added

- Initial version: Includes Metrics and Inventory data
