package aws_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vixus0/skuttle/v2/internal/provider/aws"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
)

const (
	runningID    = "i-0123abcdef"
	terminatedID = "i-deadbeef69"
	missingID    = "i-aabbccdd00"
	errorID      = "i-xxxxxxxx"
)

var _ = Describe("AWS Provider", func() {
	var (
		provider *aws.Provider
	)

	BeforeEach(func() {
		provider = &aws.Provider{
			Client: &MockEC2Client{
				instances: []*MockInstance{
					{
						ID:    runningID,
						State: "running",
					},
					{
						ID:    terminatedID,
						State: "terminated",
					},
				},
			},
		}
	})

	Describe("Checking if an EC2 instance exists", func() {
		It("should be true for running instances", func() {
			providerID := fmt.Sprintf("aws:///region/%s", runningID)
			exists, err := provider.InstanceExists(providerID)

			Expect(err).To(BeNil())
			Expect(exists).To(BeTrue())
		})

		It("should not be true for stopped or terminated instances", func() {
			providerID := fmt.Sprintf("aws:///region/%s", terminatedID)
			exists, err := provider.InstanceExists(providerID)

			Expect(err).To(BeNil())
			Expect(exists).To(BeFalse())
		})

		It("should not be true when the instance was not found", func() {
			providerID := fmt.Sprintf("aws:///region/%s", missingID)
			exists, err := provider.InstanceExists(providerID)

			Expect(err).To(BeNil())
			Expect(exists).To(BeFalse())
		})

		It("should propagate errors", func() {
			providerID := fmt.Sprintf("aws:///region/%s", errorID)
			exists, err := provider.InstanceExists(providerID)

			Expect(err).To(HaveOccurred())
			Expect(exists).To(BeFalse())
		})
	})
})

type MockInstance struct {
	ID    string
	State string
}

type MockEC2Client struct {
	ec2.DescribeInstancesAPIClient
	instances []*MockInstance
}

func (c *MockEC2Client) DescribeInstances(ctx context.Context, input *ec2.DescribeInstancesInput, fn ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	var reservations []ec2types.Reservation

	// We only ever search for one instance ID at a time
	id := input.InstanceIds[0]

	switch id {
	case missingID:
		return nil, &smithy.GenericAPIError{
			Code:    "InvalidInstanceID.NotFound",
			Message: "Mock instance not found error",
			Fault:   smithy.FaultServer,
		}
	case errorID:
		return nil, &smithy.GenericAPIError{
			Code:    "SomeError",
			Message: "Mock error",
			Fault:   smithy.FaultUnknown,
		}
	}

	// Assuming one reservation for a matching instance ID
	for _, instance := range c.instances {
		if instance.ID == id && instance.State == "running" {
			reservations = append(reservations, ec2types.Reservation{
				Instances: []ec2types.Instance{
					{
						InstanceId: awssdk.String(id),
					},
				},
			})
		}
	}

	return &ec2.DescribeInstancesOutput{
		Reservations: reservations,
	}, nil
}
