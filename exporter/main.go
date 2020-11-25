package main

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"sync"
)

type EnvSensor struct {
	T1 float64
	P1 float64
	H1 float64
	T2 float64
	P2 float64
	H2 float64
	L  float64
}

type Collector struct {
	describes []*prometheus.Desc
	metrics   []prometheus.Metric
	sync.RWMutex
}

func NewCollector() *Collector {
	return &Collector{
		describes: []*prometheus.Desc{
			prometheus.NewDesc("envsensor_temperature", "Temperature (celsius)", []string{"sensor"}, nil),
			prometheus.NewDesc("envsensor_humidity", "Humidity (%)", []string{"sensor"}, nil),
			prometheus.NewDesc("envsensor_air_pressure", "Air Pressure (hPa)", []string{"sensor"}, nil),
			prometheus.NewDesc("envsensor_luminance", "Luminance (lux)", []string{"sensor"}, nil),
		},
		metrics: []prometheus.Metric{},
	}
}

func (c *Collector) Describe(descs chan<- *prometheus.Desc) {
	for _, d := range c.describes {
		descs <- d
	}
}

func (c *Collector) Collect(metrics chan<- prometheus.Metric) {
	c.RLock()
	defer c.RUnlock()
	for _, m := range c.metrics {
		metrics <- m
	}
}

func main() {
	nc, err := nats.Connect(os.Getenv("NATS_SERVER"))
	if err != nil {
		log.Fatal(err)
	}

	c := NewCollector()
	prometheus.MustRegister(c)

	_, err = nc.Subscribe("work.wtks.home.envsensor", func(msg *nats.Msg) {
		var values EnvSensor
		if err := json.Unmarshal(msg.Data, &values); err != nil {
			log.Println(err)
			c.Lock()
			c.metrics = []prometheus.Metric{}
			c.Unlock()
			return
		}

		metrics := []prometheus.Metric{
			prometheus.MustNewConstMetric(c.describes[0], prometheus.GaugeValue, values.T1, "bme280_ch1"),
			prometheus.MustNewConstMetric(c.describes[1], prometheus.GaugeValue, values.H1, "bme280_ch1"),
			prometheus.MustNewConstMetric(c.describes[2], prometheus.GaugeValue, values.P1, "bme280_ch1"),
			prometheus.MustNewConstMetric(c.describes[0], prometheus.GaugeValue, values.T2, "bme280_ch2"),
			prometheus.MustNewConstMetric(c.describes[1], prometheus.GaugeValue, values.H2, "bme280_ch2"),
			prometheus.MustNewConstMetric(c.describes[2], prometheus.GaugeValue, values.P2, "bme280_ch2"),
			prometheus.MustNewConstMetric(c.describes[3], prometheus.GaugeValue, values.L, "tsl25721"),
		}

		c.Lock()
		c.metrics = metrics
		c.Unlock()
	})
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
