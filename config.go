package gotop

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/shibukawa/configdir"
	"github.com/xxxserxxx/gotop/v4/colorschemes"
	"github.com/xxxserxxx/lingo/v2"
)

// TODO: load (additional) colorschemes, layouts, languages from online repo, if user requests
// TODO utility program for benchmarking between versions

//go:embed "dicts/*.toml"
var Dicts embed.FS

// CONFFILE is the name of the default config file
const CONFFILE = "gotop.conf"

type Config struct {
	ConfigDir            configdir.ConfigDir
	GraphHorizontalScale int
	HelpVisible          bool
	Colorscheme          colorschemes.Colorscheme
	UpdateInterval       time.Duration
	AverageLoad          bool
	PercpuLoad           bool
	Statusbar            bool
	TempScale            TempScale
	NetInterface         []string
	Layout               string
	MaxLogSize           int64
	ExportPort           string
	Metrics              *metrics.Set
	Mbps                 bool
	Temps                []string
	Test                 bool
	ExtensionVars        map[string]string
	ConfigFile           string
	Tr                   lingo.Translations
	Nvidia               bool
	NvidiaRefresh        time.Duration
	Headless             bool
	Devices              []string
	Remotes              map[string]Remote
	NoLocal              bool
}

type Remote struct {
	Name    string
	URL     string
	Refresh time.Duration
}

type TempScale rune

const (
	Celsius    TempScale = 'C'
	Fahrenheit           = 'F'
)

func AllDevices() []string {
	return []string{"batt", "cpu", "disk", "mem", "net", "temp"}
}

func NewConfig() Config {
	cd := configdir.New("", "gotop")
	cd.LocalPath, _ = filepath.Abs(".")
	conf := Config{
		ConfigDir:            cd,
		GraphHorizontalScale: 7,
		HelpVisible:          false,
		UpdateInterval:       time.Second,
		AverageLoad:          false,
		PercpuLoad:           true,
		TempScale:            Celsius,
		Statusbar:            false,
		NetInterface:         make([]string, 0),
		MaxLogSize:           5000000,
		Layout:               "default",
		ExtensionVars:        make(map[string]string),
		Remotes:              make(map[string]Remote),
		Devices:              AllDevices(),
		Metrics:              metrics.NewSet(),
	}
	conf.Colorscheme, _ = colorschemes.FromName(conf.ConfigDir, "default")
	folder := conf.ConfigDir.QueryFolderContainsFile(CONFFILE)
	if folder != nil {
		conf.ConfigFile = filepath.Join(folder.Path, CONFFILE)
	}
	return conf
}

func (conf *Config) Load() error {
	var in []byte
	if conf.ConfigFile == "" {
		return nil
	}
	var err error
	if _, err = os.Stat(conf.ConfigFile); os.IsNotExist(err) {
		// Check for the file in the usual suspects
		folder := conf.ConfigDir.QueryFolderContainsFile(conf.ConfigFile)
		if folder == nil {
			return nil
		}
		conf.ConfigFile = filepath.Join(folder.Path, conf.ConfigFile)
	}
	if in, err = ioutil.ReadFile(conf.ConfigFile); err != nil {
		return err
	}
	return load(bytes.NewReader(in), conf)
}

func load(in io.Reader, conf *Config) error {
	r := bufio.NewScanner(in)
	var lineNo int
	for r.Scan() {
		l := strings.TrimSpace(r.Text())
		if len(l) == 0 {
			continue
		}
		if l[0] == '#' {
			continue
		}
		kv := strings.Split(l, "=")
		if len(kv) != 2 {
			return fmt.Errorf(conf.Tr.Value("config.err.configsyntax", l))
		}
		key := strings.ToLower(kv[0])
		ln := strconv.Itoa(lineNo)
		switch key {
		default:
			// Look for remotes
			if strings.HasPrefix(key, "remote-") {
				parts := strings.Split(key, "-")
				if len(parts) == 3 {
					name := parts[1]
					it := conf.Remotes[name]
					if parts[2] == "url" {
						it.URL = kv[1]
					} else if parts[2] == "refresh" {
						var err error
						it.Refresh, err = time.ParseDuration(kv[1])
						if err != nil {
							return err
						}
					} else {
						return fmt.Errorf(conf.Tr.Value("config.err.configsyntax", l))
					}
					conf.Remotes[name] = it
				} else {
					return fmt.Errorf(conf.Tr.Value("config.err.configsyntax", l))
				}
			} else {
				log.Printf(conf.Tr.Value("config.err.unknown", key))
			}
		case "configdir", "logdir", "logfile":
			log.Printf(conf.Tr.Value("config.err.deprecation", ln, key, kv[1]))
		case graphhorizontalscale:
			iv, err := strconv.Atoi(kv[1])
			if err != nil {
				return err
			}
			conf.GraphHorizontalScale = iv
		case helpvisible:
			bv, err := strconv.ParseBool(kv[1])
			if err != nil {
				return fmt.Errorf(conf.Tr.Value("config.err.line", ln, err.Error()))
			}
			conf.HelpVisible = bv
		case colorscheme:
			cs, err := colorschemes.FromName(conf.ConfigDir, kv[1])
			if err != nil {
				return fmt.Errorf(conf.Tr.Value("config.err.line", ln, err.Error()))
			}
			conf.Colorscheme = cs
		case updateinterval:
			iv, err := strconv.Atoi(kv[1])
			if err != nil {
				return fmt.Errorf(conf.Tr.Value("config.err.line", ln, err.Error()))
			}
			conf.UpdateInterval = time.Duration(iv)
		case averagecpu:
			bv, err := strconv.ParseBool(kv[1])
			if err != nil {
				return fmt.Errorf(conf.Tr.Value("config.err.line", ln, err.Error()))
			}
			conf.AverageLoad = bv
		case percpuload:
			bv, err := strconv.ParseBool(kv[1])
			if err != nil {
				return fmt.Errorf(conf.Tr.Value("config.err.line", ln, err.Error()))
			}
			conf.PercpuLoad = bv
		case tempscale:
			switch kv[1] {
			case "C":
				conf.TempScale = 'C'
			case "F":
				conf.TempScale = 'F'
			default:
				conf.TempScale = 'C'
				return fmt.Errorf(conf.Tr.Value("config.err.tempscale", kv[1]))
			}
		case statusbar:
			bv, err := strconv.ParseBool(kv[1])
			if err != nil {
				return fmt.Errorf(conf.Tr.Value("config.err.line", ln, err.Error()))
			}
			conf.Statusbar = bv
		case netinterface:
			conf.NetInterface = strings.Split(kv[1], ",")
		case layout:
			conf.Layout = kv[1]
		case maxlogsize:
			iv, err := strconv.Atoi(kv[1])
			if err != nil {
				return fmt.Errorf(conf.Tr.Value("config.err.line", ln, err.Error()))
			}
			conf.MaxLogSize = int64(iv)
		case export:
			conf.ExportPort = kv[1]
		case mbps:
			conf.Mbps = true
		case temperatures:
			conf.Temps = strings.Split(kv[1], ",")
		case nvidia:
			nv, err := strconv.ParseBool(kv[1])
			if err != nil {
				return fmt.Errorf(conf.Tr.Value("config.err.line", ln, err.Error()))
			}
			conf.Nvidia = nv
		case headless:
			nv, err := strconv.ParseBool(kv[1])
			if err != nil {
				return fmt.Errorf(conf.Tr.Value("config.err.line", ln, err.Error()))
			}
			conf.Headless = nv
		case nolocal:
			nv, err := strconv.ParseBool(kv[1])
			if err != nil {
				return fmt.Errorf(conf.Tr.Value("config.err.line", ln, err.Error()))
			}
			conf.NoLocal = nv
		}
	}

	return nil
}

// Write serializes the configuration to a file.
// The configuration written is based on the loaded configuration, plus any
// command-line changes, so it can be used to update an existing configuration
// file.  The file will be written to the specificed `--config` argument file,
// if one is set; otherwise, it'll create one in the user's config directory.
func (conf *Config) Write() (string, error) {
	var dir *configdir.Config
	var file string = CONFFILE
	if conf.ConfigFile == "" {
		ds := conf.ConfigDir.QueryFolders(configdir.Global)
		if len(ds) == 0 {
			ds = conf.ConfigDir.QueryFolders(configdir.Local)
			if len(ds) == 0 {
				return "", fmt.Errorf("error locating config folders")
			}
		}
		ds[0].CreateParentDir(CONFFILE)
		dir = ds[0]
	} else {
		dir = &configdir.Config{}
		dir.Path = filepath.Dir(conf.ConfigFile)
		file = filepath.Base(conf.ConfigFile)
	}
	marshalled := marshal(conf)
	err := dir.WriteFile(file, marshalled)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir.Path, file), nil
}

func marshal(c *Config) []byte {
	buff := bytes.NewBuffer(nil)
	// TODO i18n for the config comments
	fmt.Fprintln(buff, "# Scale graphs to this level; 7 is the default, 2 is zoomed out.")
	fmt.Fprintf(buff, "%s=%d\n", graphhorizontalscale, c.GraphHorizontalScale)
	fmt.Fprintln(buff, "# If true, start the UI with the help visible")
	fmt.Fprintf(buff, "%s=%t\n", helpvisible, c.HelpVisible)
	fmt.Fprintln(buff, "# The color scheme to use.  See `--list colorschemes`")
	fmt.Fprintf(buff, "%s=%s\n", colorscheme, c.Colorscheme.Name)
	fmt.Fprintln(buff, "# How frequently to update the UI, in nanoseconds")
	fmt.Fprintf(buff, "%s=%d\n", updateinterval, c.UpdateInterval)
	fmt.Fprintln(buff, "# If true, show the average CPU load")
	fmt.Fprintf(buff, "%s=%t\n", averagecpu, c.AverageLoad)
	fmt.Fprintln(buff, "# If true, show load per CPU")
	fmt.Fprintf(buff, "%s=%t\n", percpuload, c.PercpuLoad)
	fmt.Fprintln(buff, "# Temperature units. C for Celcius, F for Fahrenheit")
	fmt.Fprintf(buff, "%s=%c\n", tempscale, c.TempScale)
	fmt.Fprintln(buff, "# If true, display a status bar")
	fmt.Fprintf(buff, "%s=%t\n", statusbar, c.Statusbar)
	fmt.Fprintln(buff, "# The network interface to monitor")
	fmt.Fprintf(buff, "%s=%s\n", netinterface, c.NetInterface)
	fmt.Fprintln(buff, "# A layout name. See `--list layouts`")
	fmt.Fprintf(buff, "%s=%s\n", layout, c.Layout)
	fmt.Fprintln(buff, "# The maximum log file size, in bytes")
	fmt.Fprintf(buff, "%s=%d\n", maxlogsize, c.MaxLogSize)
	fmt.Fprintln(buff, "# If set, export data as Promethius metrics on the interface:port.\n# E.g., `:8080` (colon is required, interface is not)")
	if c.ExportPort == "" {
		fmt.Fprint(buff, "#")
	}
	fmt.Fprintf(buff, "%s=%s\n", export, c.ExportPort)
	fmt.Fprintln(buff, "# Display network IO in mpbs if true")
	fmt.Fprintf(buff, "%s=%t\n", mbps, c.Mbps)
	fmt.Fprintln(buff, "# A list of enabled temp sensors.  See `--list devices`")
	if len(c.Temps) == 0 {
		fmt.Fprint(buff, "#")
	}
	fmt.Fprintf(buff, "%s=%s\n", temperatures, strings.Join(c.Temps, ","))
	fmt.Fprintln(buff, "# Enable NVidia GPU metrics.")
	fmt.Fprintf(buff, "%s=%t\n", nvidia, c.Nvidia)
	fmt.Fprintln(buff, "# To configure the NVidia refresh rate, set a duration:")
	fmt.Fprintln(buff, "#nvidiarefresh=30s")
	fmt.Fprintln(buff, "# Set headless to true to disable TUI; most useful with `export`")
	fmt.Fprintf(buff, "%s=%t\n", headless, c.Headless)
	fmt.Fprintln(buff, "# With headless set, this determines which devices to start up (default is all)")
	fmt.Fprintf(buff, "%s=%s\n", devices, strings.Join(c.Devices, ","))
	return buff.Bytes()
}

const (
	graphhorizontalscale = "graphhorizontalscale"
	helpvisible          = "helpvisible"
	colorscheme          = "colorscheme"
	updateinterval       = "updateinterval"
	averagecpu           = "averagecpu"
	percpuload           = "percpuload"
	tempscale            = "tempscale"
	statusbar            = "statusbar"
	netinterface         = "netinterface"
	layout               = "layout"
	maxlogsize           = "maxlogsize"
	export               = "metricsexportport"
	mbps                 = "mbps"
	temperatures         = "temperatures"
	nvidia               = "nvidia"
	headless             = "headless"
	nolocal              = "nolocal"
	devices              = "devices"
)
