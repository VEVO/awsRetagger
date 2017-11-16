# awsRetagger

[![Build Status](https://travis-ci.org/VEVO/awsRetagger.svg?branch=master)](https://travis-ci.org/VEVO/awsRetagger)
[![Go Report Card](https://goreportcard.com/badge/github.com/VEVO/awsRetagger)](https://goreportcard.com/report/github.com/VEVO/awsRetagger)

## Table of Contents

  * [What's that?](#whats-that)
  * [How does it work?](#how-does-it-work)
    * [The copy_tag mapping](#the-copy_tag-mapping)
    * [The tags mapping](#the-tags-mapping)
    * [The keys mapping](#the-keys-mapping)
    * [The sanity mapping](#the-sanity-mapping)
    * [The defaults mapping](#the-defaults-mapping)
  * [Using the tool](#using-the-tool)
  * [Supported resources](#supported-resources)

## What's that?

awsRetagger is a tool to create/guess and normalize your tags over all the AWS services.

As a company that works in AWS for years we ended up with applications using a
lot of services and using a lot of different tagging standards. Not everything
is managed through terraform yet.

Then a reorganization happened and some teams got merged and suddently our cost
reporting becomes even more a nightmare...

This tool was created to apply a new norm of standard accross all our services
and avoid having to take care of misspells in tag names anymore for example.

## How does it work?

The core of the configuration uses the `config.json` (in the repo we provide a
`config-example.json` as an example) to map the existing configuration and guess
the values of your new tags.

> :exclamation: **Important:** When we talk about regular expressions for this applications, we
> always make sure that they are:
>  * case-insensitive (no need to add the case-insensitive flag)
>  * surrounded by `^` and `$` so the match of a word is an exact match

### The `copy_tag` mapping

The `copy_tag` mapping in the config.json is used to copy the content of a
potentially existing tag into a destination tag.

The rule is only applied if the destination tag is empty (and if its value is
not in the `defaults` tags mapping).

The `source` tags is the name (parsed using a case-insensitive regular
expression) of the tags you want to copy the content to the `destination` tag.
The tags are evaluated in the order given in the list, meaning  that in the
example bellow, if there are a `environment` and a `project` tag, the `service`
tag will get the content of the `environment` tag.

```json
  "copy_tags": [
    {"sources": ["division"], "destination": "team"},
    {"sources": ["env", "environment", "environmetnt", "account", "environment.*"], "destination": "env"},
    {"sources": ["servi.*ce", "application", "applicaiton", "app", "project", "micro.service"], "destination": "service"}
  ] 
```

### The `tags` mapping

The `tags` mapping allow you to guess one or more `destination` tag(s) based on
the content of a `source` tag. The name of the source tag is case-sensitive and must be
exact, but its value is a case-insensitive regular expression.

In the example bellow we use several values of the `team`, `service` and `env`
tags. The first math for a given destination tag will win. For example, if your
resource has a `Name` tag set to `myJenkins-Stg-Prd`, you will end up with:
* a `team` tag set to `infrastructure`
* a `service` tag set to `ci`
* an `env` tag set to `stg` (not that as the prd evaluation is done after, the env tag takes the 1st match)

```json
  "tags": [
    {"source": {"name": "Project", "value": "api"}, "destination":[
      {"name": "team", "value": "api"}
    ]},
    {"source": {"name": "Name", "value": ".*jenkins.*"}, "destination":[
      {"name": "team", "value": "infrastructure"},
      {"name": "service", "value": "ci"}
    ]},
    {"source": {"name": "Name", "value": ".*staging.*"}, "destination":[{"name": "env", "value": "stg"}]},
    {"source": {"name": "Name", "value": ".*stg.*"}, "destination":[{"name": "env", "value": "stg"}]},
    {"source": {"name": "Name", "value": ".*prod.*"}, "destination":[{"name": "env", "value": "prd"}]},
    {"source": {"name": "Name", "value": ".*prd.*"}, "destination":[{"name": "env", "value": "prd"}]}
  ]
```

### The `keys` mapping

The `keys` mapping allow you to map the value of one of the key elements of
your resource and apply tags based on it.

The value of the key element is, once again, a case-insensitive regular
expression.

The key elements of a resource are the following attribute (first come in the given order of a service wins) for:
* a CloudFront Distribution: `Id`, `DomainName`, Origins' `DomainName`, `Aliases`, `Comment`
* a CloudWatch LogGroup: `LogGroupName`
* an EC2 Instance: SSH `KeyName`
* an ElasticBeanstalk: `EnvironmentName`, `ApplicationName`, `CNAME`, `Description`
* an ElasticSearch Domain: `DomainId`, `DomainName`
* a RDS Instance: `DBClusterIdentifier`, `DBInstanceIdentifier`, `DBName`, `MasterUsername`
* a RDS Cluster: `DBClusterIdentifier`, `DatabaseName`, `MasterUsername`
* a Redshift Cluster: `ClusterIdentifier`, `DBName`, `MasterUsername`

With the following configuration, an instance with a SSH KeyName set to
`apple-tv-analytics-prod`, you'll end up with:
* a `team` tag set to `data` (not `tv` as evaluated later)
* an `env` tag set to `prd`
* a `service` tag set to `appletv`

```json
  "keys": [
    {"pattern": "joseph-.*", "destination":[
      {"name": "team", "value": "infrastructure"},
      {"name": "env", "value": "dev"},
      {"name": "service", "value": "personal-sandbox"}
    ]},
    {"pattern": ".*analytics.*", "destination":[{"name": "team", "value": "data"}]},
    {"pattern": ".*staging.*", "destination":[{"name": "env", "value": "stg"}]},
    {"pattern": ".*stg.*", "destination":[{"name": "env", "value": "stg"}]},
    {"pattern": ".*prd.*", "destination":[{"name": "env", "value": "prd"}]},
    {"pattern": ".*prod.*", "destination":[{"name": "env", "value": "prd"}]},
    {"pattern": ".*tv.*", "destination":[{"name": "team", "value": "tv"}]},
    {"pattern": ".*apple.*tv.*", "destination":[{"name": "service", "value": "appletv"}]}
  ]
```

### The `sanity` mapping

The `sanity` mapping allows you to make sure the values of a tag match a given
`remap` list. To those value you can give a list of case-insensitive regular
expressions. If the tag value matches one of these regular expressions, it will
be changed to the corresponding key.

In the bellow example, if your tag `env` is set to `developpement`, it will be
changed to `dev`. If the `team` tag was by mistake set to `user_servivices`, it
will be changed to `user-services`.

:white_check_mark: **Tip:** if 2 teams get merge, this is how you do it. For example here, the
`data` team and the `analytics` team gets merge into the `data` team.

```json
  "sanity": [
    {
      "tag_name": "env", "remap": {"prd": ["prod.*", "global"],"stg": ["stag.*"],"dev": ["dev.*"]}
    },
    {
      "tag_name": "team", "remap": {
        "infrastructure": ["infra.*", "systems.*", "syseng.*"],
        "android":[],
        "user-services": ["user.ser.*"],
        "data": ["data.*", "analytic.*"]
      }
    },
    {
      "tag_name": "service", "remap": {
        "encoding": [ "enconding" ],
        "vod": [".*vod-.*"],
        "personal-sandbox": ["jump", ".*jumpbox.*", ".*test.*"],
        "kubernetes": [".*-k8s-.*"],
        "elk": []
      }
    }
  ]
```

### The `defaults` mapping

The `defaults` mapping ensures that the given tags will always be tagged to at
least a value. If after passing all the previous mappings, the given tags are
not set, it will be set to its corresponding value.

:exclamation: **Important:** At the begining of the retagging process, the
tags with the default value are stripped from the list of tags. That means that
if you set to a value you want to be everywhere, the tags will be applied
everytime. If you want to have a default valuenot overwritten, use the sanity
mapping (example, add "unknown" to the `prd` key when running it in the
production account).

With the following configuration, the `env`, `team` and `service` that are not
set on a resource at the end of the process will be set to `unknown`.

The current AWS SDK doesn't allow you to filter by unset tags, so that helps you
to do some filtering more easily.

```json
  "defaults": {
    "env":     "unknown",
    "team":    "unknown",
    "service": "unknown"
  }
```

## Using the tool

To build the tool, run `make build`. Once build, you can use the `-h` option to
see the list of all options. All the options can be set using the corresponding
environment variable.

**Note:** to connect to AWS, make sure your credentials file is properly setup.
If you want to set a specific region that is not in your credentials, use the
`AWS_REGION` environment variable. If you want to use a profile different from
the default one, set the `AWS_PROFILE` environment variable. And as usual you
can use the other standard `AWS_*` environment variables as we force-enable the
shared configuration.

```
$ ./awsRetagger -h
Usage of ./awsRetagger:
  -cloudfront-distributions
        Enables the re-tagging of the CloudFront distributions. Environment variable: CLOUDFRONT_DISTRIBUTIONS
  -cloudwatch-groups
        Enables the re-tagging of the CloudWatch log groups. Environment variable: CLOUDWATCH_GROUPS
  -config string
        Path of the json configuration file. Environment variable: CONFIG (default "config.json")
  -ec2-instances
        Enables the re-tagging of the EC2 instances. Environment variable: EC2_INSTANCES
  -elasticbeanstalk-environments
        Enables the re-tagging of the ElasticBeanstalk environments. Environment variable: ELASTICBEANSTALK_ENVIRONMENTS
  -elasticsearch
        Enables the re-tagging of the ElasticSearch domains. Environment variable: ELASTICSEARCH
  -log-format string
        Log format. Accepted values: text, json. Environment variable: LOG_FORMAT (default "text")
  -log-level string
        Log level. Accepted values: debug, info, warn, error, fatal, panic. Environment variable: LOG_LEVEL (default "info")
  -rds-clusters
        Enables the re-tagging of the RDS clusters. Environment variable: RDS_CLUSTERS
  -rds-instances
        Enables the re-tagging of the RDS instances. Environment variable: RDS_INSTANCES
  -redshift-clusters
        Enables the re-tagging of the Redshift clusters. Environment variable: REDSHIFT_CLUSTERS
```

## Supported resources

Currently the awsRetagger can retag the following resources (but maybe more, so
you might check using the `-h` option of the command-line):
* CloudFront Distributions
* CloudWatch LogGroups
* EC2 Instances
* ElasticBeanstalk environments
* ElasticSearch Domains
* RDS Instances
* RDS Clusters
* Redshift Clusters
