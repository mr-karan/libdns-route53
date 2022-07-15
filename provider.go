package route53

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	r53 "github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/libdns/libdns"
)

type Opt struct {
	Region             string        `json:"region,omitempty"`
	MaxRetries         int           `json:"max_retries,omitempty"`
	MaxWaitDur         time.Duration `json:"max_wait_dur,omitempty"`
	WaitForPropogation bool          `json:"wait_for_propogation,omitempty"`
}

// Provider implements the libdns interfaces for Route53
type Provider struct {
	client *r53.Client
	opt    Opt
}

func NewProvider(ctx context.Context, opt Opt) (*Provider, error) {
	if opt.MaxRetries == 0 {
		opt.MaxRetries = 5
	}
	if opt.Region == "" {
		opt.Region = "ap-south-1"
	}
	if opt.MaxWaitDur == 0 && opt.WaitForPropogation {
		opt.MaxWaitDur = time.Minute * 1
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(opt.Region),
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(retry.NewStandard(), opt.MaxRetries)
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to load configuration, %w", err)
	}

	return &Provider{
		client: r53.NewFromConfig(cfg),
		opt:    opt,
	}, nil
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	zoneID, err := p.getZoneID(ctx, zone)
	if err != nil {
		return nil, err
	}

	records, err := p.getRecords(ctx, zoneID, zone)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zoneID, err := p.getZoneID(ctx, zone)
	if err != nil {
		return nil, err
	}

	var createdRecords []libdns.Record

	for _, record := range records {
		newRecord, err := p.createRecord(ctx, zoneID, record, zone)
		if err != nil {
			return nil, err
		}
		createdRecords = append(createdRecords, newRecord)
	}

	return createdRecords, nil
}

// DeleteRecords deletes the records from the zone. If a record does not have an ID,
// it will be looked up. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zoneID, err := p.getZoneID(ctx, zone)
	if err != nil {
		return nil, err
	}

	var deletedRecords []libdns.Record

	for _, record := range records {
		deletedRecord, err := p.deleteRecord(ctx, zoneID, record, zone)
		if err != nil {
			return nil, err
		}
		deletedRecords = append(deletedRecords, deletedRecord)
	}

	return deletedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zoneID, err := p.getZoneID(ctx, zone)
	if err != nil {
		return nil, err
	}

	var updatedRecords []libdns.Record

	for _, record := range records {
		updatedRecord, err := p.updateRecord(ctx, zoneID, record, zone)
		if err != nil {
			return nil, err
		}
		updatedRecords = append(updatedRecords, updatedRecord)
	}

	return updatedRecords, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
