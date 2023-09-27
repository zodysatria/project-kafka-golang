package main

import (
	"fmt"

	"github.com/IBM/sarama"
)

func main() {

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{"kafka:9092"}, config)
	if err != nil {
		fmt.Printf("Gagal menyiapkan konsumen Kafka: %v\n", err)
		return
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			fmt.Printf("Gagal menutup konsumen Kafka: %v\n", err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition("tiket-topic", 0, sarama.OffsetOldest)
	if err != nil {
		fmt.Printf("Gagal menggunakan partisi: %v\n", err)
		return
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			fmt.Printf("Gagal menutup partisi konsumen: %v\n", err)
		}
	}()

	// membaca pesan dari kafka topic
	for message := range partitionConsumer.Messages() {
		fmt.Printf("Menerima pesan dari partisi %d with offset %d: %s\n", message.Partition, message.Offset, message.Value)
	}

	// mengatasi Kafka error
	for {
		select {
		case err := <-partitionConsumer.Errors():
			fmt.Printf("Kesalahan konsumen Kafka: %v", err)
		}
	}

}
