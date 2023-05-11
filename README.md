# AWS EC2 Port Scanner

This is a command-line tool written in Go to scan the status of TCP ports on AWS EC2 instances. It allows you to specify profiles, regions, and port numbers to scan for open ports.

## Prerequisites

- Go 1.16 or later
- AWS ./aws/credentials profiles configured with credentials

## Installation

1. Clone the repository:
```
git clone https://github.com/jessegersensonchess/aws-ec2-port-scanner.git
```

2. Change to the project directory:
```
cd aws-ec2-port-scanner
```

3. Build the Go binary:
```
go build -o aws-ec2-port-scanner main.go
```

## Usage
The tool accepts the following command-line flags:  

-a List of AWS profiles to scan (comma-separated).  
-r List of AWS regions to scan (comma-separated).  
-p Port number to scan (default is 22).  
-t Timeout, milliseconds to wait for a response from an IP

## Example usage:
```
./aws-ec2-port-scanner -a profile1,profile2 -r us-east-1,us-west-2 -p 22
```

This command will scan port 22 on all running EC2 instances in the specified profiles and regions. The tool will display the IP address, instance ID, creation date, region, and profile for instances with the port open.


## License
This project is licensed under the MIT License. See the LICENSE file for details.

## Contributing
Contributions are welcome! Feel free to open an issue or submit a pull request for any enhancements or bug fixes.
