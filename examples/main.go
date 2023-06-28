package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/libdns/libdns"
	route53 "github.com/mr-karan/libdns-route53"
)

func main() {
	p, err := route53.NewProvider(context.Background(),
		route53.Opt{
			Region: "ap-south-1", // libdns defaults to us-east-1 so this **must** be provided.
		})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	ips := []string{}

	for i := 0; i < 3; i++ {
		// Generate a random IP address
		ip := net.IPv4(
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
			byte(rand.Intn(256)),
		).String()

		ips = append(ips, ip)
	}

	_, err = p.SetRecords(ctx, "test.internal.", []libdns.Record{
		{
			Name:  "nomad-r53-debug",
			Value: fmt.Sprintf("%s,%s,%s", ips[0], ips[1], ips[2]),
			Type:  "A",
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Set record with IPs:", ips)
}
