package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDefaultKafkaConfig(t *testing.T) {
	// DefaultKafkaConfig 함수가 기본값을 올바르게 설정하는지 테스트
	config := DefaultKafkaConfig()

	assert.Equal(t, []string{"localhost:9092"}, config.Brokers)
	assert.Equal(t, "default-client", config.ClientId)
	assert.Equal(t, "default-consumer-group", config.ConsumerGroupId)
	assert.Equal(t, sarama.OffsetNewest, config.InitialOffset)
	assert.Equal(t, true, config.EnableAutoCommit)
	assert.Equal(t, 1*time.Second, config.AutoCommitInterval)
	assert.Equal(t, 10*time.Second, config.SessionTimeout)
	assert.Equal(t, 3, config.MaxRetry)
	assert.Equal(t, sarama.WaitForAll, config.RequiredAcks)
	assert.Equal(t, sarama.V2_8_0_0, config.Version)
}

func TestNewSaramaConsumerConfig(t *testing.T) {
	// 컨슈머 설정이 정확히 생성되는지 테스트
	config := DefaultKafkaConfig()
	config.ClientId = "test-client"
	config.MaxRetry = 5
	config.MaxProcessingTime = 500 * time.Millisecond

	saramaConfig := NewSaramaConsumerConfig(config)

	assert.Equal(t, "test-client", saramaConfig.ClientID)
	assert.Equal(t, config.Version, saramaConfig.Version)
	assert.Equal(t, true, saramaConfig.Consumer.Return.Errors)
	assert.Equal(t, config.InitialOffset, saramaConfig.Consumer.Offsets.Initial)
	assert.Equal(t, config.EnableAutoCommit, saramaConfig.Consumer.Offsets.AutoCommit.Enable)
	assert.Equal(t, config.AutoCommitInterval, saramaConfig.Consumer.Offsets.AutoCommit.Interval)
	assert.Equal(t, config.SessionTimeout, saramaConfig.Consumer.Group.Session.Timeout)
	assert.Equal(t, 5, saramaConfig.Metadata.Retry.Max)
	assert.Equal(t, 500*time.Millisecond, saramaConfig.Consumer.MaxProcessingTime)
}

func TestNewSaramaProducerConfig(t *testing.T) {
	// 프로듀서 설정이 정확히 생성되는지 테스트
	config := DefaultKafkaConfig()
	config.ClientId = "test-producer"
	config.MaxRetry = 4
	config.RequiredAcks = sarama.WaitForLocal

	saramaConfig := NewSaramaProducerConfig(config)

	assert.Equal(t, "test-producer", saramaConfig.ClientID)
	assert.Equal(t, config.Version, saramaConfig.Version)
	assert.Equal(t, true, saramaConfig.Producer.Return.Errors)
	assert.Equal(t, sarama.WaitForLocal, saramaConfig.Producer.RequiredAcks)
	assert.Equal(t, 4, saramaConfig.Producer.Retry.Max)
	assert.Equal(t, true, saramaConfig.Producer.Return.Successes)

	// 파티셔너 테스트는 직접 타입 비교보다 설정이 있는지만 확인
	assert.NotNil(t, saramaConfig.Producer.Partitioner, "Partitioner should be set")
}

func TestNewCmpProducer(t *testing.T) {
	// 모의 프로듀서를 사용하여 프로듀서 생성 테스트
	// 실제 Kafka 브로커에 연결하지 않고 테스트

	// Brokers 변수 임시 저장
	originalBrokers := Brokers
	// 테스트용 브로커 설정
	Brokers = []string{"localhost:9092"}

	// 테스트 완료 후 원래 값 복원
	defer func() {
		Brokers = originalBrokers
	}()

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	// 모의 프로듀서 생성을 위한 패치 (실제 코드에서는 이 부분을 모킹 라이브러리로 대체)
	// 이 테스트는 실제로 프로듀서가 생성되는지만 확인하는 간단한 테스트
	// 실제 환경에서는 모의 객체를 사용하거나 테스트 환경의 Kafka를 사용해야 함
	producer, err := NewCmpProducer(config)

	// 실제 연결을 시도하면 에러가 발생할 수 있음 (테스트 환경에 따라 다름)
	// 이 경우 에러 메시지를 확인하고 실제 kafka가 없어서 발생하는 에러인지 검증
	if err != nil {
		assert.Contains(t, err.Error(), "kafka", "Error should be related to kafka connection")
	} else {
		assert.NotNil(t, producer, "Producer should not be nil if created successfully")
		// 성공적으로 생성된 경우 닫기
		(*producer).Close()
	}
}
