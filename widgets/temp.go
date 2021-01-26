package widgets

import (
	"fmt"
	"image"
	"sort"
	"strings"
	"time"

	"github.com/VictoriaMetrics/metrics"
	ui "github.com/gizak/termui/v3"

	"github.com/xxxserxxx/gotop/v4/devices"
	"github.com/xxxserxxx/gotop/v4/utils"
)

type TempScale rune

const (
	Celsius    TempScale = 'C'
	Fahrenheit           = 'F'
)

type TempWidget struct {
	*ui.Block      // inherits from Block instead of a premade Widget
	updateInterval time.Duration
	Data           map[string]int
	TempThreshold  int
	TempLowColor   ui.Color
	TempHighColor  ui.Color
	TempScale      TempScale
	temps          map[string]float64
	keys           ordered
}

func NewTempWidget(tempScale TempScale, filter []string) *TempWidget {
	self := &TempWidget{
		Block:          ui.NewBlock(),
		updateInterval: time.Second * 5,
		Data:           make(map[string]int),
		TempThreshold:  80,
		TempScale:      tempScale,
	}
	self.Title = tr.Value("widget.label.temp")
	if len(filter) > 0 {
		for _, t := range filter {
			self.Data[t] = 0
		}
	} else {
		for _, t := range devices.Devices(devices.Temperatures, false) {
			self.Data[t] = 0
		}
	}

	if tempScale == Fahrenheit {
		self.TempThreshold = utils.CelsiusToFahrenheit(self.TempThreshold)
	}

	self.update()

	go func() {
		for range time.NewTicker(self.updateInterval).C {
			self.Lock()
			self.update()
			self.Unlock()
		}
	}()

	return self
}

func (temp *TempWidget) EnableMetric() {
	temp.temps = make(map[string]float64)
	for k, _ := range temp.Data {
		kc := k
		metrics.NewGauge(makeName("temp", k), func() float64 {
			return float64(temp.Data[kc])
		})
	}
}

// Custom Draw method instead of inheriting from a generic Widget.
func (temp *TempWidget) Draw(buf *ui.Buffer) {
	temp.Block.Draw(buf)

	if len(temp.keys) != len(temp.Data) {
		temp.keys = make(ordered, len(temp.Data))
		i := 0
		for key := range temp.Data {
			temp.keys[i] = key
			i++
		}
		sort.Sort(temp.keys)
	}

	for y, key := range temp.keys {
		if y+1 > temp.Inner.Dy() {
			break
		}

		var fg ui.Color
		if temp.Data[key] < temp.TempThreshold {
			fg = temp.TempLowColor
		} else {
			fg = temp.TempHighColor
		}

		s := ui.TrimString(key, (temp.Inner.Dx() - 4))
		buf.SetString(s,
			ui.Theme.Default,
			image.Pt(temp.Inner.Min.X, temp.Inner.Min.Y+y),
		)

		temperature := fmt.Sprintf("%3d°%c", temp.Data[key], temp.TempScale)

		buf.SetString(
			temperature,
			ui.NewStyle(fg),
			image.Pt(temp.Inner.Max.X-(len(temperature)-1), temp.Inner.Min.Y+y),
		)
	}
}

func (temp *TempWidget) update() {
	devices.UpdateTemps(temp.Data)
	for name, val := range temp.Data {
		if temp.TempScale == Fahrenheit {
			temp.Data[name] = utils.CelsiusToFahrenheit(val)
		} else {
			temp.Data[name] = val
		}
	}
}

type ordered []string

func (o ordered) Len() int      { return len(o) }
func (o ordered) Swap(i, j int) { o[j], o[i] = o[i], o[j] }
func (o ordered) Less(i, j int) bool {
	if strings.HasPrefix(o[i], "core") {
		if !strings.HasPrefix(o[j], "core") {
			return true
		}
	} else if strings.HasPrefix(o[j], "core") {
		return false
	}
	return o[i] < o[j]
}
