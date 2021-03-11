module sham1316/bitcoin-prometheus-exporter

go 1.15

require (
	github.com/btcsuite/btcd v0.21.0-beta.0.20210305124519-d08785547a87
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/prometheus/client_golang v1.9.0
	github.com/sham1316/configparser v0.0.0-20200623154026-c5b8f6832218
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/btcsuite/btcd => github.com/sham1316/btcd v0.21.0-beta.0.20210311150252-7788eda524af
