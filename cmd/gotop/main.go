package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"time"

	//_ "net/http/pprof"

	"github.com/cloudfoundry-attic/jibber_jabber"
	"github.com/shibukawa/configdir"
	"github.com/xxxserxxx/lingo/v2"
	"github.com/xxxserxxx/opflag"

	"github.com/xxxserxxx/gotop/v4"
	"github.com/xxxserxxx/gotop/v4/colorschemes"
	"github.com/xxxserxxx/gotop/v4/devices"
	"github.com/xxxserxxx/gotop/v4/logging"
	"github.com/xxxserxxx/gotop/v4/tui"
)

// TODO add build flags to disable TUI (headless builds)

var (
	// Version of the program; set during build from git tags
	Version = "0.0.0"
	// BuildDate when the program was compiled; set during build
	BuildDate    = "Hadean"
	conf         gotop.Config
	stderrLogger = log.New(os.Stderr, "", 0)
	tr           lingo.Translations
)

func parseArgs() error {
	cds := conf.ConfigDir.QueryFolders(configdir.All)
	cpaths := make([]string, len(cds))
	for i, p := range cds {
		cpaths[i] = p.Path
	}
	help := opflag.BoolP("help", "h", false, tr.Value("args.help"))
	color := opflag.StringP("color", "c", conf.Colorscheme.Name, tr.Value("args.color"))
	opflag.IntVarP(&conf.GraphHorizontalScale, "graphscale", "S", conf.GraphHorizontalScale, tr.Value("args.scale"))
	version := opflag.BoolP("version", "v", false, tr.Value("args.version"))
	versioN := opflag.BoolP("", "V", false, tr.Value("args.version"))
	opflag.BoolVarP(&conf.PercpuLoad, "percpu", "p", conf.PercpuLoad, tr.Value("args.percpu"))
	opflag.BoolVarP(&conf.AverageLoad, "averagecpu", "a", conf.AverageLoad, tr.Value("args.cpuavg"))
	fahrenheit := opflag.BoolP("fahrenheit", "f", conf.TempScale == 'F', tr.Value("args.temp"))
	opflag.BoolVarP(&conf.Statusbar, "statusbar", "s", conf.Statusbar, tr.Value("args.statusbar"))
	opflag.DurationVarP(&conf.UpdateInterval, "rate", "r", conf.UpdateInterval, tr.Value("args.rate"))
	opflag.StringVarP(&conf.Layout, "layout", "l", conf.Layout, tr.Value("args.layout"))
	ifaces := opflag.String("interface", "", tr.Value("args.net"))
	opflag.StringVarP(&conf.ExportPort, "export", "x", conf.ExportPort, tr.Value("args.export"))
	opflag.BoolVarP(&conf.Mbps, "mbps", "", conf.Mbps, tr.Value("args.mbps"))
	opflag.BoolVar(&conf.Test, "test", conf.Test, tr.Value("args.test"))
	opflag.StringP("", "C", "", tr.Value("args.conffile"))
	opflag.BoolVarP(&conf.Nvidia, "nvidia", "", conf.Nvidia, "Enable NVidia GPU support")
	remoteName := opflag.String("remote-name", "", "Remote: name of remote gotop")
	remoteURL := opflag.String("remote-url", "", "Remote: URL of remote gotop")
	remoteRefresh := opflag.Duration("remote-refresh", 0, "Remote: Frequency to refresh data, in seconds")
	opflag.BoolVarP(&conf.NoLocal, "no-local", "", false, "Disable local(host) sensors")
	opflag.BoolVarP(&conf.Headless, "headless", "", conf.Headless, "Disable user interface")
	list := opflag.String("list", "", tr.Value("args.list"))
	wc := opflag.Bool("write-config", false, tr.Value("args.write"))
	devices := opflag.String("devices", "", tr.Value("args.devices"))
	opflag.SortFlags = false
	opflag.Usage = func() {
		fmt.Fprintf(os.Stderr, tr.Value("usage", os.Args[0]))
		opflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "Project home: https://github.com/xxxserxxx/gotop\n")
	}
	opflag.Parse()
	if *version || *versioN {
		fmt.Printf("gotop %s (%s)\n", Version, BuildDate)
		os.Exit(0)
	}
	if *help {
		opflag.Usage()
		os.Exit(0)
	}
	cs, err := colorschemes.FromName(conf.ConfigDir, *color)
	if err != nil {
		return err
	}
	if *devices != "" {
		conf.Devices = strings.Split("devices", ",")
		if len(conf.Devices) == 0 {
			conf.Devices = gotop.AllDevices()
		}
	}
	conf.Colorscheme = cs
	if *fahrenheit {
		conf.TempScale = 'F'
	} else {
		conf.TempScale = 'C'
	}
	if *ifaces != "" {
		conf.NetInterface = strings.Split(*ifaces, ",")
	}
	if *list != "" {
		switch *list {
		case "layouts":
			fmt.Println(tr.Value("help.layouts"))
		case "colorschemes":
			fmt.Println(tr.Value("help.colorschemes"))
		case "paths":
			fmt.Println(tr.Value("help.paths"))
			paths := make([]string, 0)
			for _, d := range conf.ConfigDir.QueryFolders(configdir.All) {
				paths = append(paths, d.Path)
			}
			fmt.Println(strings.Join(paths, "\n"))
			fmt.Println()
			fmt.Println(tr.Value("help.log", filepath.Join(conf.ConfigDir.QueryCacheFolder().Path, logging.LOGFILE)))
		case "devices":
			listDevices()
		case "keys":
			fmt.Println(tr.Value("help.help"))
		case "widgets":
			fmt.Println(tr.Value("help.widgets"))
		case "langs":
			err := fs.WalkDir(gotop.Dicts, ".", func(pth string, info fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() { // We skip these
					return nil
				}
				fileName := info.Name()
				if strings.HasSuffix(fileName, ".toml") {
					fmt.Println(strings.TrimSuffix(fileName, ".toml"))
				}
				return nil
			})
			if err != nil {
				return err
			}
		default:
			fmt.Printf(tr.Value("error.unknownopt", *list))
			os.Exit(1)
		}
		os.Exit(0)
	}
	if conf.Nvidia {
		conf.ExtensionVars["nvidia"] = "true"
	}
	if *remoteURL != "" {
		if u, e := url.Parse(*remoteURL); e == nil {
			r := gotop.Remote{}
			r.URL = *remoteURL
			if remoteName != nil {
				r.Name = *remoteName
			} else {
				r.Name = u.Hostname()
			}
			if remoteRefresh != nil {
				r.Refresh = *remoteRefresh
			} else {
				r.Refresh = 5 * time.Second
			}
			conf.Remotes[r.Name] = r
		} else {
			fmt.Println(e)
		}
	}
	if *wc {
		path, err := conf.Write()
		if err != nil {
			fmt.Println(tr.Value("error.writefail", err.Error()))
			os.Exit(1)
		}
		fmt.Println(tr.Value("help.written", path))
		os.Exit(0)
	}
	return nil
}

func main() {
	//go func() {
	//	log.Fatal(http.ListenAndServe(":7777", http.DefaultServeMux))
	//}()

	var ec int
	defer func() {
		if ec > 0 {
			if ec < 2 {
				logpath := filepath.Join(conf.ConfigDir.QueryCacheFolder().Path, logging.LOGFILE)
				fmt.Println(tr.Value("error.checklog", logpath))
				bs, _ := ioutil.ReadFile(logpath)
				fmt.Println(string(bs))
			}
		}
		os.Exit(ec)
	}()

	ling, err := lingo.New("en_US", ".", gotop.Dicts)
	if err != nil {
		fmt.Printf("failed to load language files: %s\n", err)
		ec = 2
		return
	}
	lang, err := jibber_jabber.DetectIETF()
	if err != nil {
		lang = "en_US"
	}
	lang = strings.Replace(lang, "-", "_", -1)
	// Get the locale from the os
	tr = ling.TranslationsForLocale(lang)
	colorschemes.SetTr(tr)
	conf = gotop.NewConfig()
	conf.Tr = tr
	// Find the config file; look in (1) local, (2) user, (3) global
	// Check the last argument first
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	cfg := fs.String("C", "", tr.Value("configfile"))
	fs.SetOutput(bufio.NewWriter(nil))
	fs.Parse(os.Args[1:])
	if *cfg != "" {
		conf.ConfigFile = *cfg
	}
	err = conf.Load()
	if err != nil {
		fmt.Println(tr.Value("error.configparse", err.Error()))
		ec = 2
		return
	}
	// Override with command line arguments
	err = parseArgs()
	if err != nil {
		fmt.Println(tr.Value("error.cliparse", err.Error()))
		ec = 2
		return
	}

	logfile, err := logging.New(conf)
	if err != nil {
		fmt.Println(tr.Value("logsetup", err.Error()))
		ec = 2
		return
	}
	defer logfile.Close()

	// device initialization errors do not stop execution
	// Build a list of requested devices
	devs := make(map[string]bool)

	if len(conf.Remotes) > 0 {
		devs["remote"] = true
	}
	if conf.Nvidia {
		devs["nvidia"] = true
	}

	if conf.Test {
		ec = runTests(conf)
		return
	}

	// TODO https://godoc.org/github.com/VictoriaMetrics/metrics#Set
	if conf.ExportPort != "" {
		go func() {
			http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
				conf.Metrics.WritePrometheus(w)
			})
			http.ListenAndServe(conf.ExportPort, nil)
		}()
	}

	if conf.Headless {
		if conf.ExportPort == "" {
			fmt.Fprintln(os.Stdout, "metrics not being exported; did you forget --export?")
			ec = 1
			return
		}
		devs := make([]string, len(conf.Devices))
		// First, check the layout; each widget has a default associated device
		if !conf.NoLocal {
			for i, dn := range conf.Devices {
				devs[i] = dn
			}
		}
		devInsts, errs := devices.Startup(devs, conf)
		for _, err := range errs {
			stderrLogger.Print(err)
			ec = 1
			return
		}
		devices.Spawn(devInsts, conf)
		// No TUI; just wait for user to interrupt
		fmt.Println("gotop running... press ^c to exit")
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
	} else {
		ui, err := tui.New(conf)
		if err != nil {
			stderrLogger.Print(err)
			ec = 1
			return
		}
		defer ui.ShutdownUI()
		err = ui.LoopUI()
		if err != nil {
			stderrLogger.Print(err)
			ec = 1
			return
		}
	}

	ec = 0
	return
}

func runTests(_ gotop.Config) int {
	fmt.Printf("PASS")
	return 0
}

func listDevices() {
	ms := devices.Domains()
	sort.Strings(ms)
	for _, m := range ms {
		fmt.Printf("%s:\n", m)
		for _, d := range devices.Devices(m, true) {
			fmt.Printf("\t%s\n", d)
		}
	}
}
