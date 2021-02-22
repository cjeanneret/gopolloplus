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

The app then connect to the /dev/ttyUSB0 and reads the data from the monitor, parses and pushes
them in InfluxDB.

### Ports
- InfluxDB is available on :8086
- Grafana is available on :3000

## License
This code is provided under the [![cc-by-sa 4.0](https://i.creativecommons.org/l/by-sa/4.0/80x15.png)](https://raw.githubusercontent.com/santisoler/cc-licenses/master/LICENSE-CC-BY-SA) license.
