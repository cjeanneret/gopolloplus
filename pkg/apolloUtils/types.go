package apolloUtils

import "fyne.io/fyne/v2"

type ApolloConfig struct {
  FullScreen bool
  Socket, Logfile, HistoryDir string
  Theme fyne.Theme
}
