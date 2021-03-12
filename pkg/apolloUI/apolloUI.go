package apolloUI

import (
  "fmt"
  "fyne.io/fyne/v2"
  "fyne.io/fyne/v2/canvas"
  "fyne.io/fyne/v2/theme"
  "github.com/cjeanneret/gopolloplus/pkg/apolloUtils"
  "image/color"
  "time"
)

const (
  BarHeight = 50
  GraphWidth = 300
  TitleFontSize = 26
  ValueFontSize = 18
)
var (
  ValueColor = color.RGBA{3, 169, 244, 255}
  CurrColor = color.Gray{Y: 110}
  AVGColor = color.Gray{Y: 85}
  MaxColor = color.Gray{Y: 65}
  WhiteColor = color.Gray{Y: 255}
)

func TimeCanvas(title string) (c *canvas.Text) {
  c = &canvas.Text{Color: theme.TextColor(),
                   TextSize: TitleFontSize,
                   Text: title,
                   TextStyle: fyne.TextStyle{Bold: true}}
  return
}

func CreateCanvas(lshift, down float32, c color.Color) (rect *canvas.Rectangle,
                                                        rect_txt *canvas.Text) {

  rect = canvas.NewRectangle(c)
  rect.Resize(fyne.Size{Height: BarHeight, Width: 3})
  rect.Move(fyne.Position{lshift, down})

  rect_txt = &canvas.Text{Color: ValueColor,
                          TextSize: ValueFontSize,
                          TextStyle: fyne.TextStyle{Bold: true}}
  rect_txt.Move(fyne.Position{lshift+10, down+20})

  return

}

func ResizeCanvas(values []uint64, curr, avg, max *canvas.Rectangle,
                  curr_txt, avg_txt, max_txt *canvas.Text,
                  is_duration bool) {

  curr_val := float32(values[len(values)-1])
  _, max_val := apolloUtils.FindMinMax(values)
  avg_val := float32(apolloUtils.Average(values))

  facter := float32(1)
  // match max to graphWidth
  if max_val > GraphWidth {
    facter = float32(max_val)*100.0/float32(GraphWidth)
  }

  max.Resize(fyne.Size{Height: BarHeight, Width: float32(max_val)*facter})
  curr.Resize(fyne.Size{Height: BarHeight, Width: curr_val*facter})
  avg.Resize(fyne.Size{Height: BarHeight, Width: avg_val*facter})

  if is_duration {
    d, _ := time.ParseDuration(fmt.Sprintf("%vs", curr_val))
    curr_txt.Text = fmt.Sprintf("%v", d)
    d, _ = time.ParseDuration(fmt.Sprintf("%vs", avg_val))
    avg_txt.Text = fmt.Sprintf("%v", d)
    d, _ = time.ParseDuration(fmt.Sprintf("%vs", max_val))
    max_txt.Text = fmt.Sprintf("%v", d)
  } else {
    curr_txt.Text = fmt.Sprintf("%.2f", curr_val)
    avg_txt.Text = fmt.Sprintf("%.2f", avg_val)
    max_txt.Text = fmt.Sprintf("%.2f", float32(max_val))
  }

  curr_txt.Refresh()
  avg_txt.Refresh()
  max_txt.Refresh()

}
