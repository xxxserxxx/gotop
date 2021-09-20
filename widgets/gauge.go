package widgets

import (
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// LineGraph implements a line graph of data points.
type Gauge struct {
	*widgets.Gauge
}

func NewGauge() *Gauge {
	return &Gauge{
		Gauge: widgets.NewGauge(),
	}
}

func (self *Gauge) Draw(buf *termui.Buffer) {
	self.Gauge.Draw(buf)
	self.Gauge.SetRect(self.Min.X, self.Min.Y, self.Inner.Dx(), self.Inner.Dy())
}
