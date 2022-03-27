package kafka

import (
	"encoding/json"

	sarama "github.com/Shopify/sarama"
	neo4j "github.com/neo4j/neo4j-go-driver/v4/neo4j"
	log "github.com/sirupsen/logrus"

	"github.com/daburch/ao_tools/ao_consumer/pkg/models"
	neo4j_util "github.com/daburch/ao_tools/ao_consumer/pkg/neo4j"
)

type Consumer struct {
	Session neo4j.Session
	Ready   chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.Ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		log.Info("Processing message: ", message.Offset)
		session.MarkMessage(message, "")

		match := models.CrystalLeagueMatch{}
		json.Unmarshal(message.Value, &match)

		transaction, err := consumer.Session.BeginTransaction()
		if err != nil {
			log.Panic("Unable to begin Neo4j transaction: ", err)
		}

		neo4j_util.ProcessMatch(transaction, match)
		if err := transaction.Commit(); err != nil {
			log.Panic("Unable to commit Neo4j transaction: ", err)
		}
	}

	return nil
}
