//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package option

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

const (
	DefaultChroot = "/"
)

const (
	envKeyChroot            = "GHW_CHROOT"
	envKeyDisableWarnings   = "GHW_DISABLE_WARNINGS"
	envKeyDisableTools      = "GHW_DISABLE_TOOLS"
	envKeySnapshotPath      = "GHW_SNAPSHOT_PATH"
	envKeySnapshotRoot      = "GHW_SNAPSHOT_ROOT"
	envKeySnapshotExclusive = "GHW_SNAPSHOT_EXCLUSIVE"
	envKeySnapshotPreserve  = "GHW_SNAPSHOT_PRESERVE"
)

// Alerter emits warnings about undesirable but recoverable errors.
// We use a subset of a logger interface only to emit warnings, and
// `Warninger` sounded ugly.
type Alerter interface {
	Printf(format string, v ...interface{})
}

var (
	NullAlerter = log.New(ioutil.Discard, "", 0)
)

// EnvOrDefaultAlerter returns the default instance ghw will use to emit
// its warnings. ghw will emit warnings to stderr by default unless the
// environs variable GHW_DISABLE_WARNINGS is specified; in the latter case
// all warning will be suppressed.
func EnvOrDefaultAlerter() Alerter {
	var dest io.Writer
	if _, exists := os.LookupEnv(envKeyDisableWarnings); exists {
		dest = ioutil.Discard
	} else {
		// default
		dest = os.Stderr
	}
	return log.New(dest, "", 0)
}

// EnvOrDefaultChroot returns the value of the GHW_CHROOT environs variable or
// the default value of "/" if not set
func EnvOrDefaultChroot() string {
	// Grab options from the environs by default
	if val, exists := os.LookupEnv(envKeyChroot); exists {
		return val
	}
	return DefaultChroot
}

// EnvOrDefaultSnapshotPath returns the value of the GHW_SNAPSHOT_PATH environs variable
// or the default value of "" (disable snapshot consumption) if not set
func EnvOrDefaultSnapshotPath() string {
	if val, exists := os.LookupEnv(envKeySnapshotPath); exists {
		return val
	}
	return "" // default is no snapshot
}

// EnvOrDefaultSnapshotRoot returns the value of the the GHW_SNAPSHOT_ROOT environs variable
// or the default value of "" (self-manage the snapshot unpack directory, if relevant) if not set
func EnvOrDefaultSnapshotRoot() string {
	if val, exists := os.LookupEnv(envKeySnapshotRoot); exists {
		return val
	}
	return "" // default is to self-manage the snapshot directory
}

// EnvOrDefaultSnapshotExclusive returns the value of the GHW_SNAPSHOT_EXCLUSIVE environs variable
// or the default value of false if not set
func EnvOrDefaultSnapshotExclusive() bool {
	if _, exists := os.LookupEnv(envKeySnapshotExclusive); exists {
		return true
	}
	return false
}

// EnvOrDefaultSnapshotPreserve returns the value of the GHW_SNAPSHOT_PRESERVE environs variable
// or the default value of false if not set
func EnvOrDefaultSnapshotPreserve() bool {
	if _, exists := os.LookupEnv(envKeySnapshotPreserve); exists {
		return true
	}
	return false
}

// EnvOrDefaultTools return true if ghw should use external tools to augment the data collected
// from sysfs. Most users want to do this most of time, so this is enabled by default.
// Users consuming snapshots may want to opt out, thus they can set the GHW_DISABLE_TOOLS
// environs variable to any value to make ghw skip calling external tools even if they are available.
func EnvOrDefaultTools() bool {
	if _, exists := os.LookupEnv(envKeyDisableTools); exists {
		return false
	}
	return true
}

// Option is used to represent optionally-configured settings. Each field is a
// pointer to some concrete value so that we can tell when something has been
// set or left unset.
type Option struct {
	// To facilitate querying of sysfs filesystems that are bind-mounted to a
	// non-default root mountpoint, we allow users to set the GHW_CHROOT environ
	// variable to an alternate mountpoint. For instance, assume that the user of
	// ghw is a Golang binary being executed from an application container that has
	// certain host filesystems bind-mounted into the container at /host. The user
	// would ensure the GHW_CHROOT environ variable is set to "/host" and ghw will
	// build its paths from that location instead of /
	Chroot *string

	// Snapshot contains options for handling ghw snapshots
	Snapshot *SnapshotOptions

	// Alerter contains the target for ghw warnings
	Alerter Alerter

	// EnableTools optionally request ghw to not call any external program to learn
	// about the hardware. The default is to use such tools if available.
	EnableTools *bool

	// PathOverrides optionally allows to override the default paths ghw uses internally
	// to learn about the system resources.
	PathOverrides PathOverrides

	// Context may contain a pointer to a `Context` struct that is constructed
	// during a call to the `context.WithContext` function. Only used internally.
	// This is an interface to get around recursive package import issues.
	Context interface{}
}

// SnapshotOptions contains options for handling of ghw snapshots
type SnapshotOptions struct {
	// Path allows users to specify a snapshot (captured using ghw-snapshot) to be
	// automatically consumed. Users need to supply the path of the snapshot, and
	// ghw will take care of unpacking it on a temporary directory.
	// Set the environment variable "GHW_SNAPSHOT_PRESERVE" to make ghw skip the cleanup
	// stage and keep the unpacked snapshot in the temporary directory.
	Path string
	// Root is the directory on which the snapshot must be unpacked. This allows
	// the users to manage their snapshot directory instead of ghw doing that on
	// their behalf. Relevant only if SnapshotPath is given.
	Root *string
	// Exclusive tells ghw if the given directory should be considered of exclusive
	// usage of ghw or not If the user provides a Root. If the flag is set, ghw will
	// unpack the snapshot in the given SnapshotRoot iff the directory is empty; otherwise
	// any existing content will be left untouched and the unpack stage will exit silently.
	// As additional side effect, give both this option and SnapshotRoot to make each
	// context try to unpack the snapshot only once.
	Exclusive bool
}

// WithChroot allows to override the root directory ghw uses.
func WithChroot(dir string) *Option {
	return &Option{Chroot: &dir}
}

// WithSnapshot sets snapshot-processing options for a ghw run
func WithSnapshot(opts SnapshotOptions) *Option {
	return &Option{
		Snapshot: &opts,
	}
}

// WithAlerter sets alerting options for ghw
func WithAlerter(alerter Alerter) *Option {
	return &Option{
		Alerter: alerter,
	}
}

// WithNullAlerter sets No-op alerting options for ghw
func WithNullAlerter() *Option {
	return &Option{
		Alerter: NullAlerter,
	}
}

// WithDisableTools sets enables or prohibts ghw to call external tools to discover hardware capabilities.
func WithDisableTools() *Option {
	false_ := false
	return &Option{EnableTools: &false_}
}

// PathOverrides is a map, keyed by the string name of a mount path, of override paths
type PathOverrides map[string]string

// WithPathOverrides supplies path-specific overrides for the context
func WithPathOverrides(overrides PathOverrides) *Option {
	return &Option{
		PathOverrides: overrides,
	}
}

// There is intentionally no Option related to GHW_SNAPSHOT_PRESERVE because we see that as
// a debug/troubleshoot aid more something users wants to do regularly.
// Hence we allow that only via the environment variable for the time being.

// Merge accepts one or more Options and merges them together, returning the
// merged Option
func Merge(opts ...*Option) *Option {
	merged := &Option{}
	for _, opt := range opts {
		if opt.Chroot != nil {
			merged.Chroot = opt.Chroot
		}
		if opt.Snapshot != nil {
			merged.Snapshot = opt.Snapshot
		}
		if opt.Alerter != nil {
			merged.Alerter = opt.Alerter
		}
		if opt.EnableTools != nil {
			merged.EnableTools = opt.EnableTools
		}
		// intentionally only programmatically
		if opt.PathOverrides != nil {
			merged.PathOverrides = opt.PathOverrides
		}
		if opt.Context != nil {
			merged.Context = opt.Context
		}
	}
	// Set the default value if missing from mergeOpts
	if merged.Chroot == nil {
		chroot := EnvOrDefaultChroot()
		merged.Chroot = &chroot
	}
	if merged.Alerter == nil {
		merged.Alerter = EnvOrDefaultAlerter()
	}
	if merged.Snapshot == nil {
		snapRoot := EnvOrDefaultSnapshotRoot()
		merged.Snapshot = &SnapshotOptions{
			Path:      EnvOrDefaultSnapshotPath(),
			Root:      &snapRoot,
			Exclusive: EnvOrDefaultSnapshotExclusive(),
		}
	}
	if merged.EnableTools == nil {
		enabled := EnvOrDefaultTools()
		merged.EnableTools = &enabled
	}
	return merged
}
