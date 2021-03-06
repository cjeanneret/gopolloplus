package apolloUI

import (
  "fmt"
  "fyne.io/fyne/v2"
  "fyne.io/fyne/v2/canvas"
  "github.com/cjeanneret/gopolloplus/pkg/apolloUtils"
  "image/color"
)

const (
  BarHeight = 50
  GraphWidth = 300
  TitleFontSize = 15
)
var (
  ValueColor = color.RGBA{3, 169, 244, 255}
  CurrColor = color.Gray{Y: 110}
  AVGColor = color.Gray{Y: 85}
  MaxColor = color.Gray{Y: 65}
)

func ResizeCanvas(values []uint64, curr, avg, max *canvas.Rectangle,
                  curr_txt, avg_txt, max_txt *canvas.Text) {

  curr_val := float32(values[len(values)-1])
  _, max_val := apolloUtils.FindMinMax(values)
  avg_val := float32(apolloUtils.Average(values))

  facter := float32(1)
  // match max to graphWidth
  if max_val > GraphWidth {
    facter = float32(max_val)*100.0/float32(GraphWidth)
  }

  max.Resize(fyne.Size{Height: BarHeight, Width: float32(max_val)*facter})
  max_txt.Text = fmt.Sprintf("%.2f", float32(max_val))
  max_txt.Refresh()

  curr.Resize(fyne.Size{Height: BarHeight, Width: curr_val*facter})
  curr_txt.Text = fmt.Sprintf("%.2f", curr_val)
  curr_txt.Refresh()

  avg.Resize(fyne.Size{Height: BarHeight, Width: avg_val*facter})
  avg_txt.Text = fmt.Sprintf("%.2f", avg_val)
  avg_txt.Refresh()

}
