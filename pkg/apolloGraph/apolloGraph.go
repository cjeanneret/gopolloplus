package apolloGraph

import (
  "fyne.io/fyne/v2"
  "fyne.io/fyne/v2/canvas"
  "fyne.io/fyne/v2/container"
  "github.com/cjeanneret/gopolloplus/pkg/apolloMonitor"
  "github.com/cjeanneret/gopolloplus/pkg/apolloUtils"
  "image/color"
  "path"
)

const (
  pos_shift = 2
)

func PlotGraph(session string, ct *fyne.Container, cfg *apolloUtils.ApolloConfig, width, height int) {
  fname := path.Join(cfg.HistoryDir, session)
  data := apolloMonitor.LoadCSV(fname)
  watt := []int64{}
  split := []int64{}
  for _, d := range data {
    watt = append(watt, int64(d.Watt))
    split = append(split, int64(d.TimeTo500m))
  }

  _, w_max := apolloUtils.FindMinMax(watt)
  _, s_max := apolloUtils.FindMinMax(split)

  w_scale := float32(1)
  s_scale := float32(1)

  if int(w_max) > height {
    w_scale = float32(w_max)*100/float32(height)
  }
  if int(s_max) > height {
    s_scale = float32(s_max)*100/float32(height)
  }

  w_color := color.RGBA{3, 169, 244, 125}
  s_color := color.RGBA{210, 132, 69, 125}

  _ct := container.NewWithoutLayout()
  _ct.Resize(fyne.Size{Width: float32(width), Height: float32(2*height)})

  for i, _ := range watt {
    if watt[i] != 0 && split[i] != 0 {
      w_size := fyne.Size{Width: pos_shift, Height: w_scale*float32(watt[i])}
      s_size := fyne.Size{Width: pos_shift, Height: s_scale*float32(split[i])}

      pos := fyne.Position{float32(i)*pos_shift*0.5, 100}

      w_rect := canvas.NewRectangle(w_color)
      w_rect.Resize(w_size)
      w_rect.Move(pos)
      s_rect := canvas.NewRectangle(s_color)
      s_rect.Resize(s_size)
      s_rect.Move(pos)
      _ct.Add(w_rect)
      _ct.Add(s_rect)
    }
  }

  scroller := container.NewHScroll(_ct)
  scroller.Resize(fyne.Size{Height: float32(height-100), Width: float32(width)})
  scroller.Move(fyne.Position{310, 0})

  // Empty the existing container
  if len(ct.Objects) == 3 {
    ct.Remove(ct.Objects[2])
  }
  ct.Add(scroller)
}
