package main

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/redshift/redshiftiface"
	"reflect"
	"testing"
)

type mockRedshiftClient struct {
	redshiftiface.RedshiftAPI
	// ResourceID is the resource that has been passed to the mocked function
	ResourceID *string
	// ResourceTags are the tags that have been passed to the mocked function when
	// setting or that is available on the mocked resource when getting
	ResourceTags []*redshift.Tag
	// ReturnError is the error that you want your mocked function to return
	ReturnError error
}

func (m *mockRedshiftClient) CreateTags(input *redshift.CreateTagsInput) (*redshift.CreateTagsOutput, error) {
	m.ResourceID = input.ResourceName
	m.ResourceTags = append(m.ResourceTags, input.Tags...)
	return &redshift.CreateTagsOutput{}, m.ReturnError
}

func (m *mockRedshiftClient) DescribeTags(input *redshift.DescribeTagsInput) (*redshift.DescribeTagsOutput, error) {
	m.ResourceID = input.ResourceName
	outTags := []*redshift.TaggedResource{}
	for _, otag := range m.ResourceTags {
		outTags = append(outTags, &redshift.TaggedResource{ResourceName: input.ResourceName, Tag: otag})
	}
	return &redshift.DescribeTagsOutput{TaggedResources: outTags}, m.ReturnError
}

func TestRedshiftSetTags(t *testing.T) {
	testData := []struct {
		inputResource, outputResource string
		inputTags                     []*TagItem
		outputTags                    []*redshift.Tag
		inputError, outputError       error
	}{
		{"my resource", "my resource", []*TagItem{{}}, []*redshift.Tag{{Key: aws.String(""), Value: aws.String("")}}, nil, nil},
		{"my resource", "my resource", []*TagItem{{Name: "foo", Value: "bar"}}, []*redshift.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, nil, nil},
		{"my resource", "my resource", []*TagItem{{Name: "foo", Value: "bar"}, {Name: "Aerosmith", Value: "rocks"}}, []*redshift.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}, {Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}, nil, nil},
		{"my resource", "my resource", []*TagItem{{Name: "foo", Value: "bar"}}, []*redshift.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, errors.New("Badaboom"), errors.New("Badaboom")},
	}
	for _, d := range testData {
		mockSvc := &mockRedshiftClient{ReturnError: d.inputError, ResourceTags: []*redshift.Tag{}}
		p := RedshiftProcessor{svc: mockSvc}

		err := p.SetTags(&d.inputResource, d.inputTags)
		if !reflect.DeepEqual(err, d.outputError) {
			t.Errorf("Expecting error: %v\nGot: %v\n", d.outputError, err)
		}

		if *mockSvc.ResourceID != d.outputResource {
			t.Errorf("Expecting to update resource: %s, got: %s\n", d.outputResource, *mockSvc.ResourceID)
		}

		if !reflect.DeepEqual(mockSvc.ResourceTags, d.outputTags) {
			t.Errorf("Expecting to update tag: %v\nGot: %v\n", d.outputTags, mockSvc.ResourceTags)
		}
	}
}

func TestRedshiftTagsToMap(t *testing.T) {
	testData := []struct {
		inputTags  []*redshift.TaggedResource
		outputTags map[string]string
	}{
		{[]*redshift.TaggedResource{}, map[string]string{}},
		{[]*redshift.TaggedResource{{Tag: &redshift.Tag{Key: aws.String("foo"), Value: aws.String("bar")}}}, map[string]string{"foo": "bar"}},
		{[]*redshift.TaggedResource{{Tag: &redshift.Tag{Key: aws.String("foo"), Value: aws.String("bar")}}, {Tag: &redshift.Tag{Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}}, map[string]string{"foo": "bar", "Aerosmith": "rocks"}},
	}
	for _, d := range testData {
		p := RedshiftProcessor{}

		res := p.TagsToMap(d.inputTags)
		if !reflect.DeepEqual(res, d.outputTags) {
			t.Errorf("Expecting to get tags: %v\nGot: %v\n", d.outputTags, res)
		}
	}
}
