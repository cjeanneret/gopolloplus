package main
import (
  "flag"
  "fmt"
  "log"
  "os"
  "path"
  "time"
  "fyne.io/fyne/v2/app"
  "fyne.io/fyne/v2/widget"
  "github.com/influxdata/influxdb-client-go/v2"
  "github.com/cjeanneret/gopolloplus/pkg/apolloUtils"
  "github.com/cjeanneret/gopolloplus/pkg/containerUtils"
  "github.com/cjeanneret/gopolloplus/pkg/usbSocket"
  "go.bug.st/serial"
)

func main() {
  var cfg *apolloUtils.ApolloConfig
  standard_cfg := path.Join(os.Getenv("HOME"), ".gopolloplus.ini")
  _, err := os.Stat(standard_cfg)
  if err == nil {
    log.Printf("Found default config file: %s", standard_cfg)
    cfg = apolloUtils.LoadConfig(standard_cfg)
  } else {
    log.Printf("File not found, checking parameters")
    config_file := flag.String("c", "", "Configuration file")
    flag.Parse()

    if *config_file == "" {
      log.Fatal("Missing '-c CONFIG_FILE' parameter")
    }
    log.Printf("Loading %v", *config_file)
    cfg = apolloUtils.LoadConfig(*config_file)
  }

  log_file, err := os.OpenFile(cfg.Logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
  if err != nil {
    log.Fatal(err)
  }
  defer log_file.Close()
  log.Printf("Writing logs to %s", cfg.Logfile)
  log.SetOutput(log_file)
  log.Print("############ NEW RUN")

  if cfg.Pod {
    log.Print("Checking and managing Pod")
    containerUtils.ManagePod(cfg, log_file)
  }

  // Ensure the socket is ready
  time.Sleep(1 * time.Second)

  // Still using 1.x InfluxDB within a local container.
  influxClient := influxdb2.NewClient(cfg.Influx_host,
    fmt.Sprintf("%s:%s", cfg.Influx_user, cfg.Influx_pwd))
  writeInflux := influxClient.WriteAPI("", cfg.Influx_db)

  mode := &serial.Mode{
    BaudRate: 9600,
  }
  log.Print("Connecting to " + cfg.Socket)
  port, err := serial.Open(cfg.Socket, mode)
  defer port.Close()
  if err != nil {
    log.Fatal(err)
  }

  // Send packet to serial
  port.Write([]byte("C\n"))

  data_flow := make(chan *apolloUtils.ApolloData)
  callback := make(chan bool)

  ui := app.New()
  window := ui.NewWindow("GoPolloPlus")
  button_quit := widget.NewButton("Quit", func() {
    select {
    case callback <- true:
    default:
    }
    close(callback)
    time.Sleep(time.Second)
    port.Close()
    writeInflux.Flush()
    influxClient.Close()
    //window.Destroy()
    ui.Quit()
  })

  window.SetContent(button_quit)

  go usbSocket.ReadSocket(port, log_file, data_flow, callback)
  go func() {
    log.Print("Start chan reader")
    for {
      d := <-data_flow
      log.Printf("%v", d)
      p := influxdb2.NewPoint(
        "RowerSession",
        map[string]string{},
        map[string]interface{}{
          "TotalTime": d.TotalTime,
          "Distance": d.Distance,
          "TimeTo500m": d.TimeTo500m,
          "SPM": d.SPM,
          "Watt": d.Watt,
          "CalPerH": d.CalPerH,
          "Level": d.Level,
        }, time.Now())
        writeInflux.WritePoint(p)
      }
  }()

  window.ShowAndRun()
}

