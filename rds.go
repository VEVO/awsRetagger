package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/sirupsen/logrus"
)

// RdsProcessor holds the rds-related actions
type RdsProcessor struct {
	svc rdsiface.RDSAPI
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

// SetTags sets tags on an rds resource
func (p *RdsProcessor) SetTags(resourceID *string, tags []*TagItem) error {
	newTags := []*rds.Tag{}
	for _, tag := range tags {
		newTags = append(newTags, &rds.Tag{Key: aws.String((*tag).Name), Value: aws.String((*tag).Value)})
	}

	input := &rds.AddTagsToResourceInput{
		ResourceName: resourceID,
		Tags:         newTags,
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
		log.WithFields(logrus.Fields{"error": err}).Fatalf("DescribeDBInstances failed")
	}

	for _, instance := range result.DBInstances {
		t, err := p.GetTags(instance.DBInstanceArn)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err, "resource": *instance.DBInstanceArn}).Fatalf("Failed to get DB instance tags")
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
		m.Retag(instance.DBInstanceArn, &tags, keys, p.SetTags)
	}
}

// RetagClusters parses all clusters and retags them
func (p *RdsProcessor) RetagClusters(m *Mapper) {
	result, err := p.svc.DescribeDBClusters(&rds.DescribeDBClustersInput{})
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatalf("DescribeDBClusters failed")
	}

	for _, cluster := range result.DBClusters {
		t, err := p.GetTags(cluster.DBClusterArn)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err, "resource": *cluster.DBClusterArn}).Fatalf("Failed to get DB cluster tags")
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
		m.Retag(cluster.DBClusterArn, &tags, keys, p.SetTags)
	}
}
