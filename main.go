package main

import (
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
	svc := ec2.New(sess)

	// Get data from AWS service
	result, err := ec2GetResponse(svc)
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

	ec2InstancesState, err := startEc2Instance(svc, ec2Instances)
	if err != nil {
		log.Println(err)
		return
	}

	for _, state := range ec2InstancesState {
		for _, change := range state.StartingInstances {
			log.Printf("INFO: Instance %s - started, previus state - %s", *change.InstanceId, *change.PreviousState.Name)
		}
	}
}

// Func to get data from AWS EC2 service
func ec2GetResponse(svc *ec2.EC2) (*ec2.DescribeInstancesOutput, error) {

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
	response, err := svc.DescribeInstances(input)
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

func startEc2Instance(svc *ec2.EC2, instances []*Ec2object) (map[string]*ec2.StartInstancesOutput, error) {
	m := map[string]*ec2.StartInstancesOutput{}

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
