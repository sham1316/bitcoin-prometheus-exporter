## Bitcoin Prometheus Exporter

This project provides a simple exporter for a bitcoin node to export metrics in the [Prometheus format](https://prometheus.io/docs/instrumenting/exposition_formats/#text-format-details).

Binary 

```bash
go build -o bitcoin_exporter
```

```bash
BTC_USER=btcuser BTC_PASS=btcpass BTC_URL=127.0.0.1:8332 ./bitcoind_exporter
```

Docker:

```bash
docker build -t bitcoin-prometheus-exporter .
```

```bash
docker run \  
  --name=bitcoin-exporter \
  -p 8334:8334 \
  -e BTC_URL=bitcoin-node \
  -e BTC_USER=alice \
  -e BTC_PASS=DONT_USE_THIS_YOU_WILL_GET_ROBBED_8ak1gI25KFTvjovL3gAM967mies3E= \
  bitcoin-prometheus-exporter
```