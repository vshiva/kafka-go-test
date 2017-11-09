package main

import (
	"flag"
	"fmt"
	"math/rand"

	"github.com/Shopify/sarama"
	"github.com/vshiva/kafka-go-test/types"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getMessage(accountName string) *types.TasProtocol {
	msg := &types.TasProtocol{
		Uuid:        randomStr(4) + "-" + randomStr(4) + "-" + randomStr(4) + "-" + randomStr(4),
		Request:     "CREATE",
		AccountName: accountName,
		ServiceType: "Wercker",
		ServiceMessage: &types.ServiceMessage{
			AccountId: randomStr(4) + "-" + randomStr(4) + "-" + randomStr(4) + "-" + randomStr(4),
			Operation: "somerandomsh",
			Region:    "us-phoenix-1",
			AdminInformation: &types.AdminInformation{
				AdminUsername:  randomStr(6),
				AdminPassword:  randomStr(6),
				AdminEmail:     randomStr(6) + "." + randomStr(8) + "@" + randomStr(3),
				AdminFirstName: randomStr(6),
				AdminLastName:  randomStr(6),
			},
			Resources: []*types.Resource{
				&types.Resource{
					Name:            "registry",
					PurchasedAmount: 10.99,
					StartDate:       "1997-07-16T19:20:30.45+01:00",
					EndDate:         "2197-07-16T19:20:30.45+01:00",
				},
			},
			Property: []*types.ServicePropertyBean{
				&types.ServicePropertyBean{
					Key:   "key1",
					Value: "value1",
				},
			},
		},
	}

	return msg
}

func main() {

	var brokers types.KafkaBrokers
	flag.Var(&brokers, "broker", "Kafka Brokers")
	topic := flag.String("topic", "CONTAINERPIPE_tas", "Kafka Topic to which the messages needs to produced")
	accountName := flag.String("accountName", randomStr(6), "Kafka Topic to which the messages needs to produced")

	flag.Parse()

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	// brokers := []string{"129.213.51.85:30191", "129.213.18.45:30191", "129.213.47.121:30191"}

	// brokers := []string{"kafka:9092"}
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		// Should not reach here. Yet we are so fail.
		panic(err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			// Should not reach here
			panic(err)
		}
	}()

	msgValue := getMessage(*accountName)

	msg := &sarama.ProducerMessage{
		Topic: *topic,
		Value: types.ProtobufEncoder{Msg: msgValue},
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", *topic, partition, offset)
}
