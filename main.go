package main

import (
	"flag"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gobike/envflag"
	"log"
	"os"
)

func main() {
	var (
		configFilePath                                                                                                                        string
		processEc2Instances, processRdsInstances, processRdsClusters, processCloudwatchLogGroups, processElasticSearch, processCloudFrontDist bool
	)
	flag.StringVar(&configFilePath, "config", "config.json", "Path of the json configuration file. Environment variable: CONFIG")
	flag.BoolVar(&processEc2Instances, "ec2-instances", false, "Enables the re-tagging of the EC2 instances. Environment variable: EC2_INSTANCES")
	flag.BoolVar(&processRdsInstances, "rds-instances", false, "Enables the re-tagging of the RDS instances. Environment variable: RDS_INSTANCES")
	flag.BoolVar(&processRdsClusters, "rds-clusters", false, "Enables the re-tagging of the RDS clusters. Environment variable: RDS_CLUSTERS")
	flag.BoolVar(&processCloudwatchLogGroups, "cloudwatch-groups", false, "Enables the re-tagging of the CloudWatch log groups. Environment variable: CLOUDWATCH_GROUPS")
	flag.BoolVar(&processElasticSearch, "elasticsearch", false, "Enables the re-tagging of the ElasticSearch domains. Environment variable: ELASTICSEARCH")
	flag.BoolVar(&processCloudFrontDist, "cloudfront-distributions", false, "Enables the re-tagging of the CloudFront distributions. Environment variable: CLOUDFRONT_DISTRIBUTIONS")
	envflag.Parse()

	// Load config
	cfg, err := os.Open(configFilePath)
	defer cfg.Close()
	if err != nil {
		log.Fatalf("Unable to read config file: %s\n", err)
	}

	m := Mapper{}
	if err = m.LoadConfig(cfg); err != nil {
		log.Fatalf("Unable to load config file: %s\n", err)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if processEc2Instances {
		e := NewEc2Processor(sess)
		e.RetagInstances(&m)
	}

	if processRdsInstances || processRdsClusters {
		r := NewRdsProcessor(sess)
		if processRdsInstances {
			r.RetagInstances(&m)
		}
		if processRdsClusters {
			r.RetagClusters(&m)
		}
	}

	if processCloudwatchLogGroups {
		c := NewCwProcessor(sess)
		c.RetagLogGroups(&m)
	}

	if processElasticSearch {
		elk := NewElkProcessor(sess)
		elk.RetagDomains(&m)
	}

	if processCloudFrontDist {
		cf := NewCloudFrontProcessor(sess)
		cf.RetagDistributions(&m)
	}
}
