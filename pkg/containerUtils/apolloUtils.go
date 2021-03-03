package containerUtils
import (
  "context"
  "io/ioutil"
  "log"
  "os"
  "text/template"
  "github.com/cjeanneret/gopolloplus/pkg/apolloUtils"
  "github.com/containers/podman/v3/pkg/bindings"
  "github.com/containers/podman/v3/pkg/bindings/play"
  "github.com/containers/podman/v3/pkg/bindings/pods"
)

func generate_pod(cfg *apolloUtils.ApolloConfig, logfile *os.File) (influx_tmp *os.File, pod_tmp *os.File) {
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

func ManagePod(cfg *apolloUtils.ApolloConfig, logfile *os.File) {
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
