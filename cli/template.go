// template.go

package main

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/steam/utils/appid"
)

/*
TemplateData represents the data passed to templates for rendering.

It includes server information, extra data, and server connection details.
*/
type TemplateData struct {
	Info  *a2s.Info // Server information from A2S
	Extra any       // Additional arbitrary data
	ID    string    // Server identifier
	Host  string    // Server host address
	Port  int       // Server port
}

/*
render applies the template string to the TemplateData.

It parses the provided template string, applies the data,
and returns the resulting string or an error if the process fails.
*/
func (t *TemplateData) render(tplStr string) (string, error) {
	funcMap := template.FuncMap{
		"AppID":           tplHelperAppIDtoString,
		"DurationEmoji":   tplHelperDurationEmoji,
		"TimeEmoji":       tplHelperTimeEmoji,
		"OSEmoji":         tplHelperOSEmoji,
		"CountryEmoji":    tplHelperCountryEmoji,
		"CodeEmoji":       tplHelperCodeEmoji,
		"ValueColorEmoji": tplHelperValueColorEmoji,
	}

	tmpl, err := template.New("template").Funcs(funcMap).Parse(tplStr)
	if err != nil {
		return "⛔ template error", err
	}

	var sb strings.Builder
	err = tmpl.Execute(&sb, t)
	if err != nil {
		return "⚠️ template error", err
	}

	return sb.String(), nil
}

/*
tplHelperDurationEmoji returns an emoji based on the given time duration.

It categorizes the time into different periods of the day:
  - 🌙 for night (0-7 hours or after 20 hours)
  - 🌞 for day (7-20 hours)
*/
func tplHelperDurationEmoji(t time.Duration) string {
	hours := math.Mod(t.Hours(), 24)
	if hours < 7 || hours > 20 {
		return "🌙"
	}
	return "🌞"
}

/*
tplHelperTimeEmoji returns an emoji based on the given time.

It categorizes the time into different periods of the day:
  - 🌙 for night (0-7 hours or after 20 hours)
  - 🌞 for day (7-20 hours)
*/
func tplHelperTimeEmoji(t time.Time) string {
	if t.Hour() < 7 || t.Hour() > 20 {
		return "🌙"
	}
	return "🌞"
}

// tplHelperOSEmoji converts an OS name to emoji representation.
func tplHelperOSEmoji(name string) string {
	switch strings.ToLower(name) {
	case "a", "m", "apple", "mac", "osx", "ios":
		return "🍎"
	case "l", "nix", "linux", "tux":
		return "🐧"
	case "w", "win", "windows", "nt":
		return "🪟"
	default:
		return "😈" // FreeBSD
	}
}

/*
AppIDtoString converts an AppID to its string representation.

It takes a uint64 AppID and returns its corresponding string.
*/
func tplHelperAppIDtoString(id uint64) string {
	return appid.AppID(id).String()
}

/*
tplHelperCountryEmoji converts the name of the country (long or short) in Emoji flag.

He uses the card only for exceptions, the rest are processed automatically.
*/
func tplHelperCountryEmoji(name string) string {
	exceptionCountryCodes := map[string]string{
		"united kingdom":       "GB",
		"usa":                  "US",
		"united states":        "US",
		"south korea":          "KR",
		"south africa":         "ZA",
		"new zealand":          "NZ",
		"north korea":          "KP",
		"east timor":           "TL",
		"united arab emirates": "AE",
	}

	lowerName := strings.ToLower(strings.TrimSpace(name))

	if code, exists := exceptionCountryCodes[lowerName]; exists {
		return tplHelperCodeEmoji(code)
	}

	if len(lowerName) >= 2 {
		var letters []rune
		for _, r := range lowerName {
			if unicode.IsLetter(r) {
				letters = append(letters, unicode.ToUpper(r))
				if len(letters) == 2 {
					break
				}
			}
		}

		if len(letters) == 2 {
			return tplHelperCodeEmoji(string(letters))
		}
	}

	return "🏳️"
}

// tplHelperCodeEmoji transforms the two-letter code of the country into emoji flag.
func tplHelperCodeEmoji(code string) string {
	if len(code) != 2 {
		return "🏳️"
	}

	runes := make([]rune, 2)
	for i, char := range code {
		if char < 'A' || char > 'Z' {
			return "🏳️"
		}
		runes[i] = 0x1F1E6 + (char - 'A')
	}
	return string(runes)
}

// tplHelperValueToColorEmoji returns color emoji based on the current meaning and maximum.
// 🟣 — 0
// 🔵 — <10%
// 🟢 — <50%
// 🟡 — <75%
// 🟠 — <90%
// 🔴 — 100%
// 🚫 — the value is less than 0 or exceeds the maximum
func tplHelperValueColorEmoji(from, to any) string {
	value := toInt64(from)
	limit := toInt64(to)

	if value == 0 {
		return "🟣"
	}

	if value < 0 || value > limit {
		return "🚫"
	}

	proportion := float64(value) / float64(limit)
	switch {
	case proportion <= 0.1:
		return "🔵" // Low
	case proportion <= 0.5:
		return "🟢" // Normal
	case proportion <= 0.75:
		return "🟡" // Mid
	case proportion <= 0.9:
		return "🟠" // High
	default:
		return "🔴" // Crit
	}
}

// force parse numbers and strings to int or return 0 otherwise
func toInt64(v any) int64 {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Int64, reflect.Int:
		return val.Int()
	case reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return int64(val.Uint()) // #nosec G115
	case reflect.Uint64:
		return int64(val.Uint()) // #nosec G115 // TODO not safe
	case reflect.Float32, reflect.Float64:
		return int64(val.Float()) // #nosec G115
	case reflect.String:
		i, err := strconv.ParseInt(val.String(), 10, 64)
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}
