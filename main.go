package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"sham1316/bitcoin-prometheus-exporter/config"
	"syscall"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/prometheus/client_golang/prometheus"
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
	setGauge("difficulty", cfg.Metrics.MetricsPrefix, "", "The proof-of-work difficulty as a multiple of the minimum difficulty", func() float64 {
		value, err := client.GetDifficulty()
		if err != nil {
			warningCounter.Inc()
			zap.S().Error(err)
		}
		return value
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

	mempoolBytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: cfg.Metrics.MetricsPrefix,
		Subsystem: "mempool",
		Name:      "bytes",
		Help:      "Bytes in mempool",
	})
	prometheus.MustRegister(mempoolBytes)
	mempoolBytes.Set(float64(42))

	setGauge("size", cfg.Metrics.MetricsPrefix, "mempool", "The number of txes in mempool", func() float64 {
		mempoolInfo, err := client.GetMempoolInfo()
		if err != nil {
			warningCounter.Inc()
			zap.S().Error(err)
		}
		mempoolBytes.Set(float64(mempoolInfo.Bytes))
		return float64(mempoolInfo.Size)
	})

	setGauge("size_on_disk", cfg.Metrics.MetricsPrefix, "", "Estimated size of the block and undo files", func() float64 {
		blockChainInfo, err := client.GetBlockChainInfo()
		if err != nil {
			warningCounter.Inc()
			zap.S().Error(err)
		}
		return float64(blockChainInfo.SizeOnDisk)
	})

	setGauge("uptime", cfg.Metrics.MetricsPrefix, "", "Number of seconds the Bitcoin daemon has been running", func() float64 {
		value, err := client.GetUptime()
		if err != nil {
			warningCounter.Inc()
			zap.S().Error(err)
		}
		return float64(value)
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

		h.ServeHTTP(rw, req)

		duration := time.Since(start)

		zap.S().Infof("uri %s, duration %d ms",
			req.RequestURI, duration.Milliseconds())
	}
	return http.HandlerFunc(loggingFn)
}
