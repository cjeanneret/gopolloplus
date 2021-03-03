package apolloUtils
import (
  "os"
  "strconv"
  "gopkg.in/ini.v1"
)

func Check_for_socket(socket string) bool {
  _, err := os.Stat(socket)
  return !os.IsNotExist(err)
}

func Parse_apollo(data string) *ApolloData {
  totalMinutes, _ := strconv.ParseInt(data[3:5], 10, 64)
  totalSeconds, _ := strconv.ParseInt(data[5:7], 10, 64)
  totalTime := totalMinutes*60+totalSeconds
  distance, _ := strconv.ParseInt(data[7:12], 10, 64)
  MinutesTo500m, _ := strconv.ParseInt(data[13:15], 10, 64)
  SecondsTo500m, _ := strconv.ParseInt(data[15:17], 10, 64)
  timeTo500m := MinutesTo500m*60+SecondsTo500m
  spm, _ := strconv.ParseInt(data[17:20], 10, 64)
  watt, _ := strconv.ParseInt(data[20:23], 10, 64)
  calph, _ := strconv.ParseInt(data[23:27], 10, 64)
  level, _ := strconv.ParseInt(data[27:29], 10, 64)

  output := ApolloData{
    TotalTime: totalTime,
    Distance: distance,
    TimeTo500m: timeTo500m,
    SPM: spm,
    Watt: watt,
    CalPerH: calph,
    Level: level,
  }

  return &output
}

func LoadConfig(config_file string) *ApolloConfig {
  cfg, err := ini.Load(config_file)
  if err != nil {
    return nil
  }
  manage_pod, err := cfg.Section("gopolloplus").Key("manage_pod").Bool()
  if err != nil {
    manage_pod = false
  }

  config := ApolloConfig{
    Pod: manage_pod,
    PodName: cfg.Section("gopolloplus").Key("pod_name").String(),
    PodSocket: cfg.Section("gopolloplus").Key("podman_socket").String(),
    Socket: cfg.Section("gopolloplus").Key("socket").String(),
    Logfile: cfg.Section("gopolloplus").Key("log_file").String(),
    Grafana_img: cfg.Section("grafana").Key("image").String(),
    Grafana_data: cfg.Section("grafana").Key("data").String(),
    Influx_host: cfg.Section("influxdb").Key("host").String(),
    Influx_adm: cfg.Section("influxdb").Key("admin_user").String(),
    Influx_adm_pwd: cfg.Section("influxdb").Key("admin_password").String(),
    Influx_user: cfg.Section("influxdb").Key("user").String(),
    Influx_pwd: cfg.Section("influxdb").Key("password").String(),
    Influx_db: cfg.Section("influxdb").Key("database").String(),
    Influx_img: cfg.Section("influxdb").Key("image").String(),
    Influx_data: cfg.Section("influxdb").Key("data").String(),
    Influx_conf: cfg.Section("influxdb").Key("config").String(),
  }

  return &config
}
