package theme

import (
	"fmt"
	"log"

	"github.com/chewxy/math32"
)

// The theming code is a port of the Lagrange theming code.
type Color struct {
	R, G, B uint8
}

func (rgb Color) String() string {
	return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
}

type Theme struct {
	colors [maxColorId]Color
}

func (t *Theme) Background() Color            { return t.colors[int(tmBackground)] }
func (t *Theme) Paragraph() Color             { return t.colors[int(tmParagraph)] }
func (t *Theme) FirstParagraph() Color        { return t.colors[int(tmFirstParagraph)] }
func (t *Theme) Quote() Color                 { return t.colors[int(tmQuote)] }
func (t *Theme) QuoteIcon() Color             { return t.colors[int(tmQuoteIcon)] }
func (t *Theme) Preformatted() Color          { return t.colors[int(tmPreformatted)] }
func (t *Theme) Heading1() Color              { return t.colors[int(tmHeading1)] }
func (t *Theme) Heading2() Color              { return t.colors[int(tmHeading2)] }
func (t *Theme) Heading3() Color              { return t.colors[int(tmHeading3)] }
func (t *Theme) BannerBackground() Color      { return t.colors[int(tmBannerBackground)] }
func (t *Theme) BannerTitle() Color           { return t.colors[int(tmBannerTitle)] }
func (t *Theme) BannerIcon() Color            { return t.colors[int(tmBannerIcon)] }
func (t *Theme) BannerSideTitle() Color       { return t.colors[int(tmBannerSideTitle)] }
func (t *Theme) OutlineHeadingAbove() Color   { return t.colors[int(tmOutlineHeadingAbove)] }
func (t *Theme) OutlineHeadingBelow() Color   { return t.colors[int(tmOutlineHeadingBelow)] }
func (t *Theme) InlineContentMetadata() Color { return t.colors[int(tmInlineContentMetadata)] }

func (t *Theme) LinkIcon() Color          { return t.colors[int(tmLinkIcon)] }
func (t *Theme) LinkIconVisited() Color   { return t.colors[int(tmLinkIconVisited)] }
func (t *Theme) LinkText() Color          { return t.colors[int(tmLinkText)] }
func (t *Theme) LinkTextHover() Color     { return t.colors[int(tmLinkTextHover)] }
func (t *Theme) LinkDomain() Color        { return t.colors[int(tmLinkDomain)] }
func (t *Theme) LinkLastVisitDate() Color { return t.colors[int(tmLinkLastVisitDate)] }

func (t *Theme) HypertextLinkIcon() Color        { return t.colors[int(tmHypertextLinkIcon)] }
func (t *Theme) HypertextLinkIconVisited() Color { return t.colors[int(tmHypertextLinkIconVisited)] }
func (t *Theme) HypertextLinkText() Color        { return t.colors[int(tmHypertextLinkText)] }
func (t *Theme) HypertextLinkTextHover() Color   { return t.colors[int(tmHypertextLinkTextHover)] }
func (t *Theme) HypertextLinkDomain() Color      { return t.colors[int(tmHypertextLinkDomain)] }
func (t *Theme) HypertextLinkLastVisitDate() Color {
	return t.colors[int(tmHypertextLinkLastVisitDate)]
}

func (t *Theme) GopherLinkIcon() Color          { return t.colors[int(tmGopherLinkIcon)] }
func (t *Theme) GopherLinkIconVisited() Color   { return t.colors[int(tmGopherLinkIconVisited)] }
func (t *Theme) GopherLinkText() Color          { return t.colors[int(tmGopherLinkText)] }
func (t *Theme) GopherLinkTextHover() Color     { return t.colors[int(tmGopherLinkTextHover)] }
func (t *Theme) GopherLinkDomain() Color        { return t.colors[int(tmGopherLinkDomain)] }
func (t *Theme) GopherLinkLastVisitDate() Color { return t.colors[int(tmGopherLinkLastVisitDate)] }

func getSeed(name string) uint32 {
	if name == "" {
		return 0
	}
	return lgrCrc32([]byte(name))
}

func NewWhiteTheme(name string) *Theme {
	t := &Theme{}
	t.applyLightLinkTheme(&lightPalette)
	t.applyWhiteTheme(&lightPalette)
	t.applyWhiteSeed(&lightPalette, getSeed(name))
	log.Printf("Seed for %s was %x", name, getSeed(name))
	return t
}

func NewColorfulDarkTheme(name string) *Theme {
	t := &Theme{}
	t.applyDarkLinkTheme(&darkPalette)
	t.applyColorfulDarkTheme(&darkPalette)
	t.applyColorfulDarkSeed(&darkPalette, getSeed(name))
	return t
}

type hslColor struct {
	Hue, Sat, Lum float32
}

func (color Color) toHSL() hslColor {
	rgb := [3]float32{
		float32(color.R) / 255,
		float32(color.G) / 255,
		float32(color.B) / 255,
	}
	var compMax, compMin int
	if rgb[0] >= rgb[1] && rgb[0] >= rgb[2] {
		compMax = 0
	} else if rgb[1] >= rgb[0] && rgb[1] >= rgb[2] {
		compMax = 1
	} else {
		compMax = 2
	}
	if rgb[0] <= rgb[1] && rgb[0] <= rgb[2] {
		compMin = 0
	} else if rgb[1] <= rgb[0] && rgb[1] <= rgb[2] {
		compMax = 1
	} else {
		compMin = 2
	}
	rgbMax := rgb[compMax]
	rgbMin := rgb[compMin]
	lum := (rgbMax + rgbMin) / 2.0
	var hue float32 = 0.0
	var sat float32 = 0.0
	if math32.Abs(rgbMax-rgbMin) > 0.00001 {
		chr := rgbMax - rgbMin
		sat = chr / (1.0 - math32.Abs(2.0*lum-1.0))
		switch compMax {
		case 0:
			hue = (rgb[1] - rgb[2]) / chr
			if rgb[1] < rgb[2] {
				hue += 6
			}
		case 1:
			hue = (rgb[2]-rgb[0])/chr + 2
		case 2:
			hue = (rgb[0]-rgb[1])/chr + 4
		}
	}
	return hslColor{hue * 60, sat, lum}
}

func wrapf01(v float32) float32 {
	return v - math32.Floor(v)
}

func clampf01(v float32) float32 {
	if v < 0 {
		return 0
	} else if v > 1 {
		return 1
	} else {
		return v
	}
}

func hue2rgb(p, q, t float32) float32 {
	if t < 0.0 {
		t += 1.0
	}
	if t > 1.0 {
		t -= 1.0
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6.0*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6.0
	}
	return p
}

func (hsl hslColor) toRGB() Color {
	var r, g, b float32
	hsl.Hue /= 360.0
	hsl.Hue = wrapf01(hsl.Hue)
	hsl.Sat = clampf01(hsl.Sat)
	hsl.Lum = clampf01(hsl.Lum)
	if hsl.Sat < 0.00001 {
		r = hsl.Lum
		g = hsl.Lum
		b = hsl.Lum
	} else {
		var q float32
		if hsl.Lum < 0.5 {
			q = hsl.Lum * (1 + hsl.Sat)
		} else {
			q = (hsl.Lum + hsl.Sat - hsl.Lum*hsl.Sat)
		}
		p := 2*hsl.Lum - q
		r = hue2rgb(p, q, hsl.Hue+1.0/3.0)
		g = hue2rgb(p, q, hsl.Hue)
		b = hue2rgb(p, q, hsl.Hue-1.0/3.0)
	}
	return Color{uint8(r * 255), uint8(g * 255), uint8(b * 255)}
}

func (hsl hslColor) addSatLum(sat, lum float32) hslColor {
	hsl.Sat = clampf01(hsl.Sat + sat)
	hsl.Lum = clampf01(hsl.Lum + lum)
	return hsl
}

func (hsl hslColor) setSat(sat float32) hslColor {
	hsl.Sat = clampf01(sat)
	return hsl
}

func (hsl hslColor) setLum(lum float32) hslColor {
	hsl.Lum = clampf01(lum)
	return hsl
}

func iabs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (c1 Color) delta(c2 Color) int {
	return iabs(int(c1.R)-int(c2.R)) +
		iabs(int(c1.G)-int(c2.G)) +
		iabs(int(c1.B)-int(c2.B))
}

func (c1 Color) mix(c2 Color, t float32) Color {
	t = clampf01(t)
	return Color{
		uint8(float32(c1.R)*(1-t) + float32(c2.R)*t),
		uint8(float32(c1.G)*(1-t) + float32(c2.G)*t),
		uint8(float32(c1.B)*(1-t) + float32(c2.B)*t),
	}
}

type palette struct {
	Black   Color
	Gray25  Color
	Gray50  Color
	Gray75  Color
	White   Color
	Brown   Color
	Orange  Color
	Teal    Color
	Cyan    Color
	Yellow  Color
	Red     Color
	Magenta Color
	Blue    Color
	Green   Color
}

var darkPalette = palette{
	Black:   Color{0, 0, 0},
	Gray25:  Color{40, 40, 40},
	Gray50:  Color{80, 80, 80},
	Gray75:  Color{160, 160, 160},
	White:   Color{255, 255, 255},
	Brown:   Color{106, 80, 0},
	Orange:  Color{255, 192, 0},
	Teal:    Color{0, 96, 128},
	Cyan:    Color{0, 192, 255},
	Yellow:  Color{255, 255, 32},
	Red:     Color{255, 64, 64},
	Magenta: Color{255, 0, 255},
	Blue:    Color{132, 132, 255},
	Green:   Color{0, 200, 0},
}

var lightPalette = palette{
	Black:   Color{0, 0, 0},
	Gray25:  Color{75, 75, 75},
	Gray50:  Color{150, 150, 150},
	Gray75:  Color{235, 235, 235},
	White:   Color{255, 255, 255},
	Brown:   Color{142, 100, 20},
	Orange:  Color{215, 210, 200},
	Teal:    Color{10, 85, 112},
	Cyan:    Color{150, 205, 220},
	Yellow:  Color{255, 255, 32},
	Red:     Color{255, 64, 64},
	Magenta: Color{255, 0, 255},
	Blue:    Color{132, 132, 255},
	Green:   Color{0, 150, 0},
}

type colorId int

const (
	tmBackground colorId = iota
	tmParagraph
	tmFirstParagraph
	tmQuote
	tmQuoteIcon
	tmPreformatted
	tmHeading1
	tmHeading2
	tmHeading3
	tmBannerBackground
	tmBannerTitle
	tmBannerIcon
	tmBannerSideTitle
	tmOutlineHeadingAbove
	tmOutlineHeadingBelow
	tmInlineContentMetadata
	tmLinkIcon
	tmLinkIconVisited
	tmLinkText
	tmLinkTextHover
	tmLinkDomain
	tmLinkLastVisitDate
	tmHypertextLinkIcon
	tmHypertextLinkIconVisited
	tmHypertextLinkText
	tmHypertextLinkTextHover
	tmHypertextLinkDomain
	tmHypertextLinkLastVisitDate
	tmGopherLinkIcon
	tmGopherLinkIconVisited
	tmGopherLinkText
	tmGopherLinkTextHover
	tmGopherLinkDomain
	tmGopherLinkLastVisitDate
	maxColorId
)

func (id colorId) isLink() bool {
	return id >= tmLinkText
}

func (id colorId) isBackground() bool {
	return id == tmBackground || id == tmBannerBackground
}

func (id colorId) isText() bool {
	return !id.isBackground()
}

func (id colorId) isLinkText() bool {
	return id == tmLinkText ||
		id == tmHypertextLinkText ||
		id == tmGopherLinkText
}

func (id colorId) isRegularText() bool {
	return id.isLinkText() || id == tmParagraph || id == tmFirstParagraph
}

func (t *Theme) applyDarkLinkTheme(p *palette) {
	t.colors[tmInlineContentMetadata] = p.Cyan
	t.colors[tmLinkText] = p.White
	t.colors[tmLinkIcon] = p.Cyan
	t.colors[tmLinkTextHover] = p.Cyan
	t.colors[tmLinkIconVisited] = p.Teal
	t.colors[tmLinkDomain] = p.Teal
	t.colors[tmLinkLastVisitDate] = p.Cyan
	t.colors[tmHypertextLinkText] = p.White
	t.colors[tmHypertextLinkIcon] = p.Orange
	t.colors[tmHypertextLinkTextHover] = p.Orange
	t.colors[tmHypertextLinkIconVisited] = p.Brown
	t.colors[tmHypertextLinkDomain] = p.Brown
	t.colors[tmHypertextLinkLastVisitDate] = p.Orange
	t.colors[tmGopherLinkText] = p.White
	t.colors[tmGopherLinkIcon] = p.Magenta
	t.colors[tmGopherLinkTextHover] = p.Blue
	t.colors[tmGopherLinkIconVisited] = p.Blue
	t.colors[tmGopherLinkDomain] = p.Magenta
	t.colors[tmGopherLinkLastVisitDate] = p.Blue
}

func (t *Theme) applyLightLinkTheme(p *palette) {
	t.colors[tmInlineContentMetadata] = p.Brown
	t.colors[tmLinkText] = p.Black
	t.colors[tmLinkIcon] = p.Teal
	t.colors[tmLinkTextHover] = p.Teal
	t.colors[tmLinkIconVisited] = p.Cyan
	t.colors[tmLinkDomain] = p.Cyan
	t.colors[tmLinkLastVisitDate] = p.Teal
	t.colors[tmHypertextLinkText] = p.Black
	t.colors[tmHypertextLinkIcon] = p.Brown
	t.colors[tmHypertextLinkTextHover] = p.Brown
	t.colors[tmHypertextLinkIconVisited] = p.Orange
	t.colors[tmHypertextLinkDomain] = p.Orange
	t.colors[tmHypertextLinkLastVisitDate] = p.Brown
	t.colors[tmGopherLinkText] = p.Black
	t.colors[tmGopherLinkIcon] = p.Magenta
	t.colors[tmGopherLinkTextHover] = p.Blue
	t.colors[tmGopherLinkIconVisited] = p.Blue
	t.colors[tmGopherLinkDomain] = p.Magenta
	t.colors[tmGopherLinkLastVisitDate] = p.Blue
}

func (t *Theme) applyColorfulDarkTheme(p *palette) {
	base := hslColor{200, 0, 0.15}
	t.colors[tmBackground] = base.toRGB()
	t.colors[tmParagraph] = p.Gray75
	t.colors[tmFirstParagraph] = base.addSatLum(0, 0.75).toRGB()
	t.colors[tmQuote] = p.Cyan
	t.colors[tmPreformatted] = p.Cyan
	t.colors[tmHeading1] = p.White
	t.colors[tmHeading2] = base.addSatLum(0.5, 0.5).toRGB()
	t.colors[tmHeading3] = base.addSatLum(1.0, 0.4).toRGB()
	t.colors[tmBannerBackground] = base.addSatLum(0, -0.05).toRGB()
	t.colors[tmBannerTitle] = p.White
	t.colors[tmBannerIcon] = p.Orange
}

func (t *Theme) applyWhiteTheme(p *palette) {
	base := hslColor{40, 0, 1.0}
	t.colors[tmBackground] = base.toRGB()
	t.colors[tmParagraph] = p.Gray25
	t.colors[tmFirstParagraph] = p.Black
	t.colors[tmQuote] = p.Brown
	t.colors[tmPreformatted] = p.Brown
	t.colors[tmHeading1] = p.Black
	t.colors[tmHeading2] = base.addSatLum(0.15, -0.7).toRGB()
	t.colors[tmHeading3] = base.addSatLum(0.3, -0.6).toRGB()
	t.colors[tmBannerBackground] = p.White
	t.colors[tmBannerTitle] = p.Gray50
	t.colors[tmBannerIcon] = p.Teal
}

const (
	redHue int = iota
	reddishOrangeHue
	yellowishOrangeHue
	yellowHue
	greenishYellowHue
	greenHue
	bluishGreenHue
	cyanHue
	skyBlueHue
	blueHue
	violetHue
	pinkHue
)

var hues = [12]float32{
	5, 25, 40, 56, 80, 120, 160, 180, 208, 231, 270, 324,
}

var altHues = [12]struct{ index [2]int }{
	{[2]int{2, 4}},  // red
	{[2]int{8, 3}},  // reddish orange
	{[2]int{7, 9}},  // yellowish orange
	{[2]int{5, 7}},  // yellow
	{[2]int{11, 2}}, // greenish yellow
	{[2]int{1, 3}},  // green
	{[2]int{2, 4}},  // bluish green
	{[2]int{2, 11}}, // cyan
	{[2]int{6, 10}}, // sky blue
	{[2]int{3, 11}}, // blue
	{[2]int{8, 9}},  // violet
	{[2]int{7, 8}},  // pink
}

var normLum = [12]float32{
	0.8, 0.7, 0.675, 0.65, 0.55, 0.6,
	0.475, 0.475, 0.75, 0.8, 0.85, 0.85,
}

func (t *Theme) applyColorfulDarkSeed(p *palette, seed uint32) {
	primIndex := 2
	if seed != 0 {
		primIndex = int(seed&0xff) % len(hues)
	}
	altIndex := [2]uint32{
		(seed & 0x4) >> 2,
		(seed & 0x40) >> 6,
	}
	altHue := hues[8]
	altHue2 := hues[8]
	if seed != 0 {
		altHue = hues[altHues[primIndex].index[altIndex[0]]]
		altHue2 = hues[altHues[primIndex].index[altIndex[1]]]
	}
	isBannerLighter := (seed & 0x4000) != 0
	isDarkBgSat := (seed&0x200000) != 0 && (primIndex < 1 || primIndex > 4)

	// Begin theme-specific code
	{
		base := hslColor{
			hues[primIndex],
			0.8 * float32(seed>>24) / 255.0,
			0.06 + 0.09*float32((seed>>5)&0x7)/7.0,
		}
		altBase := hslColor{altHue, base.Sat, base.Lum}

		t.colors[tmBackground] = base.toRGB()

		if isBannerLighter {
			t.colors[tmBannerBackground] = base.addSatLum(0.1, 0.04).toRGB()
		} else {
			t.colors[tmBannerBackground] = base.addSatLum(0.1, -0.04).toRGB()
		}
		t.colors[tmBannerTitle] = base.addSatLum(0.1, 0).setLum(0.55).toRGB()
		t.colors[tmBannerIcon] = base.addSatLum(0.35, 0).setLum(0.65).toRGB()

		titleLum := 0.2 * float32((seed>>17)&0x7) / 7.0
		t.colors[tmHeading1] = altBase.setLum(titleLum + 0.80).toRGB()
		t.colors[tmHeading2] = altBase.setLum(titleLum + 0.70).toRGB()
		t.colors[tmHeading3] = altBase.setLum(titleLum + 0.60).toRGB()

		t.colors[tmParagraph] = base.addSatLum(0.1, 0.6).toRGB()

		if t.colors[tmHeading3].delta(t.colors[tmParagraph]) <= 80 {
			// Smallest heading may be too close to body text color
			t.colors[tmHeading3] = t.colors[tmHeading3].toHSL().addSatLum(0, 0.15).toRGB()
		}

		t.colors[tmFirstParagraph] = base.addSatLum(0.2, 0.72).toRGB()
		t.colors[tmPreformatted] = hslColor{altHue2, 1.0, 0.75}.toRGB()
		t.colors[tmQuote] = t.colors[tmPreformatted]
		t.colors[tmInlineContentMetadata] = t.colors[tmHeading3]
	}
	// end theme-specific code

	// Begin dark-mode only code
	for i := colorId(0); i < maxColorId; i++ {
		color := t.colors[i].toHSL()
		if !i.isLink() {
			if isDarkBgSat {
				// Saturate background, desaturate text
				if i.isBackground() {
					if primIndex != greenHue {
						color.Sat = (color.Sat + 1) / 2
					} else {
						color.Sat *= 0.5
					}
					color.Lum *= 0.75
				} else if i.isText() {
					color.Lum = (color.Lum + 1) / 2
				}
			} else {
				// Desaturate backgroud, saturate text
				if i.isBackground() {
					color.Sat *= 0.333
				} else if i.isText() {
					color.Sat = (color.Sat + 2) / 3
					color.Lum = (2*color.Lum + 1) / 3
				}
			}
		}
		t.colors[i] = color.toRGB()
	}
	// end dark-mode only code

	t.colors[tmQuoteIcon] = t.colors[tmQuote].
		mix(t.colors[tmBackground], 0.55)
	t.colors[tmOutlineHeadingAbove] = p.White
	t.colors[tmOutlineHeadingBelow] = p.Black

	// begin theme-specific code
	t.colors[tmBannerSideTitle] = t.colors[tmBannerTitle].
		mix(t.colors[tmBackground], 0.55)
	t.colors[tmOutlineHeadingBelow] = t.colors[tmBannerTitle]
	if t.colors[tmOutlineHeadingAbove] == t.colors[tmOutlineHeadingBelow] {
		t.colors[tmOutlineHeadingBelow] = t.colors[tmHeading3]
	}
	// end theme-specific code
}

func (t *Theme) applyWhiteSeed(p *palette, seed uint32) {
	primIndex := 2
	if seed != 0 {
		primIndex = int(seed&0xff) % len(hues)
	}
	altIndex := [2]uint32{
		(seed & 0x4) >> 2,
		(seed & 0x40) >> 6,
	}
	altHue := hues[8]
	altHue2 := hues[8]
	if seed != 0 {
		altHue = hues[altHues[primIndex].index[altIndex[0]]]
		altHue2 = hues[altHues[primIndex].index[altIndex[1]]]
	}

	// Begin theme-specific code
	{
		base := hslColor{hues[primIndex], 1.0, 0.3}
		altBase := hslColor{altHue, base.Sat, base.Lum - 0.1}

		t.colors[tmBackground] = p.White
		t.colors[tmBannerBackground] = p.White
		t.colors[tmBannerTitle] = base.addSatLum(-0.6, 0.25).toRGB()
		t.colors[tmBannerIcon] = base.addSatLum(0, 0).toRGB()

		t.colors[tmHeading1] = base.toRGB()
		t.colors[tmHeading2] = base.toRGB().mix(altBase.toRGB(), 0.5)
		t.colors[tmHeading3] = altBase.toRGB()

		t.colors[tmParagraph] = base.addSatLum(0, -0.25).toRGB()
		t.colors[tmFirstParagraph] = base.addSatLum(0, -0.1).toRGB()
		t.colors[tmPreformatted] = hslColor{altHue2, 1.0, 0.25}.toRGB()
		t.colors[tmQuote] = t.colors[tmPreformatted]
		t.colors[tmInlineContentMetadata] = t.colors[tmHeading3]
	}
	// end theme-specific code

	t.colors[tmQuoteIcon] = t.colors[tmQuote].
		mix(t.colors[tmBackground], 0.55)
	t.colors[tmOutlineHeadingAbove] = p.White
	t.colors[tmOutlineHeadingBelow] = p.Black

	// begin theme-specific code
	t.colors[tmOutlineHeadingBelow] = t.colors[tmBannerIcon].
		mix(p.White, 0.6)
	// end theme-specific code
}
