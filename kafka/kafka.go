//    Copyright 2021 AERIS-Consulting e.U.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package kafka

// https://github.com/confluentinc/confluent-kafka-go
import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/spf13/cobra"
	"http-to-kafka/datapublisher"
	"log"
	"strings"
)

const (
	defaultBootstrap = "localhost:9092"
	defaultTopic     = "http-request"
)

// configuration is the type describing the configuration for the Kafka client.
type configuration struct {
	bootstrap  string
	topic      string
	properties []string
}

type kafkaDataPublisher struct {
	*kafka.Producer
}

var (
	config = configuration{
		bootstrap:  defaultBootstrap,
		topic:      defaultTopic,
		properties: []string{},
	}
	publisher *kafkaDataPublisher
)

func (k kafkaDataPublisher) Publish(key []byte, message []byte) {
	k.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &config.topic},
		Key:            key,
		Value:          message,
	}, nil)
}

// InitCommand configures the command-line options for the Kafka client.
func InitCommand(rootCommand *cobra.Command) {
	rootCommand.PersistentFlags().StringVar(&(config.bootstrap), "kafka-bootstrap", defaultBootstrap, "bootstrap for the Kafka client")
	rootCommand.PersistentFlags().StringVar(&(config.topic), "kafka-topic", defaultTopic, "topic to produce the Kafka records to")
	rootCommand.PersistentFlags().StringSliceVar(&(config.properties), "kafka-configuration", []string{}, "general properties for the Kafka client, as key=value pairs")
}

func Start() {
	producerConfig := kafka.ConfigMap{}

	for _, configItem := range append(config.properties, "bootstrap.servers= "+config.bootstrap) {
		conf := strings.SplitN(strings.TrimSpace(configItem), "=", 2)
		if len(conf) < 2 {
			panic(fmt.Sprintf("The configuration %v is not a valid Kafka configuration", configItem))
		}
		producerConfig[strings.TrimSpace(conf[0])] = strings.TrimSpace(conf[1])
	}
	log.Printf("Configuration of the Kafka producer: %v\n", producerConfig)
	producer, err := kafka.NewProducer(&producerConfig)
	if err != nil {
		panic(err)
	}

	publisher = &kafkaDataPublisher{producer}

	// Delivery report handler for produced messages.
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed in partition %v: %v\n", ev.TopicPartition, ev.String())
				}
			}
		}
	}()

	datapublisher.RegisterPublisher(publisher)

}

func Stop() {
	log.Printf("Shutting down the Kafka producer\n")
	publisher.Flush(1000)
	publisher.Close()
}
