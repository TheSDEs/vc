package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
	"vc/internal/mockas/apiv1"
	"vc/pkg/logger"
	"vc/pkg/model"
	"vc/pkg/trace"
)

//TODO: använda event, message eller record som begrepp? - message kallas det av sarama?

type EventConsumer struct {
	config *model.Cfg
	logger *logger.Log
	apiv1  *apiv1.Client
	tp     *trace.Tracer
	ctx    *context.Context
}

func NewEventConsumer(ctx *context.Context, config *model.Cfg, api *apiv1.Client, tracer *trace.Tracer, logger *logger.Log) (*EventConsumer, error) {
	logger.Info("Kafka enabled. Starting consumer ...")
	ec := &EventConsumer{
		config: config,
		logger: logger,
		apiv1:  api,
		tp:     tracer,
		ctx:    ctx,
	}
	err := ec.start()
	if err != nil {
		return nil, err
	}
	return ec, nil
}

func (ec *EventConsumer) start() error {
	//TODO: read config from file
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	config.Net.SASL.Enable = false //TODO: Activate SASL-auth when needed
	// ... (övriga säkerhetskonfigurationer)

	brokers := []string{"kafka0:9092", "kafka1:9092"}
	consumerGroup, err := sarama.NewConsumerGroup(brokers, "my-consumer-group-name-1", config)
	if err != nil {
		ec.logger.Error(err, "Failed to create Kafka consumer group")
		return err

	}
	defer consumerGroup.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for err := range consumerGroup.Errors() {
			ec.logger.Error(err, "Kafka consumer group error")
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	handler := &consumerGroupHandler{
		config: ec.config,
		logger: ec.logger,
		apiv1:  ec.apiv1,
		tp:     ec.tp,
		ctx:    ec.ctx,
	}

	topics := []string{"topic_mock_next"}

	go func() {
		attempt := 0
		for {
			if err := consumerGroup.Consume(ctx, topics, handler); err != nil {
				ec.logger.Error(err, "Error consuming from Kafka - Using exponential backoff strategy for consumtion")
				// A simple form of "throttling with exponential backoff" which limits how quickly new attempts are made. This can help avoid overwhelming the Kafka cluster or network if many errors occur in a short period.
				delay := exponentialBackoff(attempt)
				time.Sleep(delay)
				attempt++
			} else {
				attempt = 0
				ec.logger.Info("Kafka consumtion now back in normal consumtion strategy")
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	ec.logger.Info("Kafka consumer started")

	<-signals
	ec.logger.Info("Received termination signal, shutting down gracefully...")
	cancel()
	time.Sleep(20 * time.Second) // Allow time for shutdown
	return nil
}

func exponentialBackoff(attempt int) time.Duration {
	min := float64(100)            // Minimum delay in milliseconds
	max := float64(10 * 60 * 1000) // Maximum delay in milliseconds (10 minutes)
	factor := 2.0                  // Backoff factor

	delay := min * math.Pow(factor, float64(attempt))
	delay = delay + rand.Float64()*min

	if delay > max {
		delay = max
	}

	return time.Duration(delay) * time.Millisecond
}

func (ec *EventConsumer) Close(ctx context.Context) error {
	//TODO: clean up what ever needs to be cleaned up
	return nil
}

type consumerGroupHandler struct {
	config *model.Cfg
	logger *logger.Log
	apiv1  *apiv1.Client
	tp     *trace.Tracer
	ctx    *context.Context
}

// Setup is called before the session starts, prior to message consumption.
func (cgh *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup is called after the session ends, after message consumption has stopped.
func (cgh *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim must start a consumption loop. The code in this method should keep consuming messages
// from the claim.Messages() channel and call session.MarkMessage for each message consumed.
func (cgh *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		logMessageInfo(message)

		var mockNextRequest apiv1.MockNextRequest
		if err := json.Unmarshal(message.Value, &mockNextRequest); err != nil {
			cgh.logger.Error(err, "Failed to unmarshal event from Kafka")
			//TODO replace with cgh.handleErrorMessage(session, message, err)
			continue
		}

		cgh.logger.Debug("mockNextRequest:", "", mockNextRequest)

		_, err := cgh.apiv1.MockNext(*cgh.ctx, &mockNextRequest)
		if err != nil {
			cgh.logger.Error(err, "Failed to mock next")
			//TODO handleErrorMessage(session, message, err) and send to topic_mock_next_error
		}

		session.MarkMessage(message, fmt.Sprintf("GoID %d marked message as treated by handler", getGoroutineID()))
	}
	return nil
}

func logMessageInfo(message *sarama.ConsumerMessage) {
	//TODO: change to debug logging
	fmt.Println("Go routine ID:", getGoroutineID())
	fmt.Println("Raw message:", message)
	fmt.Printf("Raw message.Value: %v\n", message.Value)
	fmt.Printf("Message value as string: %s\n", string(message.Value))
	fmt.Printf("Message metadata Partition: %d, Offset: %d, Timestamp: %s, Topic: %s\n",
		message.Partition, message.Offset, message.Timestamp, message.Topic)

	fmt.Println("Headers:")
	headers := make(map[string]string)
	for _, header := range message.Headers {
		headers[string(header.Key)] = string(header.Value)
	}
	log.Println("Headers:")
	for key, value := range headers {
		fmt.Printf("  %s: %s\n", key, value)
	}

	if len(message.Key) > 0 {
		fmt.Printf("Message Key: %s\n", string(message.Key))
	}
}

// getGoroutineID Debug function to get the current id for the go routine
func getGoroutineID() int {
	stackBuf := make([]byte, 64)
	stackBuf = stackBuf[:runtime.Stack(stackBuf, false)]
	firstLine := bytes.SplitN(stackBuf, []byte("\n"), 2)[0]
	fields := bytes.Fields(firstLine)
	id, err := strconv.Atoi(string(fields[1]))
	if err != nil {
		return -1
	}
	return id
}

//TODO: one possible solution when failing to processes a consumed message
//func (cgh *consumerGroupHandler) handleErrorMessage(session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage, err error) {
//	retries := 0
//
//	for retries < cgh.retryLimit {
//		time.Sleep(cgh.retryBackoff)
//		retries++
//
//		// Retry message processing
//		var mockNextRequest apiv1.MockNextRequest
//		if err := json.Unmarshal(message.Value, &mockNextRequest); err != nil {
//			cgh.logger.Errorf("Retry %d: Failed to unmarshal event: %v", retries, err)
//			continue
//		}
//
//		_, err := cgh.apiv1.MockNext(nil, &mockNextRequest)
//		if err != nil {
//			cgh.logger.Errorf("Retry %d: Failed to mock next: %v", retries, err)
//			continue
//		}
//
//		// Successful processing, mark the message
//		session.MarkMessage(message, "")
//		return
//	}
//
//	// If retries exceeded, send to DLQ
//	cgh.sendToDLQ(message)
//}
//
//func (cgh *consumerGroupHandler) sendToDLQ(message *sarama.ConsumerMessage) {
//	// Implement DLQ logic here, e.g., producing to a DLQ topic
//	dlqProducer, err := sarama.NewSyncProducer([]string{cgh.config.KafkaBrokers}, nil)
//	if err != nil {
//		cgh.logger.Errorf("Failed to create DLQ producer: %v", err)
//		return
//	}
//	defer dlqProducer.Close()
//
//	dlqMessage := &sarama.ProducerMessage{
//		Topic: "topic_mock_next_error",
//		Key:   sarama.ByteEncoder(message.Key),
//		Value: sarama.ByteEncoder(message.Value),
//		//TODO add error info in message header?
//	}
//
//	_, _, err = dlqProducer.SendMessage(dlqMessage)
//	if err != nil {
//		cgh.logger.Error(err,"Failed to send message to DLQ")
//	}
//}

//TODO: möjlig hantering av omprövning från error topic, sätt även time to live så att ex. ett event tas bort säg efter 1 dygn eller liknande?
//for {
//// Tidpunkt A
//startTime := time.Now()
//
//// Hämta alla olästa meddelanden
//for message := range claim.Messages() {
//// ... (bearbeta meddelande) ...
//session.MarkMessage(message, "") // Markera meddelandet som läst
//}
//
//// Beräkna återstående tid till nästa timme
//timeToNextHour := time.Hour - time.Since(startTime)
//
//// Vänta tills nästa timme
//if timeToNextHour > 0 {
//time.Sleep(timeToNextHour)
//}
//}