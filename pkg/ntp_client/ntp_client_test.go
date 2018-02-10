package ntp_client

import (
	"testing"
	"time"
)

func Test_avg(t *testing.T) {
	var times []time.Duration

	for i := 1; i <= 5; i++ {
		times = append(times, time.Duration(i))
	}

	foo := &NTPClient{}
	offset, err := foo.averageOffSet(times)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if offset != 3 {
		t.Errorf("expected offset=3 got=%d", offset)
	}

}
