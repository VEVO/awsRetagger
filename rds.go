package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
)

// RdsProcessor holds the rds-related actions
type RdsProcessor struct {
	svc *rds.RDS
}

// NewRdsProcessor creates a new instance of RdsProcessor containing an already
// initialized rds client
func NewRdsProcessor(sess *session.Session) *RdsProcessor {
	return &RdsProcessor{svc: rds.New(sess)}
}

// TagsToMap transform the rds tags structure into a map[string]string for
// easier manipulations
func (p *RdsProcessor) TagsToMap(tagsInput []*rds.Tag) map[string]string {
	tagsHash := make(map[string]string)
	for _, tag := range tagsInput {
		tagsHash[*tag.Key] = *tag.Value
	}
	return tagsHash
}

// SetTag sets a tag on an rds resource
func (p *RdsProcessor) SetTag(resourceID *string, tag *TagItem) error {
	input := &rds.AddTagsToResourceInput{
		ResourceName: resourceID,
		Tags:         []*rds.Tag{{Key: aws.String((*tag).Name), Value: aws.String((*tag).Value)}},
	}
	_, err := p.svc.AddTagsToResource(input)
	return err
}

// GetTags gets the tags allocated to an rds resource
func (p *RdsProcessor) GetTags(resourceID *string) ([]*rds.Tag, error) {
	input := &rds.ListTagsForResourceInput{
		ResourceName: resourceID,
	}

	result, err := p.svc.ListTagsForResource(input)
	return result.TagList, err
}

// RetagInstances parses all instances and retags them
func (p *RdsProcessor) RetagInstances(m *Mapper) {
	result, err := p.svc.DescribeDBInstances(&rds.DescribeDBInstancesInput{})
	if err != nil {
		log.Fatalf("[ERROR] DescribeDBInstances returned: %s\n", err.Error())
	}

	for _, instance := range result.DBInstances {
		t, err := p.GetTags(instance.DBInstanceArn)
		if err != nil {
			log.Fatalf("[ERROR] Getting DB instance %s tags returned: %s\n", *instance.DBInstanceArn, err.Error())
		}

		tags := p.TagsToMap(t)
		keys := []string{}
		if instance.DBClusterIdentifier != nil {
			keys = append(keys, *instance.DBClusterIdentifier)
		}
		if instance.DBInstanceIdentifier != nil {
			keys = append(keys, *instance.DBInstanceIdentifier)
		}
		if instance.DBName != nil {
			keys = append(keys, *instance.DBName)
		}
		if instance.MasterUsername != nil {
			keys = append(keys, *instance.MasterUsername)
		}
		m.Retag(instance.DBInstanceArn, &tags, keys, p.SetTag)
	}
}

// RetagClusters parses all clusters and retags them
func (p *RdsProcessor) RetagClusters(m *Mapper) {
	result, err := p.svc.DescribeDBClusters(&rds.DescribeDBClustersInput{})
	if err != nil {
		log.Fatalf("[ERROR] DescribeDBClusters returned: %s\n", err.Error())
	}

	for _, cluster := range result.DBClusters {
		t, err := p.GetTags(cluster.DBClusterArn)
		if err != nil {
			log.Fatalf("[ERROR] Getting DB cluster %s tags returned: %s\n", *cluster.DBClusterArn, err.Error())
		}

		tags := p.TagsToMap(t)
		keys := []string{}
		if cluster.DBClusterIdentifier != nil {
			keys = append(keys, *cluster.DBClusterIdentifier)
		}
		if cluster.DatabaseName != nil {
			keys = append(keys, *cluster.DatabaseName)
		}
		if cluster.MasterUsername != nil {
			keys = append(keys, *cluster.MasterUsername)
		}
		m.Retag(cluster.DBClusterArn, &tags, keys, p.SetTag)
	}
}
