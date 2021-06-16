package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/vixus0/skuttle/v2/internal/logging"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
)

var (
	log *logging.Logger = logging.NewLogger("provider/aws")
)

type Provider struct {
	Client ec2.DescribeInstancesAPIClient
}

func NewProvider(ctx context.Context) (*Provider, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration, %v", err)
	}

	ec2client := ec2.NewFromConfig(cfg)

	// Do a dry run to check we have the right IAM permissions
	var apiErr smithy.APIError

	log.Info("performing ec2 dry run")
	_, err = ec2client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		DryRun: aws.Bool(true),
	})

	if err != nil {
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "DryRunOperation":
				log.Info("dry run successful")
			case "UnauthorizedOperation":
				log.Info("dry run failed")
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &Provider{
		Client: ec2client,
	}, nil
}

func (provider *Provider) InstanceExists(providerID string) (bool, error) {
	// should have been checked already, being defensive
	if !strings.HasPrefix(providerID, "aws:///") {
		return false, fmt.Errorf("providerID %s does not start with aws:///", providerID)
	}

	// assume EC2 instance ID is the last path segment of a provider ID
	parts := strings.Split(providerID, "/")
	instanceID := parts[len(parts)-1]

	out, err := provider.Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			instanceID,
		},
		Filters: []ec2types.Filter{
			{
				Name: aws.String("instance-state-name"),
				Values: []string{
					"running",
				},
			},
		},
	})

	// Deal with API errors
	if err != nil {
		var apiErr smithy.APIError

		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "InvalidInstanceID.NotFound":
				log.Info("no instance found for instance ID %s", instanceID)
				return false, nil
			default:
				log.Error("aws API error - code: %v, message: %v", apiErr.ErrorCode(), apiErr.ErrorMessage())
			}
		}

		return false, err
	}

	if len(out.Reservations) == 0 {
		log.Info("no reservations for instance ID %s", instanceID)
		return false, nil
	}

	return true, nil
}
