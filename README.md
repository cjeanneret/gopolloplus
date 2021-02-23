# GoPolloPlus
Talk to your FDF Apollo Plus console!

## Status
[![Go](https://github.com/cjeanneret/gopolloplus/actions/workflows/go.yml/badge.svg)](https://github.com/cjeanneret/gopolloplus/actions/workflows/go.yml)

## What does it do?
Using the Apollo Plus USB port, this application pushes the data in an InfluxDB service,
allowing to display your stats in Grafana.

## Dependencies (based on Fedora 33)
- device-mapper-devel
- gpgme-devel
- btrfs-progs-devel

Those dependencies are for the Podman integration.

## How does it work?
The main app, ```gopolloplus``` starts a podman pod running two containers:
- InfluxDB
- Grafana

The app then connects to the /dev/ttyUSB0 and reads the data from the monitor, parses and pushes
them in InfluxDB.

### Ports
- InfluxDB is available on :8086
- Grafana is available on :3000

## Configuration
The app is configured using an INI file. An example is located in the "configs/" directory.

### Location
You can either create ~/.gopolloplus.ini, or pass the file running ```gopolloplus -c CONFIG_FILE```.

### Sections
Here are the supported (and mandatory) parameters for each sections.

#### gopolloplus
* ```pod_name```: Name of the pod managed by gopolloplus
* ```manage_pod```: Whether or not gopolloplus must manage the pod
* ```podman_socket```: Location of the socket for podman (starting 2.2, 3 is better)
* ```socket```: FDF Apollo Plus socket - usually /dev/ttyUSB0
* ```log_file```: Full path to the log file

#### grafana
* ```image```: Container image for the grafana container (in case manage_pod is true)
* ```data```: directory path for grafana data. Note that podman currently doesn't support volumes within a play.

#### influxdb
* ```host```: Host URI for the InfluxDB instance
* ```admin_user```: InfluxDB Admin username (for bootstrap)
* ```admin_password```: InfluxDB Admin password (for bootstrap)
* ```user```: InfluxDB standard user (for data insertion)
* ```password```: InfluxDB standard user password (for data insertion)
* ```database```: InfluxDB database name
* ```image```: Container image for InfluxDB
* ```data```: directory path for InfluxDB data
* ```config```: full path to your custom InfluxDB configuration. A sample is available in configs/

## License
This code is provided under the [![cc-by-sa 4.0](https://i.creativecommons.org/l/by-sa/4.0/80x15.png)](https://raw.githubusercontent.com/santisoler/cc-licenses/master/LICENSE-CC-BY-SA) 4.0 license.
