package system

import (
	"errors"
	"os"
	"path"
	"strconv"
	"strings"
)

type SizedFont struct {
	Name string
	Size float32
}

type RGB struct {
	R, G, B uint8
}

type kdeVars struct {
	GeneralThemeName   string
	GeneralColorScheme string
	GeneralFont        SizedFont

	WMActiveBg   RGB
	WMActiveFg   RGB
	WMInactiveBg RGB
	WMInactiveFg RGB
	WMActiveFont SizedFont

	IconsTheme            string
	KDEColorScheme        string
	KDELookAndFeelPackage string
}

func exists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

func kdeGlobalsPath() (string, bool) {
	xdg := GetXDG()

	filepath := path.Join(xdg.ConfigHome, "kdeglobals")
	if exists(filepath) {
		return filepath, true
	}

	filepath = "/etc/xdg/kdeglobals"
	if exists(filepath) {
		return filepath, true
	}

	return "", false
}

var _kdeVars kdeVars
var _kdeVars_set = false
var _kdeVars_err error = nil

type kdeGlobalsParser struct {
	lines   []string
	line    int
	section string
}

var _eof = string(rune(0))

func (p *kdeGlobalsParser) next() string {
	p.line += 1
	return p.current()
}

func (p *kdeGlobalsParser) current() string {
	if p.line >= len(p.lines) {
		return _eof
	}
	return p.lines[p.line]
}

func (p *kdeGlobalsParser) parseSectionHeader() bool {
	_, after, found := strings.Cut(p.current(), "[")
	if !found {
		return false
	}
	section, _, found := strings.Cut(after, "]")
	if !found {
		return false
	}
	p.next()
	p.section = section
	return true
}

func (p *kdeGlobalsParser) parseVariable() (string, string, bool) {
	before, after, found := strings.Cut(p.current(), "=")
	if !found {
		return "", "", false
	}
	p.next()
	return strings.TrimSpace(before), strings.TrimSpace(after), true
}

func parseSizedFont(val string) (SizedFont, bool) {
	fontprops := strings.Split(val, ",")
	if len(fontprops) < 2 {
		return SizedFont{}, false
	}
	name := fontprops[0]
	size, err := strconv.ParseFloat(fontprops[1], 32)
	if err != nil {
		return SizedFont{}, false
	}
	return SizedFont{name, float32(size)}, true
}

func parseRGB(val string) (RGB, bool) {
	rgbprops := strings.Split(val, ",")
	if len(rgbprops) < 3 {
		return RGB{}, false
	}
	r, err := strconv.Atoi(rgbprops[0])
	if err != nil {
		return RGB{}, false
	}
	g, err := strconv.Atoi(rgbprops[1])
	if err != nil {
		return RGB{}, false
	}
	b, err := strconv.Atoi(rgbprops[2])
	if err != nil {
		return RGB{}, false
	}

	return RGB{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
	}, true
}

func parseKDEVars(kdeglobals string) (kdeVars, bool) {
	p := kdeGlobalsParser{lines: strings.Split(kdeglobals, "\n")}

	var out kdeVars
	for p.current() != _eof {
		if p.parseSectionHeader() {
			continue
		}
		if p.section != "KDE" && p.section != "WM" && p.section != "Icons" && p.section != "General" {
			p.next()
			continue
		}

		name, val, ok := p.parseVariable()
		if !ok {
			p.next()
			continue
		}

		switch p.section {
		case "KDE":
			switch name {
			case "LookAndFeelPackage":
				out.KDELookAndFeelPackage = val

			case "ColorScheme":
				out.KDEColorScheme = val
			}

		case "WM":
			switch name {
			case "ActiveFont":
				font, ok := parseSizedFont(val)
				if !ok {
					return kdeVars{}, false
				}
				out.WMActiveFont = font

			case "ActiveForeground":
				rgb, ok := parseRGB(val)
				if !ok {
					return kdeVars{}, false
				}
				out.WMActiveFg = rgb

			case "ActiveBackground":
				rgb, ok := parseRGB(val)
				if !ok {
					return kdeVars{}, false
				}
				out.WMActiveBg = rgb

			case "InactiveForeground":
				rgb, ok := parseRGB(val)
				if !ok {
					return kdeVars{}, false
				}
				out.WMInactiveFg = rgb

			case "InactiveBackground":
				rgb, ok := parseRGB(val)
				if !ok {
					return kdeVars{}, false
				}
				out.WMInactiveBg = rgb
			}

		case "Icons":
			if name != "Theme" {
				continue
			}
			out.IconsTheme = val

		case "General":
			switch name {
			case "Font":
				font, ok := parseSizedFont(val)
				if !ok {
					return kdeVars{}, false
				}
				out.GeneralFont = font

			case "ColorScheme":
				out.GeneralColorScheme = val

			case "GeneralThemeName":
				out.GeneralThemeName = val
			}
		}
	}

	return out, true
}

func getKDEVars() (kdeVars, error) {
	if !_kdeVars_set {
		_kdeVars_set = true

		filepath, ok := kdeGlobalsPath()
		if !ok {
			_kdeVars_err = errors.New("Failed to find kdeglobals file!")
			return kdeVars{}, _kdeVars_err
		}

		text, err := os.ReadFile(filepath)

		if err != nil {
			_kdeVars_err = err
			return kdeVars{}, err
		}

		_kdeVars, ok = parseKDEVars(string(text))
		if !ok {
			_kdeVars_err = errors.New("Failed to parse kdeVars")
			return kdeVars{}, _kdeVars_err
		}
	}

	if _kdeVars_err != nil {
		return kdeVars{}, _kdeVars_err
	}

	return _kdeVars, nil
}
