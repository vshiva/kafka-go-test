package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/vshiva/kafka-go-test/types"
	"github.com/vshiva/kafka-go-test/utils"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	brokers     = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The Kafka brokers to connect to, as a comma separated list")
	topic       = flag.String("topic", "CONTAINERPIPE_tas", "Kafka Topic to which the messages needs to produced")
	accountName = flag.String("account-name", randomStr(8), "Account Name")
	caFile      = flag.String("kafka-cafile", "", "CA File")
	keyFile     = flag.String("kafka-keyfile", "", "Client Key File")
	certFile    = flag.String("kafka-certfile", "", "Client Cert File")
	verbose     = flag.Bool("verbose", false, "Turn on logging")
	verifySsl   = flag.Bool("verify", false, "Optional verify ssl certificates chain")
)

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

	flag.Parse()
	if *verbose {
		sarama.Logger = log.New(os.Stdout, "[kafka-test] ", log.LstdFlags)
	}

	if *brokers == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	brokerList := strings.Split(*brokers, ",")
	log.Printf("Kafka brokers: %s", strings.Join(brokerList, ", "))

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	tlsConfig, err := utils.NewTLSConfig(*certFile, *keyFile, *caFile, !*verifySsl)
	if err != nil {
		panic(err)
	}

	if tlsConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	producer, err := sarama.NewSyncProducer(brokerList, config)
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
