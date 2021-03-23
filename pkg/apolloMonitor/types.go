package apolloMonitor

import (
  "encoding/csv"
  "encoding/json"
  "fmt"
  "log"
  "os"
  "strconv"
  "strings"
  "time"
  "go.bug.st/serial"
  retry "github.com/avast/retry-go"
)

type Monitor struct {
  Port string // port address
  SerialPort serial.Port // open port
  BaudRate int
  IsAlive bool
  IsResetSuccess bool
  Level uint64
  HRate int64
  Data string
}

func NewMonitor(p string, b int) *Monitor{
  mon := &Monitor{Port: p, BaudRate: b,}
  return mon
}

func (m *Monitor) WaitPort() bool {
  err := retry.Do(
    func() error {
      _, err := os.Stat(m.Port)
      if err != nil {
        log.Printf("RETRY: port %v", m.Port)
        return err
      }
      return nil
    },
    retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
      return retry.BackOffDelay(n, err, config)
    }),
    retry.Attempts(10),
  )
  return (err == nil)
}

func (m *Monitor) Connect() (err error) {
  mode := &serial.Mode{BaudRate: m.BaudRate}
  port, err := serial.Open(m.Port, mode)
  m.SerialPort = port
  return
}

func (m *Monitor) SetMode(baud int) {
  m.BaudRate = baud
}

func (m *Monitor) Write(b []byte) {
  _, err := m.SerialPort.Write(b)
  if err != nil {
    log.Printf("Write: %v", err)
  }
}

func (m *Monitor) Read(size int) (msg string, err error) {
  // Ensure HRate is null before reading
  m.HRate = 0

  buff := make([]byte, size)
  n, err := m.SerialPort.Read(buff)
  msg = string(buff[:n])
  switch {
  case strings.HasPrefix(msg, "A"):
    m.Data = strings.Trim(msg, "\r\n")
  case strings.HasPrefix(msg, "H"):
    m.HRate, _ = strconv.ParseInt(string(msg[1:4]), 10, 64)
    m.Data = ""
  case strings.HasPrefix(msg, "L"):
    m.Level, _ = strconv.ParseUint(string(msg[1]), 10, 64)
    m.Data = ""
  case strings.HasPrefix(msg, "R"):
    m.IsResetSuccess = true;
    m.Data = ""
  case strings.HasPrefix(msg, "K"):
    m.IsAlive = true;
    m.Data = ""
  case strings.HasPrefix(msg, "W"):
    m.KeepAlive();
    m.Data = ""
  case strings.HasPrefix(msg, "C"):
    // Console ACK clear
    m.Data = ""
  default:
    m.Data = m.Data + strings.Trim(msg, "\r\n")
  }
  return
}

func (m *Monitor) ResetSession() {
  m.Write([]byte("C\n"))
}

func (m *Monitor) GetType() (string) {
  m.Write([]byte("T\n"))
  msg, err := m.Read(3)
  if err != nil {
    return ""
  }
  return msg
}

func (m *Monitor) GetVersion() (v string) {
  m.Write([]byte("\n"))
  msg, err := m.Read(10)
  if err != nil {
    return ""
  }
  return msg
}

func (m *Monitor) SetLevel(level int) {
  m.Write([]byte(fmt.Sprintf("L%v\n", level)))
}

func (m *Monitor) GetHeartRate() (hc string) {
  m.Write([]byte("H\n"))
  msg, err := m.Read(4)
  if err != nil {
    return ""
  }
  return msg
}

func (m *Monitor) KeepAlive() {
  m.Write([]byte("K\n"))
}

func (m *Monitor) Disconnect() {
  m.Write([]byte("D\n"))
  m.SerialPort.Close()
}

type ApolloData struct {
  TotalTime, Distance, TimeTo500m int64
  SPM, Watt, CalPerH, Level int64
  Timestamp int64
  Raw string
}

func (d *ApolloData) ToJSON() ([]byte, error) {
  return json.Marshal(d)
}

func (d *ApolloData) ToCSV() (string) {
  return fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v\n",
                     d.Timestamp, d.TotalTime, d.Distance, d.TimeTo500m,
                     d.SPM, d.Watt, d.CalPerH, d.Level, d.Raw)
}

func LoadCSV(filename string) (data []*ApolloData) {
  f, err := os.Open(filename)
  if err != nil {
    log.Fatalf("LoadCSV: %v", err)
  }
  defer f.Close()
  lines, err := csv.NewReader(f).ReadAll()
  if err != nil {
    log.Fatalf("LoadCSV: %v", err)
  }

  for _, line := range lines {
    timestamp, _ := strconv.ParseInt(line[0], 10, 64)
    totalTime, _ := strconv.ParseInt(line[1], 10, 64)
    distance, _ := strconv.ParseInt(line[2], 10, 64)
    split, _ := strconv.ParseInt(line[3], 10, 64)
    spm, _ := strconv.ParseInt(line[4], 10, 64)
    watt, _ := strconv.ParseInt(line[5], 10, 64)
    calph, _ := strconv.ParseInt(line[6], 10, 64)
    lvl, _ := strconv.ParseInt(line[7], 10, 64)

    data = append(data, &ApolloData{
      Timestamp: timestamp,
      TotalTime: totalTime,
      Distance: distance,
      TimeTo500m: split,
      SPM: spm,
      Watt: watt,
      CalPerH: calph,
      Level: lvl,
      Raw: line[8],
    })
  }
  return
}

func (m Monitor) ParseData() *ApolloData {
  t := time.Now()

  timestamp, _ := strconv.ParseInt(t.Format("20060102150405"), 10, 64)
  totalMinutes, _ := strconv.ParseInt(m.Data[3:5], 10, 64)
  totalSeconds, _ := strconv.ParseInt(m.Data[5:7], 10, 64)
  distance, _ := strconv.ParseInt(m.Data[7:12], 10, 64)
  MinutesTo500m, _ := strconv.ParseInt(m.Data[13:15], 10, 64)
  SecondsTo500m, _ := strconv.ParseInt(m.Data[15:17], 10, 64)
  spm, _ := strconv.ParseInt(m.Data[17:20], 10, 64)
  watt, _ := strconv.ParseInt(m.Data[20:23], 10, 64)
  calph, _ := strconv.ParseInt(m.Data[23:27], 10, 64)
  level, _ := strconv.ParseInt(m.Data[27:29], 10, 64)

  totalTime := totalMinutes*60+totalSeconds
  timeTo500m := MinutesTo500m*60+SecondsTo500m

  output := &ApolloData{
    Timestamp: timestamp,
    TotalTime: totalTime,
    Distance: distance,
    TimeTo500m: timeTo500m,
    SPM: spm,
    Watt: watt,
    CalPerH: calph,
    Level: level,
    Raw: m.Data,
  }

  return output
}
