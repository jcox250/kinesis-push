package main

import (
	"bufio"
	"flag"
	"fmt"
	stdlog "log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/jcox250/log"
)

var (
	streamName      string
	roleArn         string
	awsRegion       string
	kinesisEndpoint string
	debug           bool
)

func init() {
	flag.StringVar(&streamName, "stream-name", "", "name of the kinesis stream")
	flag.StringVar(&roleArn, "role-arn", "", "role to assume")
	flag.StringVar(&awsRegion, "aws-region", "", "aws region e.g us-east-1")
	flag.StringVar(&kinesisEndpoint, "kinesis-endpoint", "", "endpoint for kinesis, used for dev with localstack")
	flag.BoolVar(&debug, "debug", false, "enables debug logging")
	flag.Parse()
}

func main() {
	logger := log.NewLeveledLogger(os.Stderr, debug)

	logger.Info("msg", "kinesis config", "stream-name", streamName, "role-arn", roleArn, "aws-region", awsRegion, "endpoint", kinesisEndpoint)
	logger.Info("msg", "service config", "debug", debug)

	kc, err := NewKinesisClient(KinesisConfig{
		StreamName: streamName,
		RoleArn:    roleArn,
		AWSRegion:  awsRegion,
		Endpoint:   kinesisEndpoint,
		Logger:     logger,
	})
	if err != nil {
		stdlog.Fatalf("failed to create kinesis client: %s", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		logger.Debug("msg", "writing to kinesis...")
		if err := kc.Pub(scanner.Bytes()); err != nil {
			logger.Error("msg", "failed to publish to kinesis", "err", err)
		}
		logger.Debug("msg", "successfully wrote to kinesis", "data", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		logger.Debug("msg", "failed to read from stdin", "err", err)
	}
}

type KinesisConfig struct {
	StreamName string
	RoleArn    string
	AWSRegion  string
	Endpoint   string
	Logger     log.LeveledLogger
}

type KinesisClient struct {
	logger     log.LeveledLogger
	client     *kinesis.Kinesis
	streamName string
	shards     *StrSliceIterator
}

func NewKinesisClient(k KinesisConfig) (*KinesisClient, error) {
	s := session.Must(session.NewSession())
	config := &aws.Config{
		Region:   aws.String(k.AWSRegion),
		Endpoint: aws.String(k.Endpoint),
	}

	if k.RoleArn != "" {
		config.Credentials = stscreds.NewCredentials(s, k.RoleArn)
	}

	kc := &KinesisClient{
		logger:     k.Logger,
		client:     kinesis.New(s, config),
		streamName: k.StreamName,
	}

	shards, err := kc.getShards()
	if err != nil {
		return nil, err
	}
	kc.shards = NewStrSliceIterator(shards)
	return kc, nil
}

func (k *KinesisClient) Pub(b []byte) error {
	shard := k.shards.Next()
	k.logger.Debug("msg", "put_record", "shard", shard)
	_, err := k.client.PutRecord(&kinesis.PutRecordInput{
		Data:         b,
		PartitionKey: aws.String(shard),
		StreamName:   aws.String(k.streamName),
	})
	if err != nil {
		return err
	}
	return nil
}

func (k *KinesisClient) getShards() ([]string, error) {
	res, err := k.client.ListShards(&kinesis.ListShardsInput{
		StreamName: aws.String(k.streamName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list shards: %s", err)
	}

	shards := []string{}
	for _, s := range res.Shards {
		shards = append(shards, *s.ShardId)
	}
	return shards, nil
}

type StrSliceIterator struct {
	currIdx    int
	slice      []string
	calledOnce bool
}

func NewStrSliceIterator(s []string) *StrSliceIterator {
	return &StrSliceIterator{
		currIdx: 0,
		slice:   s,
	}
}

func (s *StrSliceIterator) Next() string {
	defer func() {
		s.currIdx++
	}()

	// Reset back to zero if we'lve incremented past the size of the slice
	if s.currIdx > len(s.slice)-1 {
		s.currIdx = 0
	}

	return s.slice[s.currIdx]
}
