package apolloMonitor

import (
  "fmt"
  "encoding/json"
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
  TotalTime, Distance, TimeTo500m uint64
  SPM, Watt, CalPerH, Level uint64
  Timestamp, Raw string
}

func (d *ApolloData) ToJSON() ([]byte, error) {
  return json.Marshal(d)
}

func (d *ApolloData) ToCSV() (string) {
  return fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v\n",
                     d.Timestamp, d.TotalTime, d.Distance, d.TimeTo500m,
                     d.SPM, d.Watt, d.CalPerH, d.Level, d.Raw)
}

func (m Monitor) ParseData() *ApolloData {
  totalMinutes, _ := strconv.ParseUint(m.Data[3:5], 10, 64)
  totalSeconds, _ := strconv.ParseUint(m.Data[5:7], 10, 64)
  distance, _ := strconv.ParseUint(m.Data[7:12], 10, 64)
  MinutesTo500m, _ := strconv.ParseUint(m.Data[13:15], 10, 64)
  SecondsTo500m, _ := strconv.ParseUint(m.Data[15:17], 10, 64)
  spm, _ := strconv.ParseUint(m.Data[17:20], 10, 64)
  watt, _ := strconv.ParseUint(m.Data[20:23], 10, 64)
  calph, _ := strconv.ParseUint(m.Data[23:27], 10, 64)
  level, _ := strconv.ParseUint(m.Data[27:29], 10, 64)

  totalTime := totalMinutes*60+totalSeconds
  timeTo500m := MinutesTo500m*60+SecondsTo500m

  t := time.Now()

  output := &ApolloData{
    Timestamp: t.Format("20060102150405"),
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
