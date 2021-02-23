package main
import (
  "flag"
  "fmt"
  "log"
  "os"
  "os/signal"
  "strings"
  "syscall"
  "time"
  "github.com/influxdata/influxdb-client-go/v2"
  "go.bug.st/serial"
  "github.com/cjeanneret/gopolloplus/ApolloUtils"
)

func main() {
  config_file := flag.String("c", "", "Configuration file")
  flag.Parse()

  if *config_file == "" {
    log.Fatal("Missing '-c CONFIG_FILE' parameter")
  }

  cfg := ApolloUtils.LoadConfig(*config_file)

  log_file, err := os.OpenFile(cfg.Logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
  if err != nil {
    log.Fatal(err)
  }
  defer log_file.Close()
  log.SetOutput(log_file)

  if cfg.Pod {
    log.Print("Checking and managing Pod")
    ApolloUtils.ManagePod(cfg, log_file)
  }

  for {
    if !ApolloUtils.Check_for_socket(cfg.Socket) {
      log.Print("Port " + cfg.Socket + " does not exist!")
      <-time.After(time.Duration(2 * time.Second))
    } else {
      log.Print("Port " + cfg.Socket + " is present!")
      break
    }
  }

  // Ensure the socket is ready
  time.Sleep(1 * time.Second)

  // Still using 1.x InfluxDB within a local container.
  influxClient := influxdb2.NewClient(cfg.Influx_host,
    fmt.Sprintf("%s:%s", cfg.Influx_user, cfg.Influx_pwd))
  // Be non-blocking. Though we shouldn't have any issue with a 2 second delay....
  writeAPI := influxClient.WriteAPI("", cfg.Influx_db)

  mode := &serial.Mode{
    BaudRate: 9600,
  }
  log.Print("Connecting to " + cfg.Socket)
  port, err := serial.Open(cfg.Socket, mode)
  if err != nil {
    log.Print(err)
    os.Exit(2)
  }

  // Send packet to serial
  port.Write([]byte("C\n"))

  buff := make([]byte, 29)
  output := ""

  // Catch ctrl+c in order to ensure we flush+close InfluxDB client
  c := make(chan os.Signal)
  signal.Notify(c, os.Interrupt, syscall.SIGTERM)
  go func() {
    <-c
    log.Print("Exiting...")
    writeAPI.Flush()
    influxClient.Close()
    port.Close()
    os.Exit(0)
  }()
  // Read serial - infinite loop
  log.Print("Reading from " + cfg.Socket)
  for {
    n, err := port.Read(buff)
    var data *ApolloUtils.ApolloData
    if err != nil {
      log.Printf("ERROR: %v", err)
    }
    content := string(buff[:n])
    if strings.HasPrefix(content, "A8") {
      output = strings.Trim(content, "\r\n")
    } else {
      output += strings.Trim(string(content), "\r\n")
    }
    if len(output) == 29 {
      log.Print("Parse data")
      data = ApolloUtils.Parse_apollo(output)
    }

    if data != nil {
      log.Print("Pushing to InfluxDB")
      p := influxdb2.NewPoint(
            "RowerSession",
            map[string]string{},
            map[string]interface{}{
              "TotalTime": data.TotalTime,
              "Distance": data.Distance,
              "TimeTo500m": data.TimeTo500m,
              "SPM": data.SPM,
              "Watt": data.Watt,
              "CalPerH": data.CalPerH,
              "Level": data.Level,
            }, time.Now())
      writeAPI.WritePoint(p)
    }
  }
}

