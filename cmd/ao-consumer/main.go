package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/daburch/ao_tools/ao_consumer/pkg/kafka"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	log "github.com/sirupsen/logrus"
)

const (
	// kafka
	VERSION             = "VERSION"
	BROKER              = "BROKER"
	CRYSTAL_MATCH_TOPIC = "CRYSTAL_MATCH_TOPIC"
	CONSUMER_GROUP      = "CONSUMER_GROUP"

	// neo4j
	NEO4J_URL      = "NEO4J_URL"
	NEO4J_USERNAME = "NEO4J_USERNAME"
	NEO4J_PASSWORD = "NEO4J_PASSWORD"

	MISSING_ENV_VAR_ERR = "missing environment variable: %s"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableQuote: true,
	})
}

func main() {

	godotenv.Load()

	topics, ok := os.LookupEnv(CRYSTAL_MATCH_TOPIC)
	if !ok {
		log.Panicf(MISSING_ENV_VAR_ERR, CRYSTAL_MATCH_TOPIC)
	}

	url, ok := os.LookupEnv(NEO4J_URL)
	if !ok {
		log.Panicf(MISSING_ENV_VAR_ERR, NEO4J_PASSWORD)
	}

	un, ok := os.LookupEnv(NEO4J_USERNAME)
	if !ok {
		log.Panicf(MISSING_ENV_VAR_ERR, NEO4J_USERNAME)
	}

	pw, ok := os.LookupEnv(NEO4J_PASSWORD)
	if !ok {
		log.Panicf(MISSING_ENV_VAR_ERR, NEO4J_PASSWORD)
	}

	group, ok := os.LookupEnv(CONSUMER_GROUP)
	if !ok {
		log.Panicf(MISSING_ENV_VAR_ERR, CONSUMER_GROUP)
	}

	brokers, ok := os.LookupEnv(BROKER)
	if !ok {
		log.Panicf(MISSING_ENV_VAR_ERR, BROKER)
	}

	v, ok := os.LookupEnv(VERSION)
	if !ok {
		log.Panicf(MISSING_ENV_VAR_ERR, VERSION)
	}

	version, err := sarama.ParseKafkaVersion(v)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	config := sarama.NewConfig()
	config.Version = version
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	driver, err := neo4j.NewDriver(url, neo4j.BasicAuth(un, pw, ""))
	if err != nil {
		log.Panic("Unable to create Neo4j driver", err)
	}
	defer driver.Close()

	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	consumer := kafka.Consumer{
		Session: session,
		Ready:   make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), group, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, strings.Split(topics, ","), &consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.Ready = make(chan bool)
		}
	}()

	// Wait for consumer to become ready
	<-consumer.Ready
	log.Info("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	keepRunning := true
	for keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			log.Println("terminating: via signal")
			keepRunning = false
		}
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}
