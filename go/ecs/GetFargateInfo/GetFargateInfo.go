package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func debugPrint(debug *bool, s string) {
	if *debug {
		fmt.Println(s)
	}
}

func main() {
	debug := flag.Bool("d", false, "Whether to include additional debugging info")
	flag.Parse()

	debugPrint(debug, "Debugging enabled")

	// Get the regions
	partition := endpoints.AwsPartition()
	regions := partition.Regions()

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Look for clusters in all regions
	for _, region := range regions {
		fmt.Println("Looking for ECS clusters in " + region.ID())
		fmt.Println("")

		svc := ecs.New(sess, &aws.Config{
			Region: aws.String(region.ID()),
		})

		// Get clusters
		result, err := svc.DescribeClusters(nil)
		if err != nil {
			continue
		}

		// Look in each cluster
		for _, cluster := range result.Clusters {
			fmt.Println("Cluster name: " + *cluster.ClusterName)
		}

		fmt.Println("")
	}
}
