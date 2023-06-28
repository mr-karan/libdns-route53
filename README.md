Route53 for `libdns`
=======================

# Why a fork

This is a fork of https://github.com/libdns/route53. Main changes are in the following areas:

- Fallback to AWS SDK to handle the config. Fixes https://github.com/libdns/route53/issues/13
- Configurable option to wait for records to propogate to all Route53 servers. Fixes https://github.com/libdns/route53/issues/14
- Add a `New()` method for initialising the provider instead of calling `init` in each operation.
- Easily set multiple IPs in A/AAAA records. Not possible in the original repo.

This is exclusively and specifically designed for integrating with https://github.com/mr-karan/nomad-external-dns/. If you're looking for a general purpose Route53 provider, please use the original repo.

## Example

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/libdns/libdns"
	route53 "github.com/mr-karan/libdns-route53"
)

func main() {
	// Create a new AWS Route53 provider. The region is explicitly set to "ap-south-1".
	p, err := route53.NewProvider(context.Background(), route53.Opt{Region: "ap-south-1"})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	var (
		// Set the IP address to be assigned to the dummy record
		ip = "192.168.0.1"
		// Set the zone to be used. This is the domain name for which the DNS records will be set.
		zone = "test.internal."
	)

	// Use the provider to set a record on Route53. The libdns.Record struct describes a DNS record.
	// The name, value, and type are set for this record.
	_, err = p.SetRecords(ctx, zone, []libdns.Record{
		{
			Name:  "dummy", // The record's name
			Value: ip,      // The record's value, in this case an IP address
			Type:  "A",     // The record type, A for Address record
		},
	})

	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Set record with IP:", ip)
}
```

For a more complete example, see [the example directory](./example).

---

[![godoc reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/libdns/route53)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for AWS [Route53](https://aws.amazon.com/route53/).

## Authenticating

This package supports all the credential configuration methods described in the [AWS Developer Guide](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#specifying-credentials), such as `Environment Variables`, `Shared configuration files`, the `AWS Credentials file` located in `.aws/credentials`, and `Static Credentials`. You may also pass in static credentials directly (or via caddy's configuration).

The following IAM policy is a minimal working example to give `libdns` permissions to manage DNS records:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "",
            "Effect": "Allow",
            "Action": [
                "route53:ListResourceRecordSets",
                "route53:GetChange",
                "route53:ChangeResourceRecordSets"
            ],
            "Resource": [
                "arn:aws:route53:::hostedzone/ZABCD1EFGHIL",
                "arn:aws:route53:::change/*"
            ]
        },
        {
            "Sid": "",
            "Effect": "Allow",
            "Action": [
                "route53:ListHostedZonesByName",
                "route53:ListHostedZones"
            ],
            "Resource": "*"
        }
    ]
}
```
