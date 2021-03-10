# GoPolloPlus
Talk to your FDF Apollo Plus console!

## Status
[![Go](https://github.com/cjeanneret/gopolloplus/actions/workflows/go.yml/badge.svg)](https://github.com/cjeanneret/gopolloplus/actions/workflows/go.yml)

## What does it do?
This application listens to the FDF USB Serial port, logs the data in a CSV, and shows simple graphs.

## Build dependencies (based on Ubuntu)
- libgl1-mesa-dev
- xorg-dev

Those dependencies are for Fyne support

## How does it work?
The app connects to the /dev/ttyUSB0 and reads the data from the monitor, parses and pushes
them in a CSV file (one per session), and shows simple graphs using Fyne Canvas.

See the Configuration section for more information.

## Configuration
The app is configured using an INI file. An example is located in the "configs/" directory.

### Location
You can either create ~/.gopolloplus.ini, or pass the file running ```gopolloplus -c CONFIG_FILE```.

### Sections
Here are the supported (and mandatory) parameters for each sections.

#### gopolloplus
* ```socket```: FDF Apollo Plus socket - usually /dev/ttyUSB0
* ```log_file```: Full path to the log file
* ```fullscreen```: Whether the window must be in fullscreen or not. Boolean. Defaults to false.
* ```history_dir```: Location for the history files. Better if it exists. Supports "~/" in the path.


## License
This code is provided under the [![cc-by-sa 4.0](https://i.creativecommons.org/l/by-sa/4.0/80x15.png)](https://raw.githubusercontent.com/santisoler/cc-licenses/master/LICENSE-CC-BY-SA) 4.0 license.
