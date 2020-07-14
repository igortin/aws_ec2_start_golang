package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	// profileaName = os.Getenv("AWS_PROFILE")
	// awsRegion    = os.Getenv("AWS_REGION")
	profileaName        = "private"
	awsRegion           = "us-east-1"
	tagKey              = "tag:role"
	tagValue            = "etcd"
	instanceStateFilter = "stopped"
)

func main() {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: aws.String(awsRegion)},
		Profile: profileaName,
	})
	if err != nil {
		log.Println("Error: Can not create new session.")
	}

	// Get data from AWS service
	result, err := ec2GetResponse(sess)
	if err != nil {
		log.Println("Error: Can noot get data")
		return
	}

	// Check  length of array ec2.DescribeInstancesOutput.Reservations
	if len(result.Reservations) == 0 {
		log.Println("Error: ec2.DescribeInstancesOutput.Reservations is empty")
		return
	}

	// Func parse result to []*Ec2object
	ec2Instances, err := parseEc2Response(result)
	if err != nil {
		log.Println(err)
		return
	}

	ec2InstancesState, err := startEc2Instance(sess, ec2Instances)
	if err != nil {
		log.Println(err)
		return
	}

	for _, state := range ec2InstancesState {
		for _, change := range state.StartingInstances {
			fmt.Printf("INFO: Instance %s - started, previus state - %s", *change.InstanceId, *change.PreviousState.Name)
		}
	}
}

// Func to get data from AWS EC2 service
func ec2GetResponse(sess *session.Session) (*ec2.DescribeInstancesOutput, error) {

	// New client with a session
	ec2svc := ec2.New(sess)

	// Init instance request of structure DescribeInstancesInput
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String(tagKey),
				Values: []*string{
					aws.String(tagValue),
				},
			},
		},
	}

	// Return instance structure of DescribeInstancesOutput by Method
	response, err := ec2svc.DescribeInstances(input)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func parseEc2Response(data *ec2.DescribeInstancesOutput) ([]*Ec2object, error) {
	Ec2objectList := []*Ec2object{}
	// iterate []*Reservations
	for _, reservation := range data.Reservations {

		//  iterate []*Instance
		for _, instance := range reservation.Instances {
			//log.Printf("Procced number %d - instnace ID: %v\n", index + 1, *instance.InstanceId)

			object := Ec2object{
				InstanceID:     *instance.InstanceId,
				InstanceState:  *instance.State.Name,
				PrivateDNSName: *instance.PrivateDnsName,
				PublicDNSame:   *instance.PublicDnsName,
			}

			// Append to list of pointers a new pointer to object in memory
			Ec2objectList = append(Ec2objectList, &object)
		}
	}
	return Ec2objectList, nil
}

func startEc2Instance(sess *session.Session, instances []*Ec2object) (map[string]*ec2.StartInstancesOutput, error) {
	m := map[string]*ec2.StartInstancesOutput{}
	svc := ec2.New(sess)

	// Double check that instance in stopped state
	for _, instance := range instances {
		// Skip not stoppped state
		if instance.InstanceState != "stopped" {
			continue
		}

		instanceInput := &ec2.StartInstancesInput{
			DryRun:      aws.Bool(false),
			InstanceIds: []*string{&instance.InstanceID},
		}

		instanceOutput, err := svc.StartInstances(instanceInput)
		if err != nil {
			return m, err
		}

		m[instance.InstanceID] = instanceOutput
	}

	if len(m) == 0 {
		log.Println("INFO: stopped instances not found")
	}
	return m, nil
}
