package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gobike/envflag"
	"github.com/sirupsen/logrus"

	"github.com/VEVO/awsRetagger/mapper"
	"github.com/VEVO/awsRetagger/providers"
)

var log *logrus.Entry

// NewLogger creates a new logger instance
func NewLogger(logLevel, format string, output io.Writer) (*logrus.Entry, error) {
	switch format {
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		return nil, fmt.Errorf("invalid format requested: %s", format)
	}
	logrus.SetOutput(output)

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}
	logrus.SetLevel(level)
	context := logrus.WithFields(logrus.Fields{
		"app": "awsRetagger",
	})

	return context, nil
}

func main() {
	var (
		configFilePath, logLevel, logFormat                                                                                                                                                                          string
		processEc2Instances, processRdsInstances, processRdsClusters, processCloudwatchLogGroups, processElasticSearch, processCloudFrontDist, processRedshiftClusters, processElasticBeanstalkEnv, processS3Buckets bool
		err                                                                                                                                                                                                          error
	)
	flag.StringVar(&configFilePath, "config", "config.json", "Path of the json configuration file. Environment variable: CONFIG")
	flag.StringVar(&logLevel, "log-level", "info", "Log level. Accepted values: debug, info, warn, error, fatal, panic. Environment variable: LOG_LEVEL")
	flag.StringVar(&logFormat, "log-format", "text", "Log format. Accepted values: text, json. Environment variable: LOG_FORMAT")
	flag.BoolVar(&processEc2Instances, "ec2-instances", false, "Enables the re-tagging of the EC2 instances. Environment variable: EC2_INSTANCES")
	flag.BoolVar(&processRdsInstances, "rds-instances", false, "Enables the re-tagging of the RDS instances. Environment variable: RDS_INSTANCES")
	flag.BoolVar(&processRdsClusters, "rds-clusters", false, "Enables the re-tagging of the RDS clusters. Environment variable: RDS_CLUSTERS")
	flag.BoolVar(&processCloudwatchLogGroups, "cloudwatch-groups", false, "Enables the re-tagging of the CloudWatch log groups. Environment variable: CLOUDWATCH_GROUPS")
	flag.BoolVar(&processElasticSearch, "elasticsearch", false, "Enables the re-tagging of the ElasticSearch domains. Environment variable: ELASTICSEARCH")
	flag.BoolVar(&processCloudFrontDist, "cloudfront-distributions", false, "Enables the re-tagging of the CloudFront distributions. Environment variable: CLOUDFRONT_DISTRIBUTIONS")
	flag.BoolVar(&processRedshiftClusters, "redshift-clusters", false, "Enables the re-tagging of the Redshift clusters. Environment variable: REDSHIFT_CLUSTERS")
	flag.BoolVar(&processElasticBeanstalkEnv, "elasticbeanstalk-environments", false, "Enables the re-tagging of the ElasticBeanstalk environments. Environment variable: ELASTICBEANSTALK_ENVIRONMENTS")
	flag.BoolVar(&processS3Buckets, "s3-buckets", false, "Enables the re-tagging of the S3 buckets. Environment variable: S3_BUCKETS")
	envflag.Parse()

	if log, err = NewLogger(logLevel, logFormat, os.Stdout); err != nil {
		fmt.Printf("Error while setting up the logger: %s\n", err)
		os.Exit(1)
	}
	mapper.SetLogger(log)
	providers.SetLogger(log)
	// Load config
	cfg, err := os.Open(configFilePath)
	defer cfg.Close()
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Unable to read config file")
	}

	m := mapper.Mapper{}
	if err = m.LoadConfig(cfg); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Unable to load config file")
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if processEc2Instances {
		e := providers.NewEc2Processor(sess)
		e.RetagInstances(&m)
	}

	if processRdsInstances || processRdsClusters {
		r := providers.NewRdsProcessor(sess)
		if processRdsInstances {
			r.RetagInstances(&m)
		}
		if processRdsClusters {
			r.RetagClusters(&m)
		}
	}

	if processCloudwatchLogGroups {
		c := providers.NewCwProcessor(sess)
		c.RetagLogGroups(&m)
	}

	if processElasticSearch {
		elk := providers.NewElkProcessor(sess)
		elk.RetagDomains(&m)
	}

	if processCloudFrontDist {
		cf := providers.NewCloudFrontProcessor(sess)
		cf.RetagDistributions(&m)
	}
	if processRedshiftClusters {
		rs, err := providers.NewRedshiftProcessor(sess)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err}).Fatal("Unable to initialize the Redshift client")
		}
		rs.RetagClusters(&m)
	}
	if processElasticBeanstalkEnv {
		eb := providers.NewElasticBeanstalkProcessor(sess)
		eb.RetagEnvironments(&m)
	}

	if processS3Buckets {
		sp := providers.NewS3Processor(sess)
		sp.RetagBuckets(&m)
	}
}
