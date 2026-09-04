//go:build ignore

package main

import (
	"context"
	"net/http"
	"time"

	"github.com/KoJaco/leakwatch"
	promexport "github.com/KoJaco/leakwatch/export/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	watcher := leakwatch.New(leakwatch.WithInterval(time.Minute))
	watcher.Start(context.Background())

	reg := prometheus.NewRegistry()
	reg.MustRegister(promexport.NewCollector(watcher))

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	_ = http.ListenAndServe(":9090", nil)
}
