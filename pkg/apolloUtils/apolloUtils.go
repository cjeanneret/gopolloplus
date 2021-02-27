package apolloUtils
import (
  "context"
  "io/ioutil"
  "log"
  "os"
  "strconv"
  "text/template"
  "github.com/containers/podman/v3/pkg/bindings"
  "github.com/containers/podman/v3/pkg/bindings/play"
  "github.com/containers/podman/v3/pkg/bindings/pods"
  "gopkg.in/ini.v1"
)

func generate_pod(cfg *ApolloConfig, logfile *os.File) (influx_tmp *os.File, pod_tmp *os.File) {
  log.SetOutput(logfile)
  var err error

  influx_tmpl, err := template.ParseFiles("templates/influxdb-config.yaml")
  if err != nil {
    log.Fatalf("ERROR: %v", err)
  }
  pod_tmpl, err := template.ParseFiles("templates/pod.yaml")
  if err != nil {
    log.Fatalf("ERROR: %v", err)
  }

  influx_tmp, err = ioutil.TempFile("", "influx-init.*.yaml")
  if err != nil {
    log.Fatalf("ERROR: %v", err)
  }
  fh, err := os.OpenFile(influx_tmp.Name(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
  if err != nil {
    log.Fatalf("ERROR: %v", err)
  }
  influx_tmpl.Execute(fh, cfg)
  fh.Close()

  pod_tmp, err = ioutil.TempFile("", "gopollo-pod.*.yaml")
  if err != nil {
    log.Fatalf("ERROR: %v", err)
  }
  fh, err = os.OpenFile(pod_tmp.Name(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
  if err != nil {
    log.Fatalf("ERROR: %v", err)
  }
  pod_tmpl.Execute(fh, cfg)
  fh.Close()

  return
}

func ensure_pod(ctx context.Context, nameOrId string, logfile *os.File) {
  log.SetOutput(logfile)
  report, err := pods.Inspect(ctx, nameOrId, nil)
  if err != nil {
    log.Fatalf("ERROR: %v", err)
  }
  restart := false
  for _, ct := range(report.Containers) {
    log.Printf("Checking: %s", ct.Name)
    if ct.State != "running" {
      restart = true
    }
  }
  if restart {
    log.Print("At least one container is down. Restarting Pod!")
    _, err := pods.Restart(ctx, nameOrId, nil)
    if err != nil {
      log.Fatal(err)
    }
  }
}

func ManagePod(cfg *ApolloConfig, logfile *os.File) {
  log.SetOutput(logfile)
  pod_name := cfg.PodName

  log.Print("Starting Podman connection")
  ctx, err := bindings.NewConnection(context.Background(), cfg.PodSocket)
  if err != nil {
    log.Fatal(err)
  }

  // Check if pod already exists
  var exists bool
  log.Print("Checking if pod exists")
  exists, err = pods.Exists(ctx, pod_name, nil)
  if err != nil {
    log.Fatal(err)
  }
  if exists {
    log.Print("Pod exists!")
    ensure_pod(ctx, pod_name, logfile)
    // Exit early
    return
  }

  // Generate pod configuration files
  influx_tmp, pod_tmp := generate_pod(cfg, logfile)
  // Kube options
  kube_opt := new(play.KubeOptions).WithStart(true)
  log.Print("Configuring InfluxDB")
  _, err = play.Kube(ctx, influx_tmp.Name(), kube_opt)
  if err != nil {
    log.Fatal(err)
  }
  log.Print("Clean temporary InfluxDB")
  _, err = pods.Remove(ctx, pod_name, new(pods.RemoveOptions).WithForce(true))
  if err != nil {
    log.Fatal(err)
  }
  log.Print("Create application Pod")
  _, err = play.Kube(ctx, pod_tmp.Name(), kube_opt)
  if err != nil {
    log.Fatal(err)
  }
}

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