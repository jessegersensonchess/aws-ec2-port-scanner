package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type InstanceInfo struct {
	PublicIP       string
	InstanceID     string
	Date           string
	Region         string
	Profile        string
	Name           string
	SecurityGroups []string
}

func checkRegion(region string, profile string, port int, timeout int, wg *sync.WaitGroup, resultChan chan<- InstanceInfo) {
	defer wg.Done()

	if len(region) <= 0 {
		return
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	)
	if err != nil {
		panic("failed to load AWS configuration")
	}

	client := ec2.NewFromConfig(cfg)

	// Retrieve instances in batches
	batchSize := int32(100) // Number of instances to retrieve per batch

	var nextToken *string
	for {
		output, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
			MaxResults: &batchSize,
			NextToken:  nextToken,
		})
		if err != nil {
			fmt.Println("ERROR: failed to describe instances")
			return
		}

		if len(output.Reservations) == 0 {
			// No instances found in the region, return early
			return
		}

		mychan := make(chan InstanceInfo)
		var wgInstances sync.WaitGroup

		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				if instance.State.Name == "running" {
					wgInstances.Add(1)

					instanceId := *instance.InstanceId
					publicIpAddress := *instance.PublicIpAddress
					launchTime := instance.LaunchTime.String()

					// check all IPs concurrently
					go func(instanceId string, publicIpAddress string, launchTime string) {
						defer wgInstances.Done()

						if isPortOpen(publicIpAddress, port, timeout) {
							time := strings.Split(launchTime, " ")
							date := time[0]
							name, err := getInstanceName(instanceId, client)
							if err != nil {
								name = "no name"
								panic("failed to get instance name")
							}

							securityGroupNames, err := getSecurityGroupNames(instanceId, client)
							if err != nil {
								panic("failed to get security group names")
							}

							instanceInfo := InstanceInfo{
								PublicIP:       publicIpAddress,
								InstanceID:     instanceId,
								Date:           date,
								Region:         region,
								Profile:        profile,
								Name:           name,
								SecurityGroups: securityGroupNames,
							}

							mychan <- instanceInfo
						}

					}(instanceId, publicIpAddress, launchTime)

				}
			}
		}

		go func() {
			wgInstances.Wait()
			close(mychan)
		}()

		for instanceInfo := range mychan {
			resultChan <- instanceInfo
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}
}

func getSecurityGroupNames(instanceID string, client *ec2.Client) ([]string, error) {
	output, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return nil, err
	}

	if len(output.Reservations) > 0 && len(output.Reservations[0].Instances) > 0 {
		var securityGroupNames []string
		for _, sg := range output.Reservations[0].Instances[0].SecurityGroups {
			securityGroupNames = append(securityGroupNames, *sg.GroupName)
		}
		return securityGroupNames, nil
	}

	return nil, nil
}

func getInstanceName(instanceID string, client *ec2.Client) (string, error) {
	output, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", err
	}

	if len(output.Reservations) > 0 && len(output.Reservations[0].Instances) > 0 && len(output.Reservations[0].Instances[0].Tags) > 0 {
		return *output.Reservations[0].Instances[0].Tags[0].Value, nil
	}

	return "", nil
}

func isPortOpen(ip string, port int, timeout int) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	resultChan := make(chan bool)

	go func() {
		conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Millisecond)
		if err != nil {
			resultChan <- false
			return
		}
		defer conn.Close()
		resultChan <- true
	}()

	select {
	case result := <-resultChan:
		return result
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return false
	}
}

func main() {
	var profiles string
	var regions string
	var port int
	var timeout int

	// Define flags
	flag.StringVar(&profiles, "a", "5233,8055,4511stage", "List of profiles")
	flag.StringVar(&regions, "r", "ap-south-1,eu-north-1,eu-west-3,eu-west-2,eu-west-1,ap-northeast-3,ap-northeast-2,ap-northeast-1,ca-central-1,sa-east-1,ap-southeast-1,ap-southeast-2,eu-central-1,us-east-1,us-east-2,us-west-1,us-west-2", "List of regions")
	flag.IntVar(&port, "p", 22, "Port number")
	flag.IntVar(&timeout, "t", 1000, "Timeout in milliseconds")

	flag.Parse()

	// Split profiles and regions into slices
	profileList := strings.Split(profiles, ",")
	regionList := strings.Split(regions, ",")

	var wg sync.WaitGroup
	resultChan := make(chan InstanceInfo)

	fmt.Println("ip, id, created, region, profile, security groups")

	for _, profile := range profileList {
		for _, region := range regionList {
			wg.Add(1)
			go checkRegion(region, profile, port, timeout, &wg, resultChan)
		}
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()
	for instanceInfo := range resultChan {
		output := fmt.Sprintf("%-15s %-15s %-15s %-15s %-15s %-15s (%s)",
			instanceInfo.PublicIP, instanceInfo.InstanceID, instanceInfo.Date, instanceInfo.Region, instanceInfo.Profile, instanceInfo.Name, strings.Join(instanceInfo.SecurityGroups, ", "))
		fmt.Println(output)
	}
}
