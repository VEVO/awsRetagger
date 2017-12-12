package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

type mockRdsClient struct {
	rdsiface.RDSAPI
	// ResourceID is the resource that has been passed to the mocked function
	ResourceID *string
	// ResourceTags are the tags that have been passed to the mocked function when
	// setting or that is available on the mocked resource when getting
	ResourceTags []*rds.Tag
	// ReturnError is the error that you want your mocked function to return
	ReturnError error
}

func (m *mockRdsClient) AddTagsToResource(input *rds.AddTagsToResourceInput) (*rds.AddTagsToResourceOutput, error) {
	m.ResourceID = input.ResourceName
	if input.Tags != nil {
		m.ResourceTags = append(m.ResourceTags, input.Tags...)
	}
	return &rds.AddTagsToResourceOutput{}, m.ReturnError
}

func (m *mockRdsClient) ListTagsForResource(input *rds.ListTagsForResourceInput) (*rds.ListTagsForResourceOutput, error) {
	m.ResourceID = input.ResourceName
	return &rds.ListTagsForResourceOutput{TagList: m.ResourceTags}, m.ReturnError
}

func TestRdsTagsToMap(t *testing.T) {
	testData := []struct {
		inputTags  []*rds.Tag
		outputTags map[string]string
	}{
		{[]*rds.Tag{}, map[string]string{}},
		{[]*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, map[string]string{"foo": "bar"}},
		{[]*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}, {Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}, map[string]string{"foo": "bar", "Aerosmith": "rocks"}},
	}
	for _, d := range testData {
		p := RdsProcessor{}

		res := p.TagsToMap(d.inputTags)
		if !reflect.DeepEqual(res, d.outputTags) {
			t.Errorf("Expecting to get tags: %v\nGot: %v\n", d.outputTags, res)
		}
	}
}

func TestRdsSetTags(t *testing.T) {
	testData := []struct {
		inputResource, outputResource string
		inputTags                     []*TagItem
		outputTags                    []*rds.Tag
		inputError, outputError       error
	}{
		{"my resource", "my resource", []*TagItem{{}}, []*rds.Tag{{Key: aws.String(""), Value: aws.String("")}}, nil, nil},
		{"my resource", "my resource", []*TagItem{{Name: "foo", Value: "bar"}}, []*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, nil, nil},
		{"my resource", "my resource", []*TagItem{{Name: "foo", Value: "bar"}, {Name: "Aerosmith", Value: "rocks"}}, []*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}, {Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}, nil, nil},
		{"my resource", "my resource", []*TagItem{{Name: "foo", Value: "bar"}}, []*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, errors.New("Badaboom"), errors.New("Badaboom")},
	}
	for _, d := range testData {
		mockSvc := &mockRdsClient{ReturnError: d.inputError, ResourceTags: []*rds.Tag{}}
		p := RdsProcessor{svc: mockSvc}

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

func TestRdsGetTags(t *testing.T) {
	testData := []struct {
		inputResource, outputResource string
		inputTags                     []*rds.Tag
		outputTags                    []*rds.Tag
		inputError, outputError       error
	}{
		{"my resource", "my resource", []*rds.Tag{}, []*rds.Tag{}, nil, nil},
		{"my resource", "my resource", []*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, []*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, nil, nil},
		{"my resource", "my resource", []*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}, {Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}, []*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}, {Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}, nil, nil},
		{"my resource", "my resource", []*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, []*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, errors.New("Badaboom"), errors.New("Badaboom")},
	}
	for _, d := range testData {
		mockSvc := &mockRdsClient{ReturnError: d.inputError, ResourceTags: d.inputTags}
		p := RdsProcessor{svc: mockSvc}

		res, err := p.GetTags(&d.inputResource)
		if !reflect.DeepEqual(err, d.outputError) {
			t.Errorf("Expecting error: %v\nGot: %v\n", d.outputError, err)
		}

		if *mockSvc.ResourceID != d.outputResource {
			t.Errorf("Expecting resource: %s, got: %s\n", d.outputResource, *mockSvc.ResourceID)
		}

		if !reflect.DeepEqual(res, d.outputTags) {
			t.Errorf("Expecting to get tags: %v\nGot: %v\n", d.outputTags, res)
		}
	}
}
