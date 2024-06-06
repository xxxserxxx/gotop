package goopt

// An almost-drop-in replacement for flag.  It is intended to work
// basically the same way, but to parse flags like getopt does.

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

var opts = make([]opt, 0, 8)

// Redefine this function to change the way usage is printed
var Usage = func() string {
	programName := os.Args[0][strings.LastIndex(os.Args[0], "/")+1:]
	usage := fmt.Sprintf("Usage of %s:\n", programName)
	if Summary != "" {
		usage += fmt.Sprintf("\t%s", Summary)
	}
	usage += fmt.Sprintf("\n%s", Help())
	if ExtraUsage != "" {
		usage += fmt.Sprintf("%s\n", ExtraUsage)
	}
	return usage
}

// Redefine this to change the summary of your program (used in the
// default Usage() and man page)
var Summary = ""

// Redefine this to change the additional usage of your program (used in the
// default Usage() and man page)
var ExtraUsage = ""

// Redefine this to change the author of your program (used in the
// default man page)
var Author = ""

// Redefine this to change the displayed version of your program (used
// in the default man page)
var Version = ""

// Redefine this to change the suite of your program (e.g. the
// application name) (used in the default manpage())
var Suite = ""

// Redefine this to force flags to come before all options or be
// treated as if they were options
var RequireOrder = false

// Variables for expansion using Expand(), which is automatically
// called on help text for flags
var Vars = make(map[string]string)

// Expand all variables in Vars within the given string.  This does
// not assume any prefix or suffix that sets off a variable from the
// rest of the text, so a var of A set to HI expanded into HAPPY will
// become HHIPPY.
func Expand(x string) string {
	for k, v := range Vars {
		x = strings.Join(strings.Split(x, k), v)
	}
	return x
}

// Override the way help is displayed (not recommended)
var Help = func() string {
	h0 := new(bytes.Buffer)
	h := tabwriter.NewWriter(h0, 0, 8, 2, ' ', 0)
	if len(opts) > 1 {
		fmt.Fprintln(h, "Options:")
	}
	for _, o := range opts {
		fmt.Fprint(h, "  ")
		if len(o.shortnames) > 0 {
			for _, sn := range o.shortnames[0 : len(o.shortnames)-1] {
				fmt.Fprintf(h, "-%c, ", sn)
			}
			fmt.Fprintf(h, "-%c", o.shortnames[len(o.shortnames)-1])
			if o.allowsArg != nil && len(o.names) == 0 {
				fmt.Fprintf(h, " %s", *o.allowsArg)
			}
		}
		if len(o.names) > 0 {
			if len(o.shortnames) > 0 {
				fmt.Fprint(h, ", ")
			}
			for _, n := range o.names[0 : len(o.names)-1] {
				fmt.Fprintf(h, "%s, ", n)
			}
			fmt.Fprint(h, o.names[len(o.names)-1])
			if o.allowsArg != nil {
				fmt.Fprintf(h, "=%s", *o.allowsArg)
			}
		}
		fmt.Fprintf(h, "\t%v\n", Expand(o.help))
	}
	h.Flush()
	return h0.String()
}

// Override the shortened help for your program (not recommended)
var Synopsis = func() string {
	h := new(bytes.Buffer)
	for _, o := range opts {
		fmt.Fprint(h, " [")
		switch {
		case len(o.shortnames) == 0:
			for _, n := range o.names[0 : len(o.names)-1] {
				fmt.Fprintf(h, "\\-\\-%s|", n[2:])
			}
			fmt.Fprintf(h, "\\-\\-%s", o.names[len(o.names)-1][2:])
			if o.allowsArg != nil {
				fmt.Fprintf(h, " %s", *o.allowsArg)
			}
		case len(o.names) == 0:
			for _, c := range o.shortnames[0 : len(o.shortnames)-1] {
				fmt.Fprintf(h, "\\-%c|", c)
			}
			fmt.Fprintf(h, "\\-%c", o.shortnames[len(o.shortnames)-1])
			if o.allowsArg != nil {
				fmt.Fprintf(h, " %s", *o.allowsArg)
			}
		default:
			for _, c := range o.shortnames {
				fmt.Fprintf(h, "\\-%c|", c)
			}
			for _, n := range o.names[0 : len(o.names)-1] {
				fmt.Fprintf(h, "\\-\\-%s|", n[2:])
			}
			fmt.Fprintf(h, "\\-\\-%s", o.names[len(o.names)-1][2:])
			if o.allowsArg != nil {
				fmt.Fprintf(h, " %s", *o.allowsArg)
			}
		}
		fmt.Fprint(h, "]")
	}
	return h.String()
}

// Set the description used in the man page for your program.  If you
// want paragraphs, use two newlines in a row (e.g. LaTeX)
var Description = func() string {
	return `To add a description to your program, define goopt.Description.

If you want paragraphs, just use two newlines in a row, like latex.`
}

type opt struct {
	names            []string
	shortnames, help string
	needsArg         bool
	allowsArg        *string            // nil means we don't allow an argument
	process          func(string) error // returns error when it's illegal
}

func addOpt(o opt) {
	newnames := make([]string, 0, len(o.names))
	for _, n := range o.names {
		switch {
		case len(n) < 2:
			panic("Invalid very short flag: " + n)
		case n[0] != '-':
			panic("Invalid flag, doesn't start with '-':" + n)
		case len(n) == 2:
			o.shortnames = o.shortnames + string(n[1])
		case n[1] != '-':
			panic("Invalid long flag, doesn't start with '--':" + n)
		default:
			append(&newnames, n)
		}
	}
	o.names = newnames
	if len(opts) == cap(opts) { // reallocate
		// Allocate double what's needed, for future growth.
		newOpts := make([]opt, len(opts), len(opts)*2)
		for i, oo := range opts {
			newOpts[i] = oo
		}
		opts = newOpts
	}
	opts = opts[0 : 1+len(opts)]
	opts[len(opts)-1] = o
}

// Execute the given closure on the name of all known arguments
func VisitAllNames(f func(string)) {
	for _, o := range opts {
		for _, n := range o.names {
			f(n)
		}
	}
}

// Add a new flag that does not allow arguments
// Parameters:
//   names []string            These are the names that are accepted on the command-line for this flag, e.g. -v --verbose
//   help    string            The help text (automatically Expand()ed) to display for this flag
//   process func() os.Error   The function to call when this flag is processed with no argument
func NoArg(names []string, help string, process func() error) {
	addOpt(opt{names, "", help, false, nil, func(s string) error {
		if s != "" {
			return errors.New("unexpected flag: " + s)
		}
		return process()
	}})
}

// Add a new flag that requires an argument
// Parameters:
//   names []string                  These are the names that are accepted on the command-line for this flag, e.g. -v --verbose
//   argname string                  The name of the argument in help, e.g. the "value" part of "--flag=value"
//   help    string                  The help text (automatically Expand()ed) to display for this flag
//   process func(string) os.Error   The function to call when this flag is processed
func ReqArg(names []string, argname, help string, process func(string) error) {
	addOpt(opt{names, "", help, true, &argname, process})
}

// Add a new flag that may optionally have an argument
// Parameters:
//   names []string                 These are the names that are accepted on the command-line for this flag, e.g. -v --verbose
//   def     string                 The default of the argument in help, e.g. the "value" part of "--flag=value"
//   help    string                 The help text (automatically Expand()ed) to display for this flag
//   process func(string) os.Error  The function to call when this flag is processed with an argument
func OptArg(names []string, def, help string, process func(string) error) {
	addOpt(opt{names, "", help, false, &def, func(s string) error {
		if s == "" {
			return process(def)
		}
		return process(s)
	}})
}

// Create a required-argument flag that only accepts the given set of values
// Parameters:
//   names []string            These are the names that are accepted on the command-line for this flag, e.g. -v --verbose
//   vs    []string            These are the allowable values for the argument
//   help    string            The help text (automatically Expand()ed) to display for this flag
// Returns:
//   *string                   This points to a string whose value is updated as this flag is changed
func Alternatives(names, vs []string, help string) *string {
	possibilities := "[" + vs[0]
	for _, v := range vs[1:] {
		possibilities += "|" + v
	}
	possibilities += "]"
	return AlternativesWithLabel(names, vs, possibilities, help)
}

// Create a required-argument flag that only accepts the given set of valuesand has a Help() label
// Parameters:
//   names []string            These are the names that are accepted on the command-line for this flag, e.g. -v --verbose
//   vs    []string            These are the allowable values for the argument
//   label   string            Label for display in Help()
//   help    string            The help text (automatically Expand()ed) to display for this flag
// Returns:
//   *string                   This points to a string whose value is updated as this flag is changed
func AlternativesWithLabel(names, vs []string, label string, help string) *string {
	out := new(string)
	*out = vs[0]
	f := func(s string) error {
		for _, v := range vs {
			if s == v {
				*out = v
				return nil
			}
		}
		return errors.New("invalid value: " + s)
	}
	ReqArg(names, label, help, f)
	return out
}

// Create a required-argument flag that accepts string values
// Parameters:
//   names []string            These are the names that are accepted on the command-line for this flag, e.g. -v --verbose
//   def     string            Default value for the string and label in Help()
//   help    string            The help text (automatically Expand()ed) to display for this flag
// Returns:
//   *string                   This points to a string whose value is updated as this flag is changed
func String(names []string, def string, help string) *string {
	return StringWithLabel(names, def, def, help)
}

// Create a required-argument flag that accepts string values and has a Help() label
// Parameters:
//   names []string            These are the names that are accepted on the command-line for this flag, e.g. -v --verbose
//   def     string            Default value for the string
//   label   string            Label for display in Help()
//   help    string            The help text (automatically Expand()ed) to display for this flag
// Returns:
//   *string                   This points to a string whose value is updated as this flag is changed
func StringWithLabel(names []string, def string, label string, help string) *string {
	s := new(string)
	*s = def
	f := func(ss string) error {
		*s = ss
		return nil
	}
	ReqArg(names, label, help, f)
	return s
}

// Create a required-argument flag that accepts int values
// Parameters:
//   names []string            These are the names that are accepted on the command-line for this flag, e.g. -v --verbose
//   def     int               Default value for the flag and label in Help()
//   help    string            The help text (automatically Expand()ed) to display for this flag
// Returns:
//   *int                      This points to an int whose value is updated as this flag is changed
func Int(names []string, def int, help string) *int {
	return IntWithLabel(names, def, strconv.Itoa(def), help)
}

// Create a required-argument flag that accepts int values and has a Help() label
// Parameters:
//   names []string            These are the names that are accepted on the command-line for this flag, e.g. -v --verbose
//   def     int               Default value for the flag
//   label   string            Label for display in Help()
//   help    string            The help text (automatically Expand()ed) to display for this flag
// Returns:
//   *int                      This points to an int whose value is updated as this flag is changed
func IntWithLabel(names []string, def int, label string, help string) *int {
	var err error
	i := new(int)
	*i = def
	f := func(istr string) error {
		*i, err = strconv.Atoi(istr)
		return err
	}
	ReqArg(names, label, help, f)
	return i
}

// Create a required-argument flag that accepts string values but allows more than one to be specified
// Parameters:
//   names []string            These are the names that are accepted on the command-line for this flag, e.g. -v --verbose
//   def     string            The argument name of the strings that are appended (e.g. the val in --opt=val)
//   help    string            The help text (automatically Expand()ed) to display for this flag
// Returns:
//   *[]string                 This points to a []string whose value will contain the strings passed as flags
func Strings(names []string, def string, help string) *[]string {
	s := make([]string, 0, 1)
	f := func(ss string) error {
		append(&s, ss)
		return nil
	}
	ReqArg(names, def, help, f)
	return &s
}

// Create a no-argument flag that is set by either passing one of the
// "NO" flags or one of the "YES" flags.  The default value is "false"
// (or "NO").  If you want another default value, you can swap the
// meaning of "NO" and "YES".
//
// Parameters:
//   yes   []string            These flags set the boolean value to true (e.g. -i --install)
//   no    []string            These flags set the boolean value to false (e.g. -I --no-install)
//   helpyes string            The help text (automatically Expand()ed) to display for the "yes" flags
//   helpno  string            The help text (automatically Expand()ed) to display for the "no" flags
// Returns:
//   *bool                     This points to a bool whose value is updated as this flag is changed
func Flag(yes []string, no []string, helpyes, helpno string) *bool {
	b := new(bool)
	y := func() error {
		*b = true
		return nil
	}
	n := func() error {
		*b = false
		return nil
	}
	if len(yes) > 0 {
		NoArg(yes, helpyes, y)
	}
	if len(no) > 0 {
		NoArg(no, helpno, n)
	}
	return b
}

func failnoting(s string, e error) {
	if e != nil {
		fmt.Println(Usage())
		fmt.Println("\n"+s, e.Error())
		os.Exit(1)
	}
}

// This is the list of non-flag arguments after processing
var Args = make([]string, 0, 4)

// This parses the command-line arguments. It returns true if '--' was present.
// Special flags are:
//   --help               Display the generated help message (calls Help())
//   --create-manpage     Display a manpage generated by the goopt library (uses Author, Suite, etc)
//   --list-options       List all known flags
// Arguments:
//   extraopts func() []string     This function is called by --list-options and returns extra options to display
func Parse(extraopts func() []string) bool {
	// First we'll add the "--help" option.
	addOpt(opt{[]string{"--help", "-h"}, "", "Show usage message", false, nil,
		func(string) error {
			fmt.Println(Usage())
			os.Exit(0)
			return nil
		}})
	addOpt(opt{[]string{"--version"}, "", "Show version", false, nil,
		func(string) error {
			fmt.Println(Version)
			os.Exit(0)
			return nil
		}})
	// Let's now tally all the long option names, so we can use this to
	// find "unique" options.
	longnames := []string{"--list-options", "--create-manpage"}
	for _, o := range opts {
		longnames = cat(longnames, o.names)
	}
	// Now let's check if --list-options was given, and if so, list all
	// possible options.
	if any(func(a string) bool { return match(a, longnames) == "--list-options" },
		os.Args[1:]) {
		if extraopts != nil {
			for _, o := range extraopts() {
				fmt.Println(o)
			}
		}
		VisitAllNames(func(n string) { fmt.Println(n) })
		os.Exit(0)
	}
	// Now let's check if --create-manpage was given, and if so, create a
	// man page.
	if any(func(a string) bool { return match(a, longnames) == "--create-manpage" },
		os.Args[0:]) {
		makeManpage()
		os.Exit(0)
	}
	skip := 1
	earlyEnd := false
	for i, a := range os.Args {
		if skip > 0 {
			skip--
			continue
		}
		if a == "--" {
			Args = cat(Args, os.Args[i+1:])
			earlyEnd = true
			break
		}
		if len(a) > 1 && a[0] == '-' && a[1] != '-' {
			for j, s := range a[1:] {
				foundone := false
				for _, o := range opts {
					for _, c := range o.shortnames {
						if c == s {
							switch {
							case o.allowsArg != nil &&
								//	j+1 == len(a)-1 &&
								len(os.Args) > i+skip+1 &&
								len(os.Args[i+skip+1]) >= 1 &&
								(os.Args[i+skip+1] == "-" ||
									os.Args[i+skip+1][0] != '-'):
								// this last one prevents options from taking options as arguments...
								failnoting("Error in flag -"+string(c)+":",
									o.process(os.Args[i+skip+1]))
								skip++ // skip next arg in looking for flags...
							case o.needsArg:
								fmt.Printf("Flag -%c requires argument!\n", c)
								os.Exit(1)
							default:
								failnoting("Error in flag -"+string(c)+":",
									o.process(""))
							}
							foundone = true
							break
						} // Process if we find a match
					} // Loop over the shortnames that this option supports
				} // Loop over the short arguments that we know
				if !foundone {
					badflag := "-" + a[j+1:j+2]
					failnoting("Bad flag:", errors.New(badflag))
				}
			} // Loop over the characters in this short argument
		} else if len(a) > 2 && a[0] == '-' && a[1] == '-' {
			// Looking for a long flag.  Any unique prefix is accepted!
			aflag := match(os.Args[i], longnames)
			foundone := false
			if aflag == "" {
				failnoting("Bad flag:", errors.New(a))
			}
		optloop:
			for _, o := range opts {
				for _, n := range o.names {
					if aflag == n {
						if x := strings.Index(a, "="); x > 0 {
							// We have a --flag=foo argument
							if o.allowsArg == nil {
								fmt.Println("Flag", a, "doesn't want an argument!")
								os.Exit(1)
							}
							failnoting("Error in flag "+a+":",
								o.process(a[x+1:len(a)]))
						} else if o.allowsArg != nil && len(os.Args) > i+1 && len(os.Args[i+1]) >= 1 && (os.Args[i+1] == "-" || os.Args[i+1][0] != '-') {
							// last check sees if the next arg looks like a flag
							failnoting("Error in flag "+n+":",
								o.process(os.Args[i+1]))
							skip++ // skip next arg in looking for flags...
						} else if o.needsArg {
							fmt.Println("Flag", a, "requires argument!")
							os.Exit(1)
						} else { // no (optional) argument was provided...
							failnoting("Error in flag "+n+":", o.process(""))
						}
						foundone = true
						break optloop
					}
				}
			}
			if !foundone {
				failnoting("Bad flag:", errors.New(a))
			}
		} else {
			if RequireOrder {
				Args = cat(Args, os.Args[i:])
				break
			}
			append(&Args, a)
		}
	}

	return earlyEnd
}

func match(x string, allflags []string) string {
	if i := strings.Index(x, "="); i > 0 {
		x = x[0:i]
	}
	for _, f := range allflags {
		if f == x {
			return x
		}
	}
	out := ""
	for _, f := range allflags {
		if len(f) >= len(x) && f[0:len(x)] == x {
			if out == "" {
				out = f
			} else {
				return ""
			}
		}
	}
	return out
}

func makeManpage() {
	_, progname := path.Split(os.Args[0])
	version := Version
	if Suite != "" {
		version = Suite + " " + version
	}
	fmt.Printf(".TH \"%s\" 1 \"%s\" \"%s\" \"%s\"\n", progname,
		time.Now().Format("January 2, 2006"), version, Suite)
	fmt.Println(".SH NAME")
	fmt.Println(progname)
	if Summary != "" {
		fmt.Println("\\-", Summary)
	}
	fmt.Println(".SH SYNOPSIS")
	fmt.Println(progname, Synopsis())
	fmt.Println(".SH DESCRIPTION")
	fmt.Println(formatParagraphs(Description()))
	fmt.Println(".SH OPTIONS")
	for _, o := range opts {
		fmt.Println(".TP")
		switch {
		case len(o.shortnames) == 0:
			for _, n := range o.names[0 : len(o.names)-1] {
				fmt.Printf("\\-\\-%s,", n[2:])
			}
			fmt.Printf("\\-\\-%s", o.names[len(o.names)-1][2:])
			if o.allowsArg != nil {
				fmt.Printf(" %s", *o.allowsArg)
			}
		case len(o.names) == 0:
			for _, c := range o.shortnames[0 : len(o.shortnames)-1] {
				fmt.Printf("\\-%c,", c)
			}
			fmt.Printf("\\-%c", o.shortnames[len(o.shortnames)-1])
			if o.allowsArg != nil {
				fmt.Printf(" %s", *o.allowsArg)
			}
		default:
			for _, c := range o.shortnames {
				fmt.Printf("\\-%c,", c)
			}
			for _, n := range o.names[0 : len(o.names)-1] {
				fmt.Printf("\\-\\-%s,", n[2:])
			}
			fmt.Printf("\\-\\-%s", o.names[len(o.names)-1][2:])
			if o.allowsArg != nil {
				fmt.Printf(" %s", *o.allowsArg)
			}
		}
		fmt.Printf("\n%s\n", Expand(o.help))
	}
	if ExtraUsage != "" {
		fmt.Println("\\-", ExtraUsage)
	}
	if Author != "" {
		fmt.Printf(".SH AUTHOR\n%s\n", Author)
	}
}

func formatParagraphs(x string) string {
	h := new(bytes.Buffer)
	lines := strings.Split(x, "\n")
	for _, l := range lines {
		if l == "" {
			fmt.Fprintln(h, ".PP")
		} else {
			fmt.Fprintln(h, l)
		}
	}
	return h.String()
}
