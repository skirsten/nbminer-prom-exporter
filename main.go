package main

import (
	"encoding/json"
	"flag"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type NBMinerCollector struct {
	Client   *http.Client
	Endpoint string
}

// Interface guards
var (
	_ prometheus.Collector = (*NBMinerCollector)(nil)
)

var (
	labels = []string{"device_id", "device_pci_bus_id", "device_name"}

	// deviceCoreClock = prometheus.NewDesc(
	// 	"miner_device_core_clock",
	// 	"Device core clock in MHz.",
	// 	labels, nil,
	// )
	deviceHashRate = prometheus.NewDesc(
		"miner_device_hash_rate",
		"Hash rate as reported by NBMiner.",
		labels, nil,
	)
)

func (c *NBMinerCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

type nbminerStatusResponse struct {
	Miner struct {
		Devices []struct {
			Id       int    `json:"id"`
			PciBusId int    `json:"pci_bus_id"`
			Name     string `json:"info"`

			Hashrate  float64 `json:"hashrate_raw"`
			Hashrate2 float64 `json:"hashrate2_raw"`

			Temperature     int `json:"temperature"`
			MemTemperature  int `json:"memTemperature"`
			MemClock        int `json:"mem_clock"`
			CoreClock       int `json:"core_clock"`
			MemUtilization  int `json:"mem_utilization"`
			CoreUtilization int `json:"core_utilization"`

			Power int `json:"power"`
			Fan   int `json:"fan"`

			AcceptedShares int `json:"accepted_shares"`
			InvalidShares  int `json:"invalid_shares"`
			RejectedShares int `json:"rejected_shares"`
		} `json:"devices"`
	} `json:"miner"`
}

func (c *NBMinerCollector) Collect(ch chan<- prometheus.Metric) {
	resp, err := c.Client.Get(c.Endpoint)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	status := nbminerStatusResponse{}

	err = decoder.Decode(&status)
	if err != nil {
		log.Println(err)
		return
	}

	for _, device := range status.Miner.Devices {
		ch <- prometheus.MustNewConstMetric(
			deviceHashRate,
			prometheus.CounterValue,
			math.Round((device.Hashrate/(1000*1000))*1000)/1000,
			// "device_id", "device_pci_bus_id", "device_name"
			strconv.Itoa(device.Id), strconv.Itoa(device.PciBusId), device.Name,
		)
	}
}

var (
	addr           = flag.String("listen-address", ":2121", "The address to listen on for HTTP requests.")
	statusEndpoint = flag.String("nbminer-status-endpoint", "http://localhost:8080/api/v1/status", "The address to the NBMiner status endpoint.")
)

func main() {
	flag.Parse()

	reg := prometheus.NewPedanticRegistry()

	collector := &NBMinerCollector{
		Client: &http.Client{
			Timeout: time.Second * 30,
		},
		Endpoint: *statusEndpoint,
	}

	reg.MustRegister(collector)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
