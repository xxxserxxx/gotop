package lingo

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Locale is locale value from the Accept-Language header in request
type Locale struct {
	Lang, Country string
	Qual          float64
}

// Name returns the locale value in 'lang' or 'lang_country' format
// eg: de_DE, en_US, gb
func (l *Locale) Name() string {
	if len(l.Country) > 0 {
		return l.Lang + "_" + l.Country
	}
	return l.Lang
}

// ParseLocale creates a Locale from a locale string
func ParseLocale(locale string) Locale {
	locsplt := strings.Split(locale, "_")
	resp := Locale{}
	resp.Lang = locsplt[0]
	if len(locsplt) > 1 {
		resp.Country = locsplt[1]
	}
	return resp
}

const (
	acceptLanguage = "Accept-Language"
)

func supportedLocales(alstr string) []Locale {
	locales := make([]Locale, 0)
	alstr = strings.Replace(alstr, " ", "", -1)
	if alstr == "" {
		return locales
	}
	al := strings.Split(alstr, ",")
	for _, lstr := range al {
		locales = append(locales, Locale{
			Lang:    parseLang(lstr),
			Country: parseCountry(lstr),
			Qual:    parseQual(lstr),
		})
	}
	return locales
}

// GetLocales returns supported locales for the given requet
func GetLocales(r *http.Request) []Locale {
	return supportedLocales(r.Header.Get(acceptLanguage))
}

// GetPreferredLocale return preferred locale for the given reuqest
// returns error if there is no preferred locale
func GetPreferredLocale(r *http.Request) (*Locale, error) {
	locales := GetLocales(r)
	if len(locales) == 0 {
		return &Locale{}, errors.New("No locale found")
	}
	return &locales[0], nil
}

func parseLang(val string) string {
	locale := strings.Split(val, ";")[0]
	lang := strings.Split(locale, "-")[0]
	return lang
}

func parseCountry(val string) string {
	locale := strings.Split(val, ";")[0]
	spl := strings.Split(locale, "-")
	if len(spl) > 1 {
		return spl[1]
	}
	return ""
}

func parseQual(val string) float64 {
	spl := strings.Split(val, ";")
	if len(spl) > 1 {
		qual, err := strconv.ParseFloat(strings.Split(spl[1], "=")[1], 64)
		if err != nil {
			return 1
		}
		return qual
	}
	return 1
}
