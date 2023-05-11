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

func isPortOpen(ip string, port int, timeout int) bool {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func checkRegion(region string, profile string, port int, timeout int, wg *sync.WaitGroup) {
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
			panic("failed to describe instances")
		}

		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				if instance.State.Name == "running" {
					if isPortOpen(*instance.PublicIpAddress, port, timeout) {
						time := strings.Split(instance.LaunchTime.String(), " ")
						date := time[0]
						fmt.Printf("%s %s %s %s %s\n", *instance.PublicIpAddress, *instance.InstanceId, date, region, profile)
					}
				}
			}
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}
}

func main() {
	var profiles string
	var regions string
	var port int
	var timeout int

	// Define flags
	flag.StringVar(&profiles, "a", "5233,8055,4511stage", "List of profiles")
	flag.StringVar(&regions, "r", "ap-south-1,eu-north-1,eu-west-3,eu-west-2,eu-west-1,ap-northeast-3,ap-northeast-2,ap-northeast-1,ca-central-1,sa-east-1,us-east-1,us-east-2,us-west-1,us-west-2", "List of regions")
	flag.IntVar(&port, "p", 22, "Port number")
	flag.IntVar(&timeout, "t", 500, "Timeout in milliseconds")

	flag.Parse()

	// Replace ", " with "," to remove spaces in the input
	profiles = strings.ReplaceAll(profiles, ", ", ",")
	regions = strings.ReplaceAll(regions, ", ", ",")

	// Split profiles and regions into slices
	profileList := strings.Split(profiles, ",")
	regionList := strings.Split(regions, ",")

	var wg sync.WaitGroup

	fmt.Println("ip, id, created, region, profile")
	for _, profile := range profileList {
		for _, region := range regionList {
			wg.Add(1)
			go checkRegion(region, profile, port, timeout, &wg)
		}
	}

	wg.Wait()
}
