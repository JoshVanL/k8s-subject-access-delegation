package ntp_client

import (
	"testing"
	"time"
)

func Test_avg(t *testing.T) {
	var times []time.Duration

	client := &NTPClient{}

	for i := 1; i <= 5; i++ {
		times = append(times, time.Duration(i))
	}

	offset, err := client.averageOffSet(times)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if offset != 3 {
		t.Errorf("expected offset=3 got=%d", offset)
	}

	offset, err = client.averageOffSet([]time.Duration{5, 0, 10})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if offset != 5 {
		t.Errorf("expected offset=5 got=%d", offset)
	}
}
