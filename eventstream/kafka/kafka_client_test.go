package kafka

import (
	"github.com/Shopify/sarama"
	"reflect"
	"testing"
)

func TestNewCmpProducer(t *testing.T) {
	type args struct {
		brokers  []string
		ackLevel sarama.RequiredAcks
	}
	tests := []struct {
		name    string
		args    args
		want    sarama.SyncProducer
		wantErr bool
	}{
		// TODO: Add test cases.
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCmpProducer(tt.args.brokers, tt.args.ackLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCmpProducer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCmpProducer() got = %v, want %v", got, tt.want)
			}
		})
	}
}
