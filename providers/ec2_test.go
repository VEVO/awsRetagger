package providers

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestEc2TagsToMap(t *testing.T) {
	testData := []struct {
		inputTags  []*ec2.Tag
		outputTags map[string]string
	}{
		{[]*ec2.Tag{}, map[string]string{}},
		{[]*ec2.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, map[string]string{"foo": "bar"}},
		{[]*ec2.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}, {Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}, map[string]string{"foo": "bar", "Aerosmith": "rocks"}},
	}
	for _, d := range testData {
		p := Ec2Processor{}

		res := p.TagsToMap(d.inputTags)
		if !reflect.DeepEqual(res, d.outputTags) {
			t.Errorf("Expecting to get tags: %v\nGot: %v\n", d.outputTags, res)
		}
	}
}
