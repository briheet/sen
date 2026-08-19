package styles

import (
	"charm.land/lipgloss/v2"
)

// This file bundles every palette from the Lospec anime tag
// (https://lospec.com/palette-list/tag/anime) as a Theme. Lospec palettes are
// flat pixel-art swatches with no ANSI roles, so each semantic field is mapped
// by hand to the color that best fits its meaning:
//
//	Primary    - the palette's signature accent
//	Secondary  - a contrasting accent
//	Text       - the lightest ink for dark palettes, the darkest ink for light
//	            palettes
//	TextMuted  - a desaturated color for low-emphasis text
//	Border     - the base surface color (dark palettes) or a dark separator
//	            color (light palettes)
//	NodeIdle   - a calm, mid-tone color
//	NodeActive - a bright, readable accent
//	NodeHot    - the warmest or most saturated color
//	CPU        - a warm color (gold / orange)
//	Memory     - a green or teal color
//	Success    - the greenest color
//	Warning    - a yellow / orange color
//	Error      - the reddest color
//
// Variable names follow the Lospec slugs, and PixelArtThemes registers every
// palette by its slug.

// Miyazaki 16 — skeddles. A nonlinear 16-color landscape palette inspired by Hayao Miyazaki's art.
var Miyazaki16 = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#2485A6"), // signature accent
	Secondary: lipgloss.Color("#E3C054"), // contrasting accent

	Text:      lipgloss.Color("#EBECDC"), // foreground
	TextMuted: lipgloss.Color("#878573"), // muted text
	Border:    lipgloss.Color("#232228"), // subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#54BAD2"), // dormant
	NodeActive: lipgloss.Color("#55A058"), // active
	NodeHot:    lipgloss.Color("#C65046"), // hot

	// Profiling.
	CPU:    lipgloss.Color("#A1BF41"), // warm
	Memory: lipgloss.Color("#54BAD2"), // green / teal

	// State.
	Success: lipgloss.Color("#55A058"),
	Warning: lipgloss.Color("#E3C054"),
	Error:   lipgloss.Color("#C65046"),
}

// Sailor Moon Background — OldNinjaCat. Dreamy pastel sailor-moon ocean blues and pinks.
var SailorMoonBackground = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#378DAE"), // signature accent
	Secondary: lipgloss.Color("#E055B8"), // contrasting accent

	Text:      lipgloss.Color("#E8F8D7"), // foreground
	TextMuted: lipgloss.Color("#3C7596"), // muted text
	Border:    lipgloss.Color("#00245C"), // subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#A0E1D9"), // dormant
	NodeActive: lipgloss.Color("#5DB8CA"), // active
	NodeHot:    lipgloss.Color("#E055B8"), // hot

	// Profiling.
	CPU:    lipgloss.Color("#FDD3B0"), // warm
	Memory: lipgloss.Color("#80BCB9"), // green / teal

	// State.
	Success: lipgloss.Color("#A0E1D9"),
	Warning: lipgloss.Color("#B65BB2"),
	Error:   lipgloss.Color("#E055B8"),
}

// Heart Gem — danica_cecile. Pastel pinks and blues pulled from a Sailor Moon jewel; light in tone.
var HeartGem = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#88A1D7"), // signature accent
	Secondary: lipgloss.Color("#FF3E8B"), // contrasting accent

	Text:      lipgloss.Color("#76445F"), // foreground
	TextMuted: lipgloss.Color("#925B6E"), // muted text
	Border:    lipgloss.Color("#527E81"), // subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#AE72D4"), // dormant
	NodeActive: lipgloss.Color("#ED7ED7"), // active
	NodeHot:    lipgloss.Color("#EE0E64"), // hot

	// Profiling.
	CPU:    lipgloss.Color("#FBBF2D"), // warm
	Memory: lipgloss.Color("#70AB99"), // green / teal

	// State.
	Success: lipgloss.Color("#70AB99"),
	Warning: lipgloss.Color("#ECA03E"),
	Error:   lipgloss.Color("#EE0E64"),
}

// Bath House — Doph. A 38-color palette based on Spirited Away; warm and celshaded.
var BathHouse = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#E26560"), // signature accent
	Secondary: lipgloss.Color("#B58FB6"), // contrasting accent

	Text:      lipgloss.Color("#DED4C8"), // foreground
	TextMuted: lipgloss.Color("#94837A"), // muted text
	Border:    lipgloss.Color("#181E28"), // subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#30BAB3"), // dormant
	NodeActive: lipgloss.Color("#ABCF5F"), // active
	NodeHot:    lipgloss.Color("#BF2F37"), // hot

	// Profiling.
	CPU:    lipgloss.Color("#E5BE3E"), // warm
	Memory: lipgloss.Color("#789949"), // green / teal

	// State.
	Success: lipgloss.Color("#ABCF5F"),
	Warning: lipgloss.Color("#F18B49"),
	Error:   lipgloss.Color("#BF2F37"),
}

// Heart Gem 2 — danica_cecile. An expanded, brighter Heart Gem with more golds, plums, and greens.
var HeartGem2 = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#6A67C5"), // signature accent
	Secondary: lipgloss.Color("#FF3E8B"), // contrasting accent

	Text:      lipgloss.Color("#673146"), // foreground
	TextMuted: lipgloss.Color("#925B6E"), // muted text
	Border:    lipgloss.Color("#527E81"), // subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#AE72D4"), // dormant
	NodeActive: lipgloss.Color("#8AD29D"), // active
	NodeHot:    lipgloss.Color("#EE0E64"), // hot

	// Profiling.
	CPU:    lipgloss.Color("#F2A600"), // warm
	Memory: lipgloss.Color("#70AB99"), // green / teal

	// State.
	Success: lipgloss.Color("#8AD29D"),
	Warning: lipgloss.Color("#D78612"),
	Error:   lipgloss.Color("#EE0E64"),
}

// Manga Pallete — Kuroi. A 2-color (1-bit) palette of near-black and off-white, like alcohol-marker manga covers.
var MangaPallete = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#F2FBEB"), // signature accent
	Secondary: lipgloss.Color("#171219"), // contrasting accent

	Text:      lipgloss.Color("#F2FBEB"), // foreground
	TextMuted: lipgloss.Color("#171219"), // muted text
	Border:    lipgloss.Color("#171219"), // subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#F2FBEB"), // dormant
	NodeActive: lipgloss.Color("#F2FBEB"), // active
	NodeHot:    lipgloss.Color("#F2FBEB"), // hot

	// Profiling.
	CPU:    lipgloss.Color("#F2FBEB"), // warm
	Memory: lipgloss.Color("#F2FBEB"), // green / teal

	// State.
	Success: lipgloss.Color("#F2FBEB"),
	Warning: lipgloss.Color("#F2FBEB"),
	Error:   lipgloss.Color("#F2FBEB"),
}

// KOTOMASHO-8 — WalGallen. The 8-color palette for the game KOTOMASHO: I Can't Believe This NEET Guy Turned Into a Magical Girl!
var Kotomasho8 = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#5979A6"), // signature accent
	Secondary: lipgloss.Color("#773971"), // contrasting accent

	Text:      lipgloss.Color("#EFEFEF"), // foreground
	TextMuted: lipgloss.Color("#CBC7D6"), // muted text
	Border:    lipgloss.Color("#40263E"), // subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#84C2A3"), // dormant
	NodeActive: lipgloss.Color("#5979A6"), // active
	NodeHot:    lipgloss.Color("#D06060"), // hot

	// Profiling.
	CPU:    lipgloss.Color("#EFE8C3"), // warm
	Memory: lipgloss.Color("#84C2A3"), // green / teal

	// State.
	Success: lipgloss.Color("#84C2A3"),
	Warning: lipgloss.Color("#EFE8C3"),
	Error:   lipgloss.Color("#D06060"),
}

// Kawaii16 — Arisuki. A rainbow 16-color palette for anime characters within PC-98 limits; bright and light.
var Kawaii16 = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#5989A3"), // signature accent
	Secondary: lipgloss.Color("#D793FA"), // contrasting accent

	Text:      lipgloss.Color("#1D173C"), // foreground
	TextMuted: lipgloss.Color("#74518E"), // muted text
	Border:    lipgloss.Color("#65471E"), // subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#7BC188"), // dormant
	NodeActive: lipgloss.Color("#8ED3F8"), // active
	NodeHot:    lipgloss.Color("#EC4646"), // hot

	// Profiling.
	CPU:    lipgloss.Color("#FFA322"), // warm
	Memory: lipgloss.Color("#7BC188"), // green / teal

	// State.
	Success: lipgloss.Color("#7BC188"),
	Warning: lipgloss.Color("#F9FA93"),
	Error:   lipgloss.Color("#EC4646"),
}

// Graph Paper — Jude Buffum. A lo-fi pastel palette for vaporwave geometry and mech doodles; light in tone.
var GraphPaper = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#DDA0AF"), // signature accent
	Secondary: lipgloss.Color("#B9EEDC"), // contrasting accent

	Text:      lipgloss.Color("#555568"), // foreground
	TextMuted: lipgloss.Color("#8E9191"), // muted text
	Border:    lipgloss.Color("#7C7E7E"), // subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#B9EEDC"), // dormant
	NodeActive: lipgloss.Color("#EEB9C7"), // active
	NodeHot:    lipgloss.Color("#DDA0AF"), // hot

	// Profiling.
	CPU:    lipgloss.Color("#EEB9C7"), // warm
	Memory: lipgloss.Color("#B9EEDC"), // green / teal

	// State.
	Success: lipgloss.Color("#B9EEDC"),
	Warning: lipgloss.Color("#EEB9C7"),
	Error:   lipgloss.Color("#DDA0AF"),
}

// Evangelist 16 — bepisman. A 16-color palette heavily inspired by Neon Genesis Evangelion.
var Evangelist16 = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#EB3B35"), // signature accent
	Secondary: lipgloss.Color("#5E1DA8"), // contrasting accent

	Text:      lipgloss.Color("#E3D4FF"), // foreground
	TextMuted: lipgloss.Color("#797DB2"), // muted text
	Border:    lipgloss.Color("#1B2825"), // subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#6A907D"), // dormant
	NodeActive: lipgloss.Color("#B5BBFF"), // active
	NodeHot:    lipgloss.Color("#9F1F4A"), // hot

	// Profiling.
	CPU:    lipgloss.Color("#F2E89D"), // warm
	Memory: lipgloss.Color("#B2C084"), // green / teal

	// State.
	Success: lipgloss.Color("#6A907D"),
	Warning: lipgloss.Color("#F2E89D"),
	Error:   lipgloss.Color("#9F1F4A"),
}

// PixelArtThemes registers every Lospec anime palette by its slug.
var PixelArtThemes = map[string]Theme{
	"miyazaki-16":            Miyazaki16,
	"sailor-moon-background": SailorMoonBackground,
	"heart-gem":              HeartGem,
	"bath-house":             BathHouse,
	"heart-gem-2":            HeartGem2,
	"manga-pallete":          MangaPallete,
	"kotomasho-8":            Kotomasho8,
	"kawaii16":               Kawaii16,
	"graph-paper":            GraphPaper,
	"evangelist-16":          Evangelist16,
}
