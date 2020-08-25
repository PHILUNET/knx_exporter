# knx_exporter
Exporter for metrics from KNX IP Interface to use with https://prometheus.io

# flags
Name     | Description | Default
---------|-------------|---------
version | Print version information. |
listen-address | Address on which to expose metrics and web interface. | :8080
path | Path under which to expose metrics. | /metrics
gateway-address | IP and port from knx ip interface | 192.168.1.144:3671
devices | File mapping knx addresses to prometheus | devices.yaml

# metrics

ToDo

## Install
```bash
go get -u github.com/philunet/knx_exporter
```

## Usage
```bash
./knx_exporter -gateway-address="192.168.144:3671"
```

## Third Party Components
This software uses components of the following projects
* Prometheus Go client library (https://github.com/prometheus/client_golang)
* knx-go library from Ole Krüger (https://github.com/vapourismo/knx-go)

## License
(c) Philip Berndroth (PHILUNET GmbH), 2020. Licensed under [MIT](LICENSE) license.

## Prometheus
see https://prometheus.io
