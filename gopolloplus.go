package main
import (
  "fmt"
  "log"
  "os"
  "go.bug.st/serial"
)

func main() {
  var (
    socket string = "/dev/ttyUSB0"
  )
  log_file, err := os.OpenFile("/tmp/info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
  if err != nil {
    log.Fatal(err)
  }
  defer log_file.Close()
  log.SetOutput(log_file)


  if !check_for_socket(socket) {
    log.Print("Socket " + socket + " does not exist!")
    os.Exit(1)
  }

  mode := &serial.Mode{
    BaudRate: 9600,
  }
  port, err := serial.Open(socket, mode)
  if err != nil {
    log.Print(err)
    os.Exit(2)
  }

  buff := make([]byte, 100)
  for {
    n, err := port.Read(buff)
    if err != nil {
      log.Print(err)
    }
    if n == 29 {
      fmt.Printf("%v", string(buff[:n]))
    }
  }

}

func check_for_socket(socket string) (bool) {
  _, err := os.Stat(socket)
  return !os.IsNotExist(err)
}
