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

```json
  "tags": [
    {"source": {"name": "Project", "value": "API"}, "destination":[
      {"name": "team", "value": "api"}
    ]},
    {"source": {"name": "Name", "value": ".*jenkins.*"}, "destination":[
      {"name": "team", "value": "infrastructure"},
      {"name": "env", "value": "prd"},
      {"name": "service", "value": "ci"}
    ]},
    {"source": {"name": "Name", "value": ".*staging.*"}, "destination":[{"name": "env", "value": "stg"}]},
    {"source": {"name": "Name", "value": ".*stg.*"}, "destination":[{"name": "env", "value": "stg"}]},
    {"source": {"name": "Name", "value": ".*prod.*"}, "destination":[{"name": "env", "value": "prd"}]},
    {"source": {"name": "Name", "value": ".*prd.*"}, "destination":[{"name": "env", "value": "prd"}]}
  ]
```

### The `keys` mapping

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

```json
  "defaults": {
    "env":     "unknown",
    "team":    "unknown",
    "service": "unknown"
  }
```

## Using the tool

```
$ ./awsRetagger -h
Usage of ./awsRetagger:
  -cloudfront-distributions
        Enables the re-tagging of the CloudFront distributions. Environment variable: CLOUDFRONT_DISTRIBUTIONS
  -cloudwatch-groups
        Enables the re-tagging of the CloudWatch log groups. Environment variable: CLOUDWATCH_GROUPS
  -ec2-instances
        Enables the re-tagging of the EC2 instances. Environment variable: EC2_INSTANCES
  -elasticsearch
        Enables the re-tagging of the ElasticSearch domains. Environment variable: ELASTICSEARCH
  -rds-clusters
        Enables the re-tagging of the RDS clusters. Environment variable: RDS_CLUSTERS
  -rds-instances
        Enables the re-tagging of the RDS instances. Environment variable: RDS_INSTANCES
```
