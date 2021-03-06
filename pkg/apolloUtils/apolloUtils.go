package apolloUtils
import (
  "log"
  "os"
  "path"
  "time"
  "gopkg.in/ini.v1"
  homedir "github.com/mitchellh/go-homedir"
)

func LoadConfig(config_file string) *ApolloConfig {
  cfg, err := ini.Load(config_file)
  if err != nil {
    return nil
  }
  fullscreen, err := cfg.Section("gopolloplus").Key("fullscreen").Bool()
  if err != nil {
    fullscreen = false
  }

  history, err := homedir.Expand(cfg.Section("gopolloplus").Key("history_dir").String())
  if err != nil {
    history = "/tmp"
  }

  config := ApolloConfig{
    Socket: cfg.Section("gopolloplus").Key("socket").String(),
    Logfile: cfg.Section("gopolloplus").Key("log_file").String(),
    FullScreen: fullscreen,
    HistoryDir: history,
  }

  return &config
}

func FindMinMax(a []uint64) (min, max uint64) {
  min = a[0]
  max = a[0]
  for _, value := range a {
    if value < min {
      min = value
    }
    if value > max {
      max = value
    }
  }
  return
}

func Average(a []uint64) (avg float64) {
  total := 0.0
  for _, value := range a{
    total += float64(value)
  }
  avg = total / float64(len(a))
  return
}

func GetHistoryFile(cfg *ApolloConfig) (p string) {
  log.Printf("Checking for %s", cfg.HistoryDir)
  _, err := os.Stat(cfg.HistoryDir)
  if os.IsNotExist(err) {
    err := os.Mkdir(cfg.HistoryDir, 0755)
    if err != nil {
      log.Fatal(err)
    }
  }

  t := time.Now()
  p = path.Join(cfg.HistoryDir, t.Format("2006-01-02-150405"))
  return
}

func CSVHeader(f *os.File) {
  header := []byte("timestamp,totalTime,distance,timeTo500m,SPM,watt,calPerH,level\n")
  f.Write(header)
}
