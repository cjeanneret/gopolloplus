package apolloUtils

type ApolloData struct {
  TotalTime int64
  Distance int64
  TimeTo500m int64
  SPM int64
  Watt int64
  CalPerH int64
  Level int64
}

type ApolloConfig struct {
  Pod bool
  PodName, PodSocket, Socket, Logfile string
  Grafana_img, Grafana_data string
  Influx_host, Influx_adm, Influx_adm_pwd string
  Influx_user, Influx_pwd string
  Influx_db, Influx_img, Influx_data, Influx_conf string
}
