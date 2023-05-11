package main

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// Unit Test for isPortOpen function
func TestIsPortOpen(t *testing.T) {
	// Start a dummy TCP server on a specific port
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		t.Fatalf("failed to start dummy TCP server: %v", err)
	}
	defer listener.Close()

	// Run the isPortOpen function with the dummy server's address and port
	result := isPortOpen("localhost", 12345)

	// Assert that the port is open
	if !result {
		t.Error("expected port to be open, but got closed")
	}
}

// Benchmark Test for isPortOpen function
func BenchmarkIsPortOpen(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isPortOpen("example.com", 80)
	}
}

// Helper function to setup AWS client for testing
func setupAWSClient() (*ec2.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return ec2.NewFromConfig(cfg), nil
}

// Integration Test for checkRegion function
func TestCheckRegion(t *testing.T) {
	// Run checkRegion function
	checkRegion("us-west-2", "8055", 22)
}

type mockDescribeInstancesClient struct{}

func TestMain(m *testing.M) {
	// Setup any test-specific configuration here
	// e.g., initialize mock AWS services

	// Run the tests
	exitCode := m.Run()

	// Teardown any test-specific resources here
	// e.g., cleanup mock AWS services

	// Exit with the proper exit code
	os.Exit(exitCode)
}
