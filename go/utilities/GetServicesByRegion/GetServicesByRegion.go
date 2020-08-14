package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/endpoints"
)

func main() {
	partition := endpoints.AwsPartition()
	regions := partition.Regions()

	for _, region := range regions {
		fmt.Println(region.ID())
		fmt.Println("")

		services := region.Services()
		for _, service := range services {
			fmt.Println("  " + service.ID())
		}

		fmt.Println("")
	}
}
