package widgets

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/xxxserxxx/gotop/v4/devices"
	"github.com/xxxserxxx/gotop/v4/utils"
)

type DiskWidget struct {
	*Table
	parts  datas
	factor uint64
}

func NewDiskWidget(refresh time.Duration) *DiskWidget {
	self := &DiskWidget{
		Table:  NewTable(),
		parts:  datas{make([]string, 0), make([]*devices.Partition, 0)},
		factor: uint64(refresh / time.Second),
	}
	self.Table.Tr = tr
	self.Title = tr.Value("widget.label.disk")
	self.Header = []string{tr.Value("widget.disk.disk"), tr.Value("widget.disk.mount"), tr.Value("widget.disk.used"), tr.Value("widget.disk.free"), tr.Value("widget.disk.rs"), tr.Value("widget.disk.ws")}
	self.ColGap = 2
	self.ColResizer = func() {
		self.ColWidths = []int{
			utils.MaxInt(4, (self.Inner.Dx()-29)/2),
			utils.MaxInt(5, (self.Inner.Dx()-29)/2),
			4, 5, 5, 5,
		}
	}

	return self
}

func (disk *DiskWidget) Attach(d devices.Disk) {
	for name, part := range d {
		disk.parts.names = append(disk.parts.names, name)
		disk.parts.parts = append(disk.parts.parts, part)
	}
	sort.Sort(disk.parts)
}

func (disk *DiskWidget) Update() {
	// converts self.Partitions into self.Rows which is a [][]String
	disk.Rows = make([][]string, len(disk.parts.names))

	for i, part := range disk.parts.parts {
		disk.Rows[i] = make([]string, 6)
		disk.Rows[i][0] = strings.Replace(strings.Replace(part.Device, "/dev/", "", -1), "mapper/", "", -1)
		disk.Rows[i][1] = part.MountPoint
		disk.Rows[i][2] = fmt.Sprintf("%d%%", int(part.UsedPercent))
		disk.Rows[i][3] = fmt.Sprintf("%d", part.Free)
		disk.Rows[i][4] = fmt.Sprintf("%d", part.BytesReadRecently/disk.factor)
		disk.Rows[i][5] = fmt.Sprintf("%d", part.BytesWrittenRecently/disk.factor)
	}
}

type datas struct {
	names []string
	parts []*devices.Partition
}

func (d datas) Len() int {
	return len(d.names)
}

func (d datas) Less(i, j int) bool {
	return d.names[i] < d.names[j]
}

func (d datas) Swap(i, j int) {
	d.names[i], d.names[j] = d.names[j], d.names[i]
	d.parts[i], d.parts[j] = d.parts[j], d.parts[i]
}
