package apolloUtils

import (
  "log"
  "os"
  "strconv"
  "fyne.io/fyne/v2"
  "gopkg.in/ini.v1"
  "fyne.io/fyne/v2/theme"
  homedir "github.com/mitchellh/go-homedir"
)

type ApolloConfig struct {
  FullScreen bool
  ConfigFile, ThemeVariant, Socket, LogFile, HistoryDir string
  Theme fyne.Theme
}

func (a *ApolloConfig) Write() {
  // TODO: dump config to file (override)
  // Note: Write must override the existing config file
  cfg_file, err := os.OpenFile(a.ConfigFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
  if err != nil {
    log.Fatal(err)
  }
  defer cfg_file.Close()

  cfg := ini.Empty()
  cfg.Section("gopolloplus").NewKey("socket", a.Socket)
  cfg.Section("gopolloplus").NewKey("log_file", a.LogFile)
  cfg.Section("gopolloplus").NewKey("fullscreen", strconv.FormatBool(a.FullScreen))
  cfg.Section("gopolloplus").NewKey("history_dir", a.HistoryDir)
  cfg.Section("gopolloplus").NewKey("theme", a.ThemeVariant)

  cfg.WriteTo(cfg_file)
}

func DefaultConfig() *ApolloConfig {
  standard_cfg, _ := homedir.Expand("~/.gopolloplus.ini")
  config := &ApolloConfig{
    Socket: "/dev/ttyUSB0",
    LogFile: "/tmp/gopolloplus.log",
    FullScreen: false,
    HistoryDir: "/tmp/gopollo-history",
    ConfigFile: standard_cfg,
    ThemeVariant: "dark",
    Theme: theme.DarkTheme(),
  }

  return config
}

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
  if err != nil || history == "" {
    history = "/tmp/gopollo-history"
  }

  logfile, err := homedir.Expand(cfg.Section("gopolloplus").Key("log_file").String())
  if err != nil || logfile == "" {
    logfile = "/tmp/gopolloplus.log"
  }

  socket := cfg.Section("gopolloplus").Key("socket").String()
  if socket == "" {
    socket = "/dev/ttyUSB0"
  }

  t := cfg.Section("gopolloplus").Key("theme").String()
  th := theme.DarkTheme()
  variant := "dark"
  if t == "light" {
    th = theme.LightTheme()
    variant = "light"
  }

  config := &ApolloConfig{
    Socket: socket,
    LogFile: logfile,
    FullScreen: fullscreen,
    HistoryDir: history,
    Theme: th,
    ThemeVariant: variant,
    ConfigFile: config_file,
  }

  return config
}
