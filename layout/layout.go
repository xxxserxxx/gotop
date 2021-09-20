package layout

import (
	"log"
	"sort"
	"strings"

	"github.com/xxxserxxx/gotop/v4"
	"github.com/xxxserxxx/gotop/v4/devices"
	"github.com/xxxserxxx/gotop/v4/widgets"
	"github.com/xxxserxxx/lingo/v2"

	"github.com/gizak/termui/v3"
)

type Layout struct {
	Rows [][]widgetRule
}

type widgetRule struct {
	Widget string
	Weight float64
	Height int
}

type Screen struct {
	*termui.Grid
	Lines   []widgets.Scalable
	Proc    *widgets.ProcWidget
	Net     *widgets.NetWidget
	Widgets widgets.Widgets
}

var widgetNames []string = []string{"cpu", "disk", "mem", "temp", "net", "procs", "batt"}
var tr lingo.Translations

func NewLayout(wl Layout, c gotop.Config, devs map[string]devices.Device) (*Screen, error) {
	tr = c.Tr
	rowDefs := wl.Rows
	uiRows := make([][]interface{}, 0)
	numRows := countNumRows(wl.Rows)
	var uiRow []interface{}
	maxHeight := 0
	heights := make([]int, 0)
	var h int
	grid := &Screen{termui.NewGrid(), nil, nil, nil, nil}
	var wids []widgets.Widget
	for len(rowDefs) > 0 {
		h, uiRow, rowDefs, wids = processRow(c, devs, numRows, rowDefs)
		grid.Widgets = append(grid.Widgets, wids...)
		maxHeight += h
		uiRows = append(uiRows, uiRow)
		heights = append(heights, h)
	}
	rgs := make([]interface{}, 0)
	for i, ur := range uiRows {
		rh := float64(heights[i]) / float64(maxHeight)
		rgs = append(rgs, termui.NewRow(rh, ur...))
	}
	grid.Set(rgs...)
	grid.Lines = deepFindScalable(rgs)
	res := deepFindWidget(uiRows, func(gs interface{}) interface{} {
		p, ok := gs.(*widgets.ProcWidget)
		if ok {
			return p
		}
		return nil
	})
	grid.Proc, _ = res.(*widgets.ProcWidget)
	res = deepFindWidget(uiRows, func(gs interface{}) interface{} {
		p, ok := gs.(*widgets.NetWidget)
		if ok {
			return p
		}
		return nil
	})
	grid.Net, _ = res.(*widgets.NetWidget)
	return grid, nil
}

// processRow eats a single row from the input list of rows and returns a UI
// row (GridItem) representation of the specification, along with a slice
// without that row.
//
// It does more than that, actually, because it may consume more than one row
// if there's a row span widget in the row; in this case, it'll consume as many
// rows as the largest row span object in the row, and produce an uber-row
// containing all that stuff. It returns a slice without the consumed elements.
func processRow(c gotop.Config, devs map[string]devices.Device, numRows int, rowDefs [][]widgetRule) (int, []interface{}, [][]widgetRule, []widgets.Widget) {
	// Recursive function #3.  See the comment in deepFindProc.
	if len(rowDefs) < 1 {
		return 0, nil, nil, nil
	}
	// The height of the tallest widget in this row; the number of rows that
	// will be consumed, and the overall height of the row that will be
	// produced.
	maxHeight := countMaxHeight([][]widgetRule{rowDefs[0]})
	var processing [][]widgetRule
	if maxHeight < len(rowDefs) {
		processing = rowDefs[0:maxHeight]
		rowDefs = rowDefs[maxHeight:]
	} else {
		processing = rowDefs[0:]
		rowDefs = [][]widgetRule{}
	}
	var colWeights []float64
	var columns [][]interface{}
	numCols := len(processing[0])
	if numCols < 1 {
		numCols = 1
	}
	for _, rd := range processing[0] {
		colWeights = append(colWeights, rd.Weight)
		columns = append(columns, make([]interface{}, 0))
	}
	colHeights := make([]int, numCols)
	wids := make([]widgets.Widget, 0)
outer:
	for i, row := range processing {
		// A definition may fill up the columns before all rows are consumed,
		// e.g. cpu/2 net/2.  This block checks for that and, if it occurs,
		// prepends the remaining rows to the "remainder" return value.
		full := true
		for _, ch := range colHeights {
			if ch <= maxHeight {
				full = false
				break
			}
		}
		if full {
			rowDefs = append(processing[i:], rowDefs...)
			break
		}
		// Not all rows have been consumed, so go ahead and place the row's
		// widgets in columns
		for w, widg := range row {
			placed := false
			for k := w; k < len(colHeights); k++ { // there are enough columns
				ch := colHeights[k]
				if ch+widg.Height <= maxHeight {
					widget := makeWidget(c, devs, widg)
					if widget == nil {
						continue
					}
					wids = append(wids, widget)
					columns[k] = append(columns[k], termui.NewRow(float64(widg.Height)/float64(maxHeight), widget))
					colHeights[k] += widg.Height
					placed = true
					break
				}
			}
			// If all columns are full, break out, return the row, and continue processing
			if !placed {
				rowDefs = append(processing[i:], rowDefs...)
				break outer
			}
		}
	}
	var uiColumns []interface{}
	for i, widgets := range columns {
		if len(widgets) > 0 {
			uiColumns = append(uiColumns, termui.NewCol(float64(colWeights[i]), widgets...))
		}
	}

	return maxHeight, uiColumns, rowDefs, wids
}

func makeWidget(c gotop.Config, devs map[string]devices.Device, widRule widgetRule) widgets.Widget {
	var w widgets.Widget
	switch widRule.Widget {
	case "disk":
		dw := widgets.NewDiskWidget(c.UpdateInterval)
		if d, ok := devs["disk"].(devices.Disk); ok {
			dw.Attach(d)
		}
		for _, d := range getRemotes(devs) {
			if r, ok := d.(*devices.Remote); ok {
				dw.Attach(r.DiskData)

			}
		}
		w = dw
	case "cpu":
		cpu := widgets.NewCPUWidget(c.UpdateInterval, c.GraphHorizontalScale, c.AverageLoad, c.PercpuLoad)
		if d, ok := devs["cpu"].(*devices.CPUs); ok {
			cpu.Attach(d)
		}
		if r, ok := devs["nvidia"].(devices.NVidia); ok {
			for _, gpu := range r {
				cpu.Attach(gpu.CPUs)
			}

		}
		for _, d := range getRemotes(devs) {
			if r, ok := d.(*devices.Remote); ok {
				cpu.Attach(r.CPUData)
			}
		}
		assignColors(cpu.Data, c.Colorscheme.CPULines, cpu.LineColors)
		w = cpu
	case "mem":
		m := widgets.NewMemWidget(c.UpdateInterval, c.GraphHorizontalScale)
		if d, ok := devs["mem"].(devices.Memory); ok {
			m.Attach(d)
		}
		if r, ok := devs["nvidia"].(devices.NVidia); ok {
			for _, gpu := range r {
				m.Attach(gpu.Mems)
			}

		}
		for _, d := range getRemotes(devs) {
			if r, ok := d.(*devices.Remote); ok {
				m.Attach(r.MemData)
			}
		}
		assignColors(m.Data, c.Colorscheme.MemLines, m.LineColors)
		w = m
	case "batt":
		b := widgets.NewBatteryWidget(c.GraphHorizontalScale)
		if d, ok := devs["batt"].(*devices.Batteries); ok {
			b.Attach(d)
		}
		assignColors(b.Data, c.Colorscheme.BattLines, b.LineColors)
		w = b
	case "temp":
		t := widgets.NewTempWidget(c.TempScale, c.Temps)
		if d, ok := devs["temp"].(*devices.Temperature); ok {
			t.Attach(d)
		}
		if r, ok := devs["nvidia"].(devices.NVidia); ok {
			for _, gpu := range r {
				t.Attach(gpu.Temps)
			}

		}
		for _, d := range getRemotes(devs) {
			if r, ok := d.(*devices.Remote); ok {
				t.Attach(r.TempData)
			}
		}
		t.TempLowColor = termui.Color(c.Colorscheme.TempLow)
		t.TempHighColor = termui.Color(c.Colorscheme.TempHigh)
		w = t
	case "net":
		n := widgets.NewNetWidget(c.UpdateInterval)
		if d, ok := devs["net"].(*devices.Network); ok {
			n.Attach(d)
		}
		for _, d := range getRemotes(devs) {
			if r, ok := d.(*devices.Remote); ok {
				n.Attach(r.NetData)
			}
		}
		n.Lines[0].LineColor = termui.Color(c.Colorscheme.Sparkline)
		n.Lines[0].TitleColor = termui.Color(c.Colorscheme.BorderLabel)
		n.Lines[1].LineColor = termui.Color(c.Colorscheme.Sparkline)
		n.Lines[1].TitleColor = termui.Color(c.Colorscheme.BorderLabel)
		n.Mbps = c.Mbps
		w = n
	case "procs":
		p := widgets.NewProcWidget()
		p.CursorColor = termui.Color(c.Colorscheme.ProcCursor)
		w = p
	case "power":
		b := widgets.NewBatteryGauge()
		if d, ok := devs["batt"].(*devices.Batteries); ok {
			b.Attach(d)
		}
		b.BarColor = termui.Color(c.Colorscheme.ProcCursor)
		w = b
	default:
		log.Printf(tr.Value("layout.error.widget", widRule.Widget, strings.Join(widgetNames, ",")))
		return nil
	}
	return w
}

func getRemotes(ds map[string]devices.Device) []devices.Device {
	rv := make([]devices.Device, 0)
	for n, d := range ds {
		if strings.HasPrefix(n, "remote-") {
			rv = append(rv, d)
		}
	}
	return rv
}

func assignColors(data map[string][]float64, colors []int, assign map[string]termui.Color) {
	// Make sure the data is always processed in the same order so that
	// colors are assigned to devices consistently
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	i := 0 // For looping around if we run out of colors
	for _, v := range keys {
		if i >= len(colors) {
			i = 0
		}
		assign[v] = termui.Color(colors[i])
		i++
	}
}

func countNumRows(rs [][]widgetRule) int {
	var ttl int
	for len(rs) > 0 {
		ttl++
		line := rs[0]
		h := 1
		for _, c := range line {
			if c.Height > h {
				h = c.Height
			}
		}
		if h < len(rs) {
			rs = rs[h:]
		} else {
			break
		}
	}
	return ttl
}

// Counts the height of the window so rows can be proportionally scaled.
func countMaxHeight(rs [][]widgetRule) int {
	var ttl int
	for len(rs) > 0 {
		line := rs[0]
		h := 1
		for _, c := range line {
			if c.Height > h {
				h = c.Height
			}
		}
		ttl += h
		if h < len(rs) {
			rs = rs[h:]
		} else {
			break
		}
	}
	return ttl
}

// deepFindWidget looks in the UI widget tree for a widget, and returns it if found or nil if not.
func deepFindWidget(gs interface{}, test func(v interface{}) interface{}) interface{} {
	// Recursive function #1.  Recursion is OK here because the number
	// of UI elements, even in a very complex UI, is going to be
	// relatively small.
	t, ok := gs.(termui.GridItem)
	if ok {
		return deepFindWidget(t.Entry, test)
	}
	es, ok := gs.([]termui.GridItem)
	if ok {
		for _, g := range es {
			v := deepFindWidget(g, test)
			if v != nil {
				return v
			}
		}
	}
	fs, ok := gs.([]interface{})
	if ok {
		for _, g := range fs {
			v := deepFindWidget(g, test)
			if v != nil {
				return v
			}
		}
	}
	fs2, ok := gs.([][]interface{})
	if ok {
		for _, g := range fs2 {
			v := deepFindWidget(g, test)
			if v != nil {
				return v
			}
		}
	}
	return test(gs)
}

// deepFindScalable looks in the UI widget tree for Scalable widgets,
// and returns them if found or an empty slice if not.
func deepFindScalable(gs interface{}) []widgets.Scalable {
	// Recursive function #1.  See the comment in deepFindProc.
	t, ok := gs.(termui.GridItem)
	if ok {
		return deepFindScalable(t.Entry)
	}
	es, ok := gs.([]termui.GridItem)
	rvs := make([]widgets.Scalable, 0)
	if ok {
		for _, g := range es {
			vs := deepFindScalable(g)
			rvs = append(rvs, vs...)
		}
		return rvs
	}
	fs, ok := gs.([]interface{})
	if ok {
		for _, g := range fs {
			vs := deepFindScalable(g)
			rvs = append(rvs, vs...)
		}
		return rvs
	}
	p, ok := gs.(widgets.Scalable)
	if ok {
		rvs = append(rvs, p)
	}
	return rvs
}
