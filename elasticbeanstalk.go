package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk/elasticbeanstalkiface"
	"github.com/sirupsen/logrus"
)

// ElasticBeanstalkProcessor holds the elasticbeanstalk-related actions
type ElasticBeanstalkProcessor struct {
	svc elasticbeanstalkiface.ElasticBeanstalkAPI
}

// NewElasticBeanstalkProcessor creates a new instance of ElasticBeanstalkProcessor containing an already
// initialized elasticbeanstalk client
func NewElasticBeanstalkProcessor(sess *session.Session) *ElasticBeanstalkProcessor {
	return &ElasticBeanstalkProcessor{svc: elasticbeanstalk.New(sess)}
}

// TagsToMap transform the elasticbeanstalk tags structure into a map[string]string for
// easier manipulations
func (p *ElasticBeanstalkProcessor) TagsToMap(tagsInput []*elasticbeanstalk.Tag) map[string]string {
	tagsHash := make(map[string]string)
	for _, tag := range tagsInput {
		tagsHash[*tag.Key] = *tag.Value
	}
	return tagsHash
}

// SetTags sets a group of tag on an elasticbeanstalk resource
func (p *ElasticBeanstalkProcessor) SetTags(resourceID *string, tags []*TagItem) error {
	newTags := []*elasticbeanstalk.Tag{}
	for _, tag := range tags {
		if (*tag).Name != "elasticbeanstalk:environment-name" && (*tag).Name != "elasticbeanstalk:environment-id" {
			newTags = append(newTags, &elasticbeanstalk.Tag{Key: aws.String((*tag).Name), Value: aws.String((*tag).Value)})
		}
	}
	if len(newTags) == 0 {
		return nil
	}
	input := &elasticbeanstalk.UpdateTagsForResourceInput{
		ResourceArn: resourceID,
		TagsToAdd:   newTags,
	}
	_, err := p.svc.UpdateTagsForResource(input)
	return err
}

// GetTags gets the tags allocated to an elasticbeanstalk resource
func (p *ElasticBeanstalkProcessor) GetTags(resourceID *string) ([]*elasticbeanstalk.Tag, error) {
	input := &elasticbeanstalk.ListTagsForResourceInput{
		ResourceArn: resourceID,
	}

	result, err := p.svc.ListTagsForResource(input)
	return result.ResourceTags, err
}

// RetagEnvironments parses all environments and retags them
func (p *ElasticBeanstalkProcessor) RetagEnvironments(m *Mapper) {
	envs, err := p.svc.DescribeEnvironments(&elasticbeanstalk.DescribeEnvironmentsInput{IncludeDeleted: aws.Bool(false)})
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("DescribeEnvironments failed")
	}
	for _, env := range envs.Environments {
		if *env.Status != "Ready" || *env.Health == "Grey" {
			continue // only the "Ready" environments can be retagged
		}
		t, err := p.GetTags(env.EnvironmentArn)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err, "resource": *env.EnvironmentArn}).Fatal("Failed to get ElasticBeanstalk environment tags")
		}
		tags := p.TagsToMap(t)
		keys := []string{}
		if env.EnvironmentName != nil {
			keys = append(keys, *env.EnvironmentName)
		}
		if env.ApplicationName != nil {
			keys = append(keys, *env.ApplicationName)
		}
		if env.CNAME != nil {
			keys = append(keys, *env.CNAME)
		}
		if env.Description != nil {
			keys = append(keys, *env.Description)
		}
		m.Retag(env.EnvironmentArn, &tags, keys, p.SetTags)
	}

}
