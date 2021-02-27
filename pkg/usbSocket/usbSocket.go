package usbSocket

import (
  "log"
  "os"
  "strings"
  "time"
  retry "github.com/avast/retry-go"
  "go.bug.st/serial"
  "github.com/cjeanneret/gopolloplus/pkg/apolloUtils"
)

func WaitForSocket(socket string, attempts uint, logfile *os.File) bool {
  log.SetOutput(logfile)
  err := retry.Do(
    func() error {
      _, err := os.Stat(socket)
      if err != nil {
        log.Print("Socket: retrying...")
        return err
      }
      return nil
    },
    retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
      return retry.BackOffDelay(n, err, config)
    }),
    retry.Attempts(attempts),
  )
  return (err == nil)
}

func ReadSocket(port serial.Port, logfile *os.File,
                data_flow chan *apolloUtils.ApolloData) {
  log.SetOutput(logfile)
  log.Print("Starting the USB Reader")

  buff := make([]byte, 29)
  var (
    output string
    data *apolloUtils.ApolloData
  )

  for {
    n, err := port.Read(buff)
    if err != nil {
      log.Printf("ERROR: %v", err)
    }
    content := string(buff[:n])
    if strings.HasPrefix(content, "A8") {
      output = strings.Trim(content, "\r\n")
    } else {
      output += strings.Trim(string(content), "\r\n")
    }

    if len(output) == 29 {
      log.Print("Parse data")
      data = apolloUtils.Parse_apollo(output)
      log.Print("Got data")
      data_flow <-data
      time.Sleep(time.Second * 1) // 1 second
      //time.Sleep(time.Millisecond * 500) // 1/2 second
    }
  }
}
