package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/golang/protobuf/proto"
	"github.com/vshiva/kafka-go-test/types"
	"github.com/vshiva/kafka-go-test/utils"
)

var (
	brokers   = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The Kafka brokers to connect to, as a comma separated list")
	topic     = flag.String("topic", "CONTAINERPIPE_tas_response", "Kafka Topic to which the messages needs to produced")
	cgroup    = flag.String("group", "meteor_response_reader_tester", "Kafka Consumer group")
	caFile    = flag.String("kafka-cafile", "", "CA File")
	keyFile   = flag.String("kafka-keyfile", "", "Client Key File")
	certFile  = flag.String("kafka-certfile", "", "Client Cert File")
	verbose   = flag.Bool("verbose", false, "Turn on logging")
	verifySsl = flag.Bool("verify", false, "Optional verify ssl certificates chain")
)

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

	// init (custom) config, enable errors and notifications
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = false

	tlsConfig, err := utils.NewTLSConfig(*certFile, *keyFile, *caFile, !*verifySsl)
	if err != nil {
		panic(err)
	}

	if tlsConfig != nil {
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	// init consumer
	consumer, err := cluster.NewConsumer(brokerList, *cgroup, []string{*topic}, config)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	// trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// consume errors
	go func() {
		for err := range consumer.Errors() {
			log.Printf("Error: %s\n", err.Error())
		}
	}()

	// consume notifications
	go func() {
		for ntf := range consumer.Notifications() {
			log.Printf("Rebalanced: %+v\n", ntf)
		}
	}()

	// consume messages, watch signals
	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				tasMsg := &types.TasProtocol{}
				err = proto.Unmarshal(msg.Value, tasMsg)
				if err != nil {
					log.Printf("Error: %s\n", err.Error())
				}
				fmt.Printf("%+v\n", tasMsg)

				fmt.Fprintf(os.Stdout, "%s/%d/%d\t%s\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Key, tasMsg.AccountName)
				consumer.MarkOffset(msg, "") // mark message as processed
			}
		case <-signals:
			return
		}
	}
}
