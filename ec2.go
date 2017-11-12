package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
)

// Ec2Processor holds the ec2-related actions
type Ec2Processor struct {
	svc *ec2.EC2
}

// NewEc2Processor creates a new instance of Ec2Processor containing an already
// initialized ec2 client
func NewEc2Processor(sess *session.Session) *Ec2Processor {
	return &Ec2Processor{svc: ec2.New(sess)}
}

// TagsToMap transform the ec2 tags structure into a map[string]string for
// easier manipulations
func (e *Ec2Processor) TagsToMap(tagsInput []*ec2.Tag) map[string]string {
	tagsHash := make(map[string]string)
	for _, tag := range tagsInput {
		tagsHash[*tag.Key] = *tag.Value
	}
	return tagsHash
}

// SetTag sets a tag on an ec2 resource
func (e *Ec2Processor) SetTag(resourceID *string, tag *TagItem) error {
	_, err := e.svc.CreateTags(&ec2.CreateTagsInput{Resources: []*string{resourceID}, Tags: []*ec2.Tag{{Key: aws.String((*tag).Name), Value: aws.String((*tag).Value)}}})
	return err
}

// RetagInstances parses all running and stopped instances and retags them
func (e *Ec2Processor) RetagInstances(m *Mapper) {
	filters := []*ec2.Filter{
		{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running"), aws.String("stopped")},
		},
	}
	result, err := e.svc.DescribeInstances(&ec2.DescribeInstancesInput{Filters: filters})
	if err != nil {
		log.Fatalf("[ERROR] DescribeInstances returned: %v\n", err)
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			tags := e.TagsToMap(instance.Tags)
			keys := []string{}
			if instance.KeyName != nil {
				keys = append(keys, *instance.KeyName)
			}
			m.Retag(instance.InstanceId, &tags, keys, e.SetTag)
		}
	}
}
