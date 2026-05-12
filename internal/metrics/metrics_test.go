package metrics

import (
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
)

func TestStartGoroutinePollingCanBeStopped(t *testing.T) {
	t.Parallel()

	stop := make(chan struct{})
	startGoroutinePolling(time.Millisecond, stop)
	defer close(stop)

	deadline := time.Now().Add(250 * time.Millisecond)
	for time.Now().Before(deadline) {
		metric := &dto.Metric{}
		if err := GoroutineCount.Write(metric); err != nil {
			t.Fatalf("read goroutine gauge: %v", err)
		}
		if metric.GetGauge().GetValue() > 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("goroutine polling did not update the metric")
}
