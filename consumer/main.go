package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	cluster "github.com/bsm/sarama-cluster"
	"github.com/golang/protobuf/proto"
	"github.com/vshiva/kafka-go-test/types"
)

func main() {

	var brokers types.KafkaBrokers
	flag.Var(&brokers, "broker", "Kafka Brokers")
	topic := flag.String("topic", "CONTAINERPIPE_tas_response", "Kafka Topic to which the messages needs to produced")
	cgroup := flag.String("group", "meteor_response_reader_tester", "Kafka Consumer group")

	flag.Parse()

	// init (custom) config, enable errors and notifications
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = false

	// init consumer
	//brokers := []string{"129.213.51.85:30191", "129.213.18.45:30191", "129.213.47.121:30191"}

	consumer, err := cluster.NewConsumer(brokers, *cgroup, []string{*topic}, config)
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
