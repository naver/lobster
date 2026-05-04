package uploader

import (
	"errors"
	"testing"

	"github.com/IBM/sarama"
)

func TestFormatProducerErrors(t *testing.T) {
	pe := func(msgs ...string) sarama.ProducerErrors {
		errs := make(sarama.ProducerErrors, len(msgs))
		for i, m := range msgs {
			errs[i] = &sarama.ProducerError{Err: errors.New(m)}
		}
		return errs
	}

	tests := []struct {
		input sarama.ProducerErrors
		want  string
	}{
		{pe("timeout"), "ProducerErrors(1): timeout"},
		{pe("timeout", "timeout", "timeout"), "ProducerErrors(3): timeout"},
		{pe("err A", "err B", "err C"), "ProducerErrors(3): err A; err B; err C"},
		{pe("err A", "err B", "err A", "err B", "err A"), "ProducerErrors(5): err A; err B"},
	}

	for _, tt := range tests {
		got := formatProducerErrors(tt.input).Error()
		if got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}
