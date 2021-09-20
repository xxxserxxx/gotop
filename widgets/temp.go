package widgets

import (
	"fmt"
	"image"
	"sort"
	"time"

	"github.com/gizak/termui/v3"

	"github.com/xxxserxxx/gotop/v4"
	"github.com/xxxserxxx/gotop/v4/devices"
	"github.com/xxxserxxx/gotop/v4/utils"
)

// TODO add thermal history graph. Update when something changes?

type TempWidget struct {
	*termui.Block  // inherits from Block instead of a premade Widget
	updateInterval time.Duration
	Data           map[string]int
	TempThreshold  int
	TempLowColor   termui.Color
	TempHighColor  termui.Color
	TempScale      gotop.TempScale
	temperature    *devices.Temperature
	keys           []string
}

func NewTempWidget(tempScale gotop.TempScale, filter []string) *TempWidget {
	self := &TempWidget{
		Block:          termui.NewBlock(),
		updateInterval: time.Second * 5,
		Data:           make(map[string]int),
		TempThreshold:  80,
		TempScale:      tempScale,
		keys:           make([]string, 0),
	}
	self.Title = tr.Value("widget.label.temp")

	if tempScale == gotop.Fahrenheit {
		self.TempThreshold = utils.CelsiusToFahrenheit(self.TempThreshold)
	}

	return self
}

func (temp *TempWidget) Attach(t *devices.Temperature) {
	temp.temperature = t
	tmps := t.Temps()
	for key, _ := range tmps {
		temp.keys = append(temp.keys, key)
	}
	sort.Strings(temp.keys)

}

// Custom Draw method instead of inheriting from a generic Widget.
func (temp *TempWidget) Draw(buf *termui.Buffer) {
	temp.Block.Draw(buf)

	for y, key := range temp.keys {
		if y+1 > temp.Inner.Dy() {
			break
		}

		var fg termui.Color
		if temp.Data[key] < temp.TempThreshold {
			fg = temp.TempLowColor
		} else {
			fg = temp.TempHighColor
		}

		s := termui.TrimString(key, (temp.Inner.Dx() - 4))
		buf.SetString(s,
			termui.Theme.Default,
			image.Pt(temp.Inner.Min.X, temp.Inner.Min.Y+y),
		)

		temperature := fmt.Sprintf("%3d°%c", temp.Data[key], temp.TempScale)

		buf.SetString(
			temperature,
			termui.NewStyle(fg),
			image.Pt(temp.Inner.Max.X-(len(temperature)-1), temp.Inner.Min.Y+y),
		)
	}
}

func (temp *TempWidget) Update() {
	tmps := temp.temperature.Temps()
	for _, name := range temp.keys {
		val := tmps[name]
		if temp.TempScale == gotop.Fahrenheit {
			temp.Data[name] = utils.CelsiusToFahrenheit(int(val))
		} else {
			temp.Data[name] = int(val)
		}
	}
}
