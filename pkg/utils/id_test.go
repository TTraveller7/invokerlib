package utils

import "testing"

func TestNewWorkerId(t *testing.T) {
	type args struct {
		workerIndex   int
		processorName string
		topicName     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				workerIndex:   0,
				processorName: "splitter",
				topicName:     "word_count_source",
			},
			want: "wut",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewWorkerId(tt.args.workerIndex, tt.args.processorName, tt.args.topicName); got != tt.want {
				t.Errorf("NewWorkerId() = %v, want %v", got, tt.want)
			}
		})
	}
}
