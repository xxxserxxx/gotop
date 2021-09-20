package widgets

import (
	"image"
	"log"
	"os"
	"time"

	"github.com/gizak/termui/v3"
)

type StatusBar struct {
	termui.Block
}

func NewStatusBar() *StatusBar {
	self := &StatusBar{*termui.NewBlock()}
	self.Border = false
	return self
}

func (sb *StatusBar) Draw(buf *termui.Buffer) {
	sb.Block.Draw(buf)

	hostname, err := os.Hostname()
	if err != nil {
		log.Printf(tr.Value("error.nohostname", err.Error()))
		return
	}
	buf.SetString(
		hostname,
		termui.Theme.Default,
		image.Pt(sb.Inner.Min.X, sb.Inner.Min.Y+(sb.Inner.Dy()/2)),
	)

	currentTime := time.Now()
	formattedTime := currentTime.Format("15:04:05")
	buf.SetString(
		formattedTime,
		termui.Theme.Default,
		image.Pt(
			sb.Inner.Min.X+(sb.Inner.Dx()/2)-len(formattedTime)/2,
			sb.Inner.Min.Y+(sb.Inner.Dy()/2),
		),
	)

	// i, e := host.Info()
	// i.Uptime // Number of seconds since boot
	buf.SetString(
		"gotop",
		termui.Theme.Default,
		image.Pt(
			sb.Inner.Max.X-6,
			sb.Inner.Min.Y+(sb.Inner.Dy()/2),
		),
	)
}
