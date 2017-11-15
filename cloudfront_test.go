package main

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"
	"reflect"
	"testing"
)

type mockCloudFrontClient struct {
	cloudfrontiface.CloudFrontAPI
	// ResourceID is the resource that has been passed to the mocked function
	ResourceID *string
	// ResourceTags are the tags that have been passed to the mocked function when
	// setting or that is available on the mocked resource when getting
	ResourceTags *cloudfront.Tags
	// ReturnError is the error that you want your mocked function to return
	ReturnError error
}

func (m *mockCloudFrontClient) TagResource(input *cloudfront.TagResourceInput) (*cloudfront.TagResourceOutput, error) {
	m.ResourceID = input.Resource
	m.ResourceTags = input.Tags
	return nil, m.ReturnError
}

func (m *mockCloudFrontClient) ListTagsForResource(input *cloudfront.ListTagsForResourceInput) (*cloudfront.ListTagsForResourceOutput, error) {
	return nil, m.ReturnError
}

func TestSetTag(t *testing.T) {
	testData := []struct {
		inputResource, outputResource string
		inputTag                      *TagItem
		outputTag                     *cloudfront.Tags
		inputError, outputError       error
	}{
		{"my resource", "my resource", &TagItem{}, &cloudfront.Tags{Items: []*cloudfront.Tag{{Key: aws.String(""), Value: aws.String("")}}}, nil, nil},
		{"my resource", "my resource", &TagItem{Name: "foo", Value: "bar"}, &cloudfront.Tags{Items: []*cloudfront.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}}, nil, nil},
		{"my resource", "my resource", &TagItem{Name: "foo", Value: "bar"}, &cloudfront.Tags{Items: []*cloudfront.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}}, errors.New("Badaboom"), errors.New("Badaboom")},
	}
	for _, d := range testData {
		mockSvc := &mockCloudFrontClient{ReturnError: d.inputError}
		p := CloudFrontProcessor{svc: mockSvc}
		err := p.SetTag(&d.inputResource, d.inputTag)
		if !reflect.DeepEqual(err, d.outputError) {
			t.Errorf("Expecting error: %v\nGot: %v\n", d.outputError, err)
		}

		if *mockSvc.ResourceID != d.outputResource {
			t.Errorf("Expecting to update resource: %s, got: %s\n", d.outputResource, *mockSvc.ResourceID)
		}

		if !reflect.DeepEqual(*mockSvc.ResourceTags, *d.outputTag) {
			t.Errorf("Expecting to update tag: %v\nGot: %v\n", *d.outputTag, *mockSvc.ResourceTags)
		}
	}
}
