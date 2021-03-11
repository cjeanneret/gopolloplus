package main
import (
  "flag"
  "fmt"
  "image/color"
  "log"
  "os"
  "path"
  "time"
  fyne "fyne.io/fyne/v2"
  "fyne.io/fyne/v2/app"
  "fyne.io/fyne/v2/canvas"
  "fyne.io/fyne/v2/container"
  "fyne.io/fyne/v2/theme"
  "fyne.io/fyne/v2/widget"
  "github.com/cjeanneret/gopolloplus/pkg/apolloMonitor"
  "github.com/cjeanneret/gopolloplus/pkg/apolloUtils"
  "github.com/cjeanneret/gopolloplus/pkg/apolloUI"
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

  monitor := apolloMonitor.NewMonitor(cfg.Socket, 9600)

  if !monitor.WaitPort() {
    log.Fatal("Port not available - exiting")
  }

  err = monitor.Connect()
  if err != nil {
    log.Fatal(err)
  }
  monitor.ResetSession()

  data_flow := make(chan *apolloMonitor.ApolloData, 5) // Make a buffered chan just in case
  killWriter := make(chan bool)

  hfile := apolloUtils.GetHistoryFile(cfg)
  history_file, _ := os.OpenFile(hfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
  defer history_file.Close()
  apolloUtils.CSVHeader(history_file)

  if err != nil {
    log.Fatal(err)
  }

  ui := app.New()
  window := ui.NewWindow("")
  window.SetTitle("GoPolloPlus - FDF Apollo Plus Rower Stats")
  window.SetMaster()
  window.CenterOnScreen()
  if cfg.FullScreen {
    window.SetFullScreen(true)
  } else {
    window.Resize(fyne.Size{Width:float32(apolloUI.GraphWidth*3 + 10), Height: 600})
    window.SetFixedSize(true)
  }

  // Define time things (clock and elapsed time)
  clockLabel := apolloUI.TimeCanvas("Time")
  val_elapsed := apolloUI.TimeCanvas("Elapsed")
  val_dist := apolloUI.TimeCanvas("Distance")

  containerTimes := container.NewGridWithColumns(
    3,
    container.NewCenter(clockLabel),
    container.NewCenter(val_elapsed),
    container.NewCenter(val_dist),
  )

  // Define buttons
  button_quit := widget.NewButtonWithIcon("Quit", theme.CancelIcon(), func() {
    killWriter <- true
    monitor.Disconnect()
    history_file.Close()
    window.Close()
    ui.Quit()
  })

  button_reset := widget.NewButtonWithIcon("New Session", theme.DeleteIcon(), func() {
    log.Print("Resetting remote monitor")
    monitor.ResetSession()
    history_file.Close()
    // Prepare a new history file
    hfile = apolloUtils.GetHistoryFile(cfg)
    history_file, _ = os.OpenFile(hfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
  })

  button_c2 := widget.NewButtonWithIcon("Send to log.C2", theme.MailForwardIcon(), func() {
    log.Print("Sending data to log.C2")
  })

  containerButtons := container.NewAdaptiveGrid(
    3,
    button_reset, button_c2, button_quit,
  )

  // Split canvas
  lshift := float32(10)
  split_title := &canvas.Text{Color: color.Gray{Y: 255}, Text: "Split Time",
                              TextSize: apolloUI.TitleFontSize,
                              TextStyle: fyne.TextStyle{Bold: true}}

  split_title.Move(fyne.Position{lshift, 5})

  split_current, split_curr_txt := apolloUI.CreateCanvas(lshift, 55, apolloUI.CurrColor)
  split_avg, split_avg_txt := apolloUI.CreateCanvas(lshift, 110, apolloUI.AVGColor)
  split_max, split_max_txt := apolloUI.CreateCanvas(lshift, 165, apolloUI.MaxColor)

  // Power canvas
  lshift = apolloUI.GraphWidth+10
  power_title := &canvas.Text{Color: color.Gray{Y: 255}, Text: "Power (Watts)",
                              TextSize: apolloUI.TitleFontSize,
                              TextStyle: fyne.TextStyle{Bold: true}}

  power_title.Move(fyne.Position{lshift, 5})
  power_current, power_curr_txt := apolloUI.CreateCanvas(lshift, 55, apolloUI.CurrColor)
  power_avg, power_avg_txt := apolloUI.CreateCanvas(lshift, 110, apolloUI.AVGColor)
  power_max, power_max_txt := apolloUI.CreateCanvas(lshift, 165, apolloUI.MaxColor)

  // SPM canvas
  lshift = (2*apolloUI.GraphWidth)+10
  spm_title := &canvas.Text{Color: color.Gray{Y: 255},
                            Text: "Strokes per minutes",
                            TextSize: apolloUI.TitleFontSize,
                            TextStyle: fyne.TextStyle{Bold: true}}

  spm_title.Move(fyne.Position{lshift, 5})

  spm_current, spm_curr_txt := apolloUI.CreateCanvas(lshift, 55, apolloUI.CurrColor)
  spm_avg, spm_avg_txt := apolloUI.CreateCanvas(lshift, 110, apolloUI.AVGColor)
  spm_max, spm_max_txt := apolloUI.CreateCanvas(lshift, 165, apolloUI.MaxColor)

  // Define graph container
  containerGraphs := container.NewWithoutLayout(split_title, split_current, split_curr_txt,
                                                split_avg, split_avg_txt,
                                                split_max, split_max_txt,
                                                power_title, power_current, power_curr_txt,
                                                power_avg, power_avg_txt,
                                                power_max, power_max_txt,
                                                spm_title, spm_current, spm_curr_txt,
                                                spm_avg, spm_avg_txt,
                                                spm_max, spm_max_txt)

  mainContainer := container.NewVBox(
    containerButtons,
    containerTimes,
    containerGraphs,
  )

  window.SetContent(mainContainer)

  // Update clock
  go func(clockLabel *canvas.Text) {
    var (
      hours, minutes, seconds int
    )
    for {
      hours, minutes, seconds = time.Time.Clock(time.Now())
      clockLabel.Text = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
      clockLabel.Refresh()
      time.Sleep(time.Second)
    }
  }(clockLabel)

  go func() {
    var (
      err error
    )
    for {
      // Let's take the monitor data!
      _, err = monitor.Read(29)
      if err != nil {
        log.Printf("Reader: %v", err)
      }
      if monitor.Data != "" && len(monitor.Data) == 29 {
        log.Printf("Reader: %s", monitor.Data)
        data_flow <- monitor.ParseData()
      }
      time.Sleep(time.Millisecond * 500) // 0.5 second
    }
  }()
  go func() {
    log.Print("Start chan reader")
    var (
      split_history = []uint64{}
      power_history = []uint64{}
      spm_history = []uint64{}
      duration time.Duration
    )
    exit := false
    for {
      select {
      case d := <-data_flow:
        duration, _ = time.ParseDuration(fmt.Sprintf("%vs", d.TotalTime))

        val_dist.Text = fmt.Sprintf("%v meters", d.Distance)
        val_dist.Refresh()
        val_elapsed.Text = fmt.Sprintf("%v", duration)
        val_elapsed.Refresh()

        split_history = append(split_history, d.TimeTo500m)
        power_history = append(power_history, d.Watt)
        spm_history = append(spm_history, d.SPM)

        go apolloUI.ResizeCanvas(split_history, split_current, split_avg, split_max,
                                 split_curr_txt, split_avg_txt, split_max_txt, true)
        go apolloUI.ResizeCanvas(power_history, power_current, power_avg, power_max,
                                 power_curr_txt, power_avg_txt, power_max_txt, false)
        go apolloUI.ResizeCanvas(spm_history, spm_current, spm_avg, spm_max,
                                 spm_curr_txt, spm_avg_txt, spm_max_txt, false)
        go func() {
          history_file.Write([]byte(d.ToCSV()))
        }()
      default:
      }

      select {
      case exit = <-killWriter:
        log.Print("Killing Writer");
      default:
      }
      if exit { break }
    }
  }()

  // show window - LAST action in main()
  window.ShowAndRun()
}

