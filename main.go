package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
	"github.com/vapourismo/knx-go/knx/dpt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Devices   []*DeviceConfig `yaml:"devices,omitempty"`
}

type DeviceConfig struct {
	Name 		string	`yaml:"name"`
	MainGroup 	uint8	`yaml:"maingroup"`
	MiddleGroup 	uint8   `yaml:"middlegroup"`
	SubGroup 	uint8   `yaml:"subgroup"`
	Type  		string  `yaml:"type"`
}

var (
	showVersion   = flag.Bool("version", false, "Print version information.")
	listenAddress = flag.String("listen-address", ":8080", "Address on which to expose metrics.")
	metricsPath   = flag.String("path", "/metrics", "Path under which to expose metrics.")
	gatewayAddr   = flag.String("gateway-address", "192.168.1.144:3671", "IP and port from knx ip interface.")
	deviceConfig  = flag.String("devices", "devices.yaml", "File mapping knx addresses to prometheus")

	tempMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prefix + "temperature_celsius",
		Help: "Current temperature from KNX DPT_9001 datapoint",
	}, []string{prefix + "addr", "name", prefix + "maingroup", prefix + "middlegroup", prefix + "subgroup"})

	config Config
	client knx.GroupTunnel
	err error
)

const version string = "0.1"
const prefix = "knx_"

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: KNX exporter [ ... ]\n\nParameters:")
		fmt.Println()
		flag.PrintDefaults()
	}
	readConfig()
	flag.Parse()
}

func main() {
	log.Printf("Starting KNX exporter (Version: %s)", version)

	// Connect to the gateway.
	client, err = knx.NewGroupTunnel(*gatewayAddr, knx.DefaultTunnelConfig)

	// check gateway connection status
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Successfully connected to KXN IP interface: %s", *gatewayAddr)
	}

	// Close upon exiting. Even if the gateway closes the connection, we still have to clean up.
	defer client.Close()

	go startServer()
	go updateMetrics()

	// Receive messages from the gateway. The inbound channel is closed with the connection.
	for msg := range client.Inbound() {

		for _, dev := range config.Devices {
			devAdr := cemi.NewGroupAddr3(dev.MainGroup, dev.MiddleGroup, dev.SubGroup)
			if msg.Destination == devAdr {
				go updatePrometheus(msg, *dev)
				break
			}
		}

		if err != nil {
			continue
		}
	}
}

func startServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>KNX Exporter</title></head>
			<body>
			<h1>KNX Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			<h2>More information:</h2>
			<p><a href="https://github.com/philunet/knx_exporter">github.com/philunet/knx_exporter</a></p>
			</body>
			</html>`))
	})

	// Expose the registered metrics via HTTP.
	http.Handle(*metricsPath, promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			//EnableOpenMetrics: true,
		},
	))

	prometheus.MustRegister(tempMetric)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func updatePrometheus(msg knx.GroupEvent, dev DeviceConfig){

	//knx data type is status (bool)
	if dev.Type == "DPT_1001" {

	}

	//knx data type is temp (float32)
	if dev.Type == "DPT_9001" {
		var temp dpt.DPT_9001
		err = temp.Unpack(msg.Data)
		strAdr := strconv.Itoa(int(dev.MainGroup)) + "_" + strconv.Itoa(int(dev.MiddleGroup)) + "_" + strconv.Itoa(int(dev.SubGroup))

		tempMetric.With(prometheus.Labels{
			"knx_addr": strAdr,
			"name": dev.Name,
			"knx_maingroup": strconv.Itoa(int(dev.MainGroup)),
			"knx_middlegroup": strconv.Itoa(int(dev.MiddleGroup)),
			"knx_subgroup": strconv.Itoa(int(dev.SubGroup)),
		}).Set(float64(temp))
	}
}

//run after start to update all metrics with current values from knx bus
func updateMetrics() {
	for _, dev := range config.Devices {
		knxAdr := cemi.NewGroupAddr3(dev.MainGroup, dev.MiddleGroup, dev.SubGroup)

		err := client.Send(knx.GroupEvent{
			Command:	 knx.GroupRead,
			Destination: knxAdr,
		})

		if err != nil {
			log.Fatal(err)
		}
	}
}

func readConfig() {
	filename, _ := filepath.Abs(*deviceConfig)

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatal(err)
	}
}
