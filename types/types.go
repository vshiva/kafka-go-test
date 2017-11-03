package types

import "strings"

// KafkaBrokers - List of Kafka Brokers
type KafkaBrokers []string

func (i *KafkaBrokers) String() string {
	return strings.Join(*i, " ")
}

// Set Value of Broker in the array
func (i *KafkaBrokers) Set(value string) error {
	*i = append(*i, value)
	return nil
}
