package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"math"
	"time"
)

var Brokers = []string{"172.20.10.92:32000"}

// var Brokers = []string{"127.0.0.1:9092", "127.0.0.1:9093", "127.0.0.1:9094"}
var Topic = "k8s-meta-topic"
var RetryCount = 3

type KafkaConfig struct {
	Brokers            []string
	ClientId           string
	ConsumerGroupId    string
	Topics             []string
	InitialOffset      int64
	EnableAutoCommit   bool // 필요한가?
	AutoCommitInterval time.Duration
	SessionTimeout     time.Duration
	MaxRetry           int
	MaxProcessingTime  time.Duration
	RequiredAcks       sarama.RequiredAcks
	Version            sarama.KafkaVersion // 디폴트 버전?

}

func DefaultKafkaConfig() KafkaConfig {
	return KafkaConfig{
		Brokers:            []string{"localhost:9092"},
		ClientId:           "default-client",
		ConsumerGroupId:    "default-consumer-group",
		Topics:             []string{},
		InitialOffset:      sarama.OffsetNewest,
		EnableAutoCommit:   true,
		AutoCommitInterval: 1 * time.Second,
		SessionTimeout:     10 * time.Second,
		MaxRetry:           3,
		RequiredAcks:       sarama.WaitForAll,
		Version:            sarama.V2_8_0_0, //fixme: 사용하는 kafka 버전 확인 필요
	}
}

func NewSaramaConsumerConfig(config KafkaConfig) *sarama.Config {
	saramaConfig := sarama.NewConfig()

	// 클라이언트 설정
	saramaConfig.ClientID = config.ClientId
	saramaConfig.Version = config.Version

	// 컨슈머 설정
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = config.InitialOffset
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = config.EnableAutoCommit
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = config.AutoCommitInterval
	saramaConfig.Consumer.Group.Session.Timeout = config.SessionTimeout
	// Consumer 설정 추가
	saramaConfig.Metadata.Retry.Max = config.MaxRetry
	saramaConfig.Consumer.MaxProcessingTime = config.MaxProcessingTime

	return saramaConfig
}

// NewSaramaProducerConfig는 Sarama 프로듀서 설정을 반환합니다
func NewSaramaProducerConfig(config KafkaConfig) *sarama.Config {
	saramaConfig := sarama.NewConfig()

	// 클라이언트 설정
	saramaConfig.ClientID = config.ClientId
	saramaConfig.Version = config.Version

	// 프로듀서 설정
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.RequiredAcks = config.RequiredAcks
	saramaConfig.Producer.Retry.Max = config.MaxRetry

	// 파티션 선택 전략 (기본: 해시 기반 파티셔닝)
	saramaConfig.Producer.Partitioner = sarama.NewHashPartitioner

	return saramaConfig
}

func NewCmpProducer(brokers []string, ackLevel sarama.RequiredAcks) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = ackLevel // 리더의 ack 만 받는상태
	config.Producer.Retry.Max = RetryCount  // 재시도 횟수
	config.Producer.Retry.BackoffFunc = func(retries, maxRetries int) time.Duration {
		backoff := 100 * time.Millisecond * time.Duration(math.Pow(2, float64(retries)))
		if backoff > 10*time.Second {
			return 10 * time.Second
		}
		return backoff
	}
	config.Producer.Return.Successes = true // 성공 응답 받기 위해 필요함

	// 선택적인 추가 설정 필요하다면
	// config.Producer.Compression = sarama.CompressionSnappy // 압축 방식
	// config.Producer.Flush.Frequency = 500 * time.Millisecond // 배치 간격
	// config.Producer.Partitioner = sarama.NewRandomPartitioner // 파티셔닝 전략
	// config.Net.MaxOpenRequests = 10 // 브로커당 최대 10개의 동시요청 허용

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("error creating producer with config: %v", err)
	}

	return producer, nil
}
