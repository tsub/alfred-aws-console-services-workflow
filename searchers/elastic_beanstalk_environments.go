package searchers

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	aw "github.com/deanishe/awgo"
	"github.com/rkoval/alfred-aws-console-services-workflow/awsworkflow"
	"github.com/rkoval/alfred-aws-console-services-workflow/caching"
	"github.com/rkoval/alfred-aws-console-services-workflow/util"
)

type ElasticBeanstalkEnvironmentSearcher struct{}

func (s ElasticBeanstalkEnvironmentSearcher) Search(wf *aw.Workflow, query string, session *session.Session, forceFetch bool, fullQuery string) error {
	cacheName := util.GetCurrentFilename()
	instances := caching.LoadElasticbeanstalkEnvironmentDescriptionArrayFromCache(wf, session, cacheName, s.fetch, forceFetch, fullQuery)
	for _, instance := range instances {
		s.addToWorkflow(wf, query, session.Config, instance)
	}
	return nil
}

func (s ElasticBeanstalkEnvironmentSearcher) fetch(session *session.Session) ([]elasticbeanstalk.EnvironmentDescription, error) {
	svc := elasticbeanstalk.New(session)

	NextToken := ""
	environments := []elasticbeanstalk.EnvironmentDescription{}
	for {
		resp, err := svc.DescribeEnvironments(&elasticbeanstalk.DescribeEnvironmentsInput{
			MaxRecords: aws.Int64(1000), // get as many as we can
			NextToken:  aws.String(NextToken),
		})
		if err != nil {
			return nil, err
		}

		for i := range resp.Environments {
			environments = append(environments, *resp.Environments[i])
		}

		if resp.NextToken != nil {
			NextToken = *resp.NextToken
		} else {
			break
		}
	}

	return environments, nil
}

func (s ElasticBeanstalkEnvironmentSearcher) addToWorkflow(wf *aw.Workflow, query string, config *aws.Config, environment elasticbeanstalk.EnvironmentDescription) {
	title := *environment.EnvironmentName
	subtitle := util.GetElasticBeanstalkHealthEmoji(*environment.Health) + " " + *environment.EnvironmentId + " " + *environment.ApplicationName
	var page string
	if *environment.Status == elasticbeanstalk.EnvironmentStatusTerminated {
		// "dashboard" page does not exist for terminated instances
		page = "events"
	} else {
		page = "dashboard"
	}
	item := util.NewURLItem(wf, title).
		Subtitle(subtitle).
		Arg(fmt.Sprintf(
			"https://%s.console.aws.amazon.com/elasticbeanstalk/home?region=%s#/environment/%s?applicationName=%s&environmentId=%s",
			*config.Region,
			*config.Region,
			page,
			*environment.ApplicationName,
			*environment.EnvironmentId,
		)).
		Icon(awsworkflow.GetImageIcon("elasticbeanstalk"))

	if strings.HasPrefix(query, "e-") {
		item.Match(*environment.EnvironmentId)
	}
}
