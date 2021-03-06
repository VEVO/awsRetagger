# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/) 
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

### Changed
- Moved mapper and providers to separate packages for easier management
- Use goreleaser to make the releases
- Simplify the build process

### Added
- Add more unit tests
- Add support for retagging s3 buckets
- Add an interface and a Mocking base class for the mapper for easier unit
  testing

## [0.1.0] - 2017-11-22

1st public release

### Changed
- Use logrus as logging library and use the fields capabilities of logrus
  instead of string concatenation
- There are now 2 types of errors out of the sanity checks and the corresponding
  tag properties are available via `TagName` and `TagValue` attributes of the
  error
- The tags are updated all at once for a resource instead of 1 update request
  per tag

### Added
- Capability to log in text or json format based on user input
- Capability for the user to use different log levels
- Support for retagging Redshift Clusters
- Support for retagging ElasticBeanstalk Environments
- Automated Docker image build

## [0.0.1] - 2017-11-12
### Added
- Ability to retag AWS resources
- Supported resources:
  * CloudFront Distributions
  * CloudWatch LogGroups
  * EC2 Instances
  * ElasticSearch Domains
  * RDS Instances
  * RDS Clusters

[Unreleased]: https://github.com/VEVO/awsRetagger/compare/0.1.0...HEAD
[0.1.0]: https://github.com/VEVO/awsRetagger/compare/0.0.1...0.1.0
[0.0.1]: https://github.com/VEVO/awsRetagger/tree/0.0.1
