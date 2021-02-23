package main

import (
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"sham1316/bitcoin-prometheus-exporter/config"
	"syscall"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setGauge(name string, prefix string, subsystem string, help string, callback func() float64) {
	gaugeFunc := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: prefix,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
	}, callback)
	prometheus.MustRegister(gaugeFunc)
}

func main() {
	cfg := config.GetInstance()

	client, err := rpcclient.New(cfg.RpcConfig, nil)
	if err != nil {
		panic(err)
	}
	defer client.Shutdown()
	warningCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: cfg.Metrics.MetricsPrefix,
		Name:      "warnings",
		Help:      "Number of network or blockchain warnings detected",
	})

	setGauge("blocks", cfg.Metrics.MetricsPrefix, "", "Block height", func() float64 {
		blockCount, err := client.GetBlockCount()
		if err != nil {
			warningCounter.Inc()
			zap.S().Panic(err)
		}
		return float64(blockCount)
	})
	setGauge("peers", cfg.Metrics.MetricsPrefix, "", "The number of connected peers", func() float64 {
		peerInfo, err := client.GetPeerInfo()
		if err != nil {
			warningCounter.Inc()
			zap.S().Error(err)
		}
		return float64(len(peerInfo))
	})
	setGauge("difficulty", cfg.Metrics.MetricsPrefix, "", "Difficulty", func() float64 {
		difficulty, err := client.GetDifficulty()
		if err != nil {
			warningCounter.Inc()
			zap.S().Error(err)
		}
		return difficulty
	})

	setGauge("hashps_neg1", cfg.Metrics.MetricsPrefix, "", "Estimated network hash rate per second since the last difficulty change", func() float64 {
		value, err := client.GetNetworkHashPS2(-1)
		if err != nil {
			warningCounter.Inc()
			zap.S().Error(err)
		}
		return float64(value)
	})

	setGauge("hashps_1", cfg.Metrics.MetricsPrefix, "", "Estimated network hash rate per second for the last block", func() float64 {
		value, err := client.GetNetworkHashPS2(1)
		if err != nil {
			warningCounter.Inc()
			zap.S().Error(err)
		}
		return float64(value)
	})
	setGauge("hashps", cfg.Metrics.MetricsPrefix, "", "Estimated network hash rate per second for the last 120 blocks", func() float64 {
		value, err := client.GetNetworkHashPS2(120)
		if err != nil {
			warningCounter.Inc()
			zap.S().Error(err)
		}
		return float64(value)
	})

	setGauge("size", cfg.Metrics.MetricsPrefix, "mempool", "The number of txes in rawmempool", func() float64 {
		value, err := client.GetRawMempool()
		if err != nil {
			warningCounter.Inc()
			zap.S().Error(err)
		}
		return float64(len(value))
	})

	http.Handle(cfg.Metrics.MetricsPath, WithLogging(promhttp.Handler()))
	zap.S().Info("Start listener on " + cfg.Metrics.MetricsHost)
	zap.S().Fatal(http.ListenAndServe(cfg.Metrics.MetricsHost, nil))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	sigName := <-sig
	zap.S().Errorf("RECEIVE SIGNAL - %s", sigName)
}

func WithLogging(h http.Handler) http.Handler {
	loggingFn := func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now()

		h.ServeHTTP(rw, req) // inject our implementation of http.ResponseWriter

		duration := time.Since(start)

		zap.S().Infof("uri %s, duration %d ms",
			req.RequestURI, duration.Milliseconds())
	}
	return http.HandlerFunc(loggingFn)
}
