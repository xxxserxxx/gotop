package lingo

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// Lingo is a translation bundle of all translations by locale, as well as
// default locale and list of supported locales
type Lingo struct {
	bundle    map[string]Translations
	deflt     string
	supported []Locale
}

// New creates the Lingo bundle. `deflt` is the default locale, which is used
// when the requested locale is not found, and when translations are not found
// in the requested locale. `path` is the absolute or relative path to the
// folder of TOML translations. `fs` is either a FileSystem or null, used to
// locate the path and translation files.
func New(deflt, path string, root fs.FS) (*Lingo, error) {
	if root == nil {
		root = os.DirFS(".")
	}
	l := &Lingo{
		bundle:    make(map[string]Translations),
		deflt:     deflt,
		supported: make([]Locale, 0),
	}
	err := fs.WalkDir(root, path, func(pth string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() { // We skip these
			return nil
		}
		fileName := info.Name()
		if !strings.HasSuffix(fileName, ".toml") {
			return nil
		}
		f, err := root.Open(pth)
		if err != nil {
			return err
		}
		dat, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		t := Translations{
			transl: make(map[string]interface{}),
		}
		err = toml.Unmarshal(dat, &t.transl)
		if err != nil {
			return fmt.Errorf("in file %s: %w", fileName, err)
		}
		locale := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		if locale != l.deflt {
			t.def = l
		}
		l.supported = append(l.supported, ParseLocale(locale))
		l.bundle[locale] = t
		return nil
	})
	return l, err
}

// TranslationsForRequest will get the best matched T for given
// Request. If no T is found, returns default T
func (l *Lingo) TranslationsForRequest(r *http.Request) Translations {
	locales := GetLocales(r)
	for _, locale := range locales {
		t, exists := l.bundle[locales[0].Name()]
		if exists {
			return t
		}
		for _, sup := range l.supported {
			if locale.Lang == sup.Lang {
				return l.bundle[sup.Name()]
			}
		}
	}
	return l.bundle[l.deflt]
}

// TranslationsForLocale will get the T for specific locale.
// If no locale is found, returns default T
func (l *Lingo) TranslationsForLocale(locale string) Translations {
	if t, exists := l.bundle[locale]; exists {
		return t
	}
	return l.bundle[l.deflt]
}

// Translations represents translations map for specific locale
type Translations struct {
	def    *Lingo
	transl map[string]interface{}
}

// Value traverses the translations map and finds translation for
// given key. If no translation is found, returns value of given key.
func (t Translations) Value(key string, args ...string) string {
	if v, ok := t.transl[key]; ok {
		if s, ok := v.(string); ok {
			return t.sprintf(s, args)
		}
	}
	ss := strings.Split(key, ".")
	cm := t.transl
	for _, k := range ss {
		if m, ok := cm[k].(map[string]interface{}); ok {
			cm = m
			continue
		}
		if v, ok := cm[k]; ok {
			if s, ok := v.(string); ok {
				return t.sprintf(s, args)
			}
			break
		}
	}
	if t.def != nil {
		return t.def.TranslationsForLocale(t.def.deflt).Value(key, args...)
	}
	return key
}

// sprintf replaces the argument placeholders with given arguments
func (t Translations) sprintf(value string, args []string) string {
	res := value
	for i := 0; i < len(args); i++ {
		tok := "{" + strconv.Itoa(i) + "}"
		res = strings.Replace(res, tok, args[i], -1)
	}
	return res
}
