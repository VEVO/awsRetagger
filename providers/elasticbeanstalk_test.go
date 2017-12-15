package providers

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
)

func TestElasticBeanstalkTagsToMap(t *testing.T) {
	testData := []struct {
		inputTags  []*elasticbeanstalk.Tag
		outputTags map[string]string
	}{
		{[]*elasticbeanstalk.Tag{}, map[string]string{}},
		{[]*elasticbeanstalk.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}}, map[string]string{"foo": "bar"}},
		{[]*elasticbeanstalk.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}, {Key: aws.String("Aerosmith"), Value: aws.String("rocks")}}, map[string]string{"foo": "bar", "Aerosmith": "rocks"}},
	}
	for _, d := range testData {
		p := ElasticBeanstalkProcessor{}

		res := p.TagsToMap(d.inputTags)
		if !reflect.DeepEqual(res, d.outputTags) {
			t.Errorf("Expecting to get tags: %v\nGot: %v\n", d.outputTags, res)
		}
	}
}
