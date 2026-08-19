// This file bundles every palette from https://terminalcolors.com as a Theme.
package styles

import (
	"charm.land/lipgloss/v2"
)

// Terminal palettes define sixteen ANSI colors plus a foreground and
// background, so sen maps them onto its semantic fields as follows:
//
//	Primary    <-> bright blue    (primary accent)
//	Secondary  <-> bright magenta (secondary accent)
//	Text       <-> foreground
//	TextMuted  <-> bright black   (muted text)
//	Border     <-> normal black   (subtle separators)
//	NodeIdle   <-> normal cyan   (dormant nodes)
//	NodeActive <-> bright cyan   (active nodes)
//	NodeHot    <-> bright red    (hot nodes)
//	CPU        <-> normal yellow (compute)
//	Memory     <-> normal green  (allocation)
//	Success    <-> bright green
//	Warning    <-> bright yellow
//	Error      <-> bright red
//
// Each variable is named after the terminalcolors.com slug (family-variant).

// Apprentice / default.
var ApprenticeDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#87AFD7"), // bright blue - main accent
	Secondary: lipgloss.Color("#8787AF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#BCBCBC"), // foreground
	TextMuted: lipgloss.Color("#444444"), // bright black - muted text
	Border:    lipgloss.Color("#1C1C1C"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#5F8787"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#5FAFAF"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF8700"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#87875F"), // normal yellow
	Memory: lipgloss.Color("#5F875F"), // normal green

	// State.
	Success: lipgloss.Color("#87AF87"),
	Warning: lipgloss.Color("#FFFFAF"),
	Error:   lipgloss.Color("#FF8700"),
}

// Ayu / dark.
var AyuDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#59C2FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#D2A6FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#BFBDB6"), // foreground
	TextMuted: lipgloss.Color("#686868"), // bright black - muted text
	Border:    lipgloss.Color("#1E232B"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#90E1C6"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#95E6CB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F07178"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F9AF4F"), // normal yellow
	Memory: lipgloss.Color("#7FD962"), // normal green

	// State.
	Success: lipgloss.Color("#AAD94C"),
	Warning: lipgloss.Color("#FFB454"),
	Error:   lipgloss.Color("#F07178"),
}

// Ayu / light.
var AyuLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#399EE6"), // bright blue - main accent
	Secondary: lipgloss.Color("#A37ACC"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#5C6166"), // foreground
	TextMuted: lipgloss.Color("#686868"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#46BA94"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#4CBF99"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F07171"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#ECA944"), // normal yellow
	Memory: lipgloss.Color("#6CBF43"), // normal green

	// State.
	Success: lipgloss.Color("#86B300"),
	Warning: lipgloss.Color("#F2AE49"),
	Error:   lipgloss.Color("#F07171"),
}

// Ayu / mirage.
var AyuMirage = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#73D0FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#DFBFFF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#CCCAC2"), // foreground
	TextMuted: lipgloss.Color("#686868"), // bright black - muted text
	Border:    lipgloss.Color("#171B24"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#90E1C6"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#95E6CB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F28779"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FACC6E"), // normal yellow
	Memory: lipgloss.Color("#87D96C"), // normal green

	// State.
	Success: lipgloss.Color("#D5FF80"),
	Warning: lipgloss.Color("#FFD173"),
	Error:   lipgloss.Color("#F28779"),
}

// Catppuccin / mocha.
var CatppuccinMocha = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#74A8FC"), // bright blue - main accent
	Secondary: lipgloss.Color("#F2AEDE"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#CDD6F4"), // foreground
	TextMuted: lipgloss.Color("#585B70"), // bright black - muted text
	Border:    lipgloss.Color("#45475A"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#94E2D5"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#6BD7CA"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F37799"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F9E2AF"), // normal yellow
	Memory: lipgloss.Color("#A6E3A1"), // normal green

	// State.
	Success: lipgloss.Color("#89D88B"),
	Warning: lipgloss.Color("#EBD391"),
	Error:   lipgloss.Color("#F37799"),
}

// Catppuccin / macchiato.
var CatppuccinMacchiato = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#78A1F6"), // bright blue - main accent
	Secondary: lipgloss.Color("#F2A9DD"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#CAD3F5"), // foreground
	TextMuted: lipgloss.Color("#5B6078"), // bright black - muted text
	Border:    lipgloss.Color("#494D64"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#8BD5CA"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#63CBC0"), // bright cyan - active
	NodeHot:    lipgloss.Color("#EC7486"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EED49F"), // normal yellow
	Memory: lipgloss.Color("#A6DA95"), // normal green

	// State.
	Success: lipgloss.Color("#8CCF7F"),
	Warning: lipgloss.Color("#E1C682"),
	Error:   lipgloss.Color("#EC7486"),
}

// Catppuccin / frappe.
var CatppuccinFrappe = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#7B9EF0"), // bright blue - main accent
	Secondary: lipgloss.Color("#F2A4DB"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C6D0F5"), // foreground
	TextMuted: lipgloss.Color("#626880"), // bright black - muted text
	Border:    lipgloss.Color("#51576D"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#81C8BE"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#5ABFB5"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E67172"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E5C890"), // normal yellow
	Memory: lipgloss.Color("#A6D189"), // normal green

	// State.
	Success: lipgloss.Color("#8EC772"),
	Warning: lipgloss.Color("#D9BA73"),
	Error:   lipgloss.Color("#E67172"),
}

// Catppuccin / latte.
var CatppuccinLatte = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#456EFF"), // bright blue - main accent
	Secondary: lipgloss.Color("#FE85D8"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#4C4F69"), // foreground
	TextMuted: lipgloss.Color("#6C6F85"), // bright black - muted text
	Border:    lipgloss.Color("#5C5F77"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#179299"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#2D9FA8"), // bright cyan - active
	NodeHot:    lipgloss.Color("#DE293E"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#DF8E1D"), // normal yellow
	Memory: lipgloss.Color("#40A02B"), // normal green

	// State.
	Success: lipgloss.Color("#49AF3D"),
	Warning: lipgloss.Color("#EEA02D"),
	Error:   lipgloss.Color("#DE293E"),
}

// Cobalt2 / default.
var Cobalt2Default = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#0088FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#FB94FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#FFFFFF"), // foreground
	TextMuted: lipgloss.Color("#0050A4"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#80FCFF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#80FCFF"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF628C"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FFC600"), // normal yellow
	Memory: lipgloss.Color("#3AD900"), // normal green

	// State.
	Success: lipgloss.Color("#3AD900"),
	Warning: lipgloss.Color("#FFC600"),
	Error:   lipgloss.Color("#FF628C"),
}

// Deus / default.
var DeusDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#73BA9F"), // bright blue - main accent
	Secondary: lipgloss.Color("#C858E9"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#EAEAEA"), // foreground
	TextMuted: lipgloss.Color("#666666"), // bright black - muted text
	Border:    lipgloss.Color("#242A32"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#70C0BA"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#2BCEC2"), // bright cyan - active
	NodeHot:    lipgloss.Color("#EC3E45"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E5C07B"), // normal yellow
	Memory: lipgloss.Color("#98C379"), // normal green

	// State.
	Success: lipgloss.Color("#90C966"),
	Warning: lipgloss.Color("#EDBF69"),
	Error:   lipgloss.Color("#EC3E45"),
}

// Dracula / default.
var DraculaDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#D6ACFF"), // bright blue - main accent
	Secondary: lipgloss.Color("#FF92DF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#F8F8F2"), // foreground
	TextMuted: lipgloss.Color("#6272A4"), // bright black - muted text
	Border:    lipgloss.Color("#21222C"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#8BE9FD"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#A4FFFF"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF6E6E"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F1FA8C"), // normal yellow
	Memory: lipgloss.Color("#50FA7B"), // normal green

	// State.
	Success: lipgloss.Color("#69FF94"),
	Warning: lipgloss.Color("#FFFFA5"),
	Error:   lipgloss.Color("#FF6E6E"),
}

// Dracula / soft.
var DraculaSoft = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#D6B4F7"), // bright blue - main accent
	Secondary: lipgloss.Color("#F49DDA"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#F6F6F4"), // foreground
	TextMuted: lipgloss.Color("#7B7F8B"), // bright black - muted text
	Border:    lipgloss.Color("#262626"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#97E1F1"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#ADF6F6"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F07C7C"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E7EE98"), // normal yellow
	Memory: lipgloss.Color("#62E884"), // normal green

	// State.
	Success: lipgloss.Color("#78F09A"),
	Warning: lipgloss.Color("#F6F6AE"),
	Error:   lipgloss.Color("#F07C7C"),
}

// Everforest / dark.
var EverforestDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#7FBBB3"), // bright blue - main accent
	Secondary: lipgloss.Color("#D699B6"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#D3C6AA"), // foreground
	TextMuted: lipgloss.Color("#859289"), // bright black - muted text
	Border:    lipgloss.Color("#343F44"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#83C092"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#83C092"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E67E80"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#DBBC7F"), // normal yellow
	Memory: lipgloss.Color("#A7C080"), // normal green

	// State.
	Success: lipgloss.Color("#A7C080"),
	Warning: lipgloss.Color("#DBBC7F"),
	Error:   lipgloss.Color("#E67E80"),
}

// Everforest / light.
var EverforestLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#3A94C5"), // bright blue - main accent
	Secondary: lipgloss.Color("#DF69BA"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#5C6A72"), // foreground
	TextMuted: lipgloss.Color("#5C6A72"), // bright black - muted text
	Border:    lipgloss.Color("#5C6A72"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#35A77C"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#35A77C"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F85552"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#DFA000"), // normal yellow
	Memory: lipgloss.Color("#8DA101"), // normal green

	// State.
	Success: lipgloss.Color("#8DA101"),
	Warning: lipgloss.Color("#DFA000"),
	Error:   lipgloss.Color("#F85552"),
}

// GitHub / dark.
var GithubDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#79C0FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#D2A8FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E6EDF3"), // foreground
	TextMuted: lipgloss.Color("#6E7681"), // bright black - muted text
	Border:    lipgloss.Color("#484F58"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#39C5CF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#56D4DD"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FFA198"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#D29922"), // normal yellow
	Memory: lipgloss.Color("#3FB950"), // normal green

	// State.
	Success: lipgloss.Color("#56D364"),
	Warning: lipgloss.Color("#E3B341"),
	Error:   lipgloss.Color("#FFA198"),
}

// GitHub / dark-dimmed.
var GithubDarkDimmed = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#6CB6FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#DCBDFB"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#ADBAC7"), // foreground
	TextMuted: lipgloss.Color("#636E7B"), // bright black - muted text
	Border:    lipgloss.Color("#545D68"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#39C5CF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#56D4DD"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF938A"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#C69026"), // normal yellow
	Memory: lipgloss.Color("#57AB5A"), // normal green

	// State.
	Success: lipgloss.Color("#6BC46D"),
	Warning: lipgloss.Color("#DAAA3F"),
	Error:   lipgloss.Color("#FF938A"),
}

// GitHub / dark-high-contrast.
var GithubDarkHighContrast = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#91CBFF"), // bright blue - main accent
	Secondary: lipgloss.Color("#DBB7FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#F0F3F6"), // foreground
	TextMuted: lipgloss.Color("#9EA7B3"), // bright black - muted text
	Border:    lipgloss.Color("#7A828E"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#39C5CF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#56D4DD"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FFB1AF"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F0B72F"), // normal yellow
	Memory: lipgloss.Color("#26CD4D"), // normal green

	// State.
	Success: lipgloss.Color("#4AE168"),
	Warning: lipgloss.Color("#F7C843"),
	Error:   lipgloss.Color("#FFB1AF"),
}

// GitHub / dark-colorblind.
var GithubDarkColorblind = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#79C0FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#D2A8FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C9D1D9"), // foreground
	TextMuted: lipgloss.Color("#6E7681"), // bright black - muted text
	Border:    lipgloss.Color("#484F58"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#39C5CF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#56D4DD"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FDAC54"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#D29922"), // normal yellow
	Memory: lipgloss.Color("#58A6FF"), // normal green

	// State.
	Success: lipgloss.Color("#79C0FF"),
	Warning: lipgloss.Color("#E3B341"),
	Error:   lipgloss.Color("#FDAC54"),
}

// GitHub / dark-legacy.
var GithubDarkLegacy = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#79B8FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#B392F0"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#D1D5DA"), // foreground
	TextMuted: lipgloss.Color("#959DA5"), // bright black - muted text
	Border:    lipgloss.Color("#586069"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#39C5CF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#56D4DD"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F97583"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FFEA7F"), // normal yellow
	Memory: lipgloss.Color("#34D058"), // normal green

	// State.
	Success: lipgloss.Color("#85E89D"),
	Warning: lipgloss.Color("#FFEA7F"),
	Error:   lipgloss.Color("#F97583"),
}

// GitHub / light.
var GithubLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#218BFF"), // bright blue - main accent
	Secondary: lipgloss.Color("#A475F9"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#1F2328"), // foreground
	TextMuted: lipgloss.Color("#57606A"), // bright black - muted text
	Border:    lipgloss.Color("#24292F"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#1B7C83"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#3192AA"), // bright cyan - active
	NodeHot:    lipgloss.Color("#A40E26"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#4D2D00"), // normal yellow
	Memory: lipgloss.Color("#116329"), // normal green

	// State.
	Success: lipgloss.Color("#1A7F37"),
	Warning: lipgloss.Color("#633C01"),
	Error:   lipgloss.Color("#A40E26"),
}

// GitHub / light-high-contrast.
var GithubLightHighContrast = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#1168E3"), // bright blue - main accent
	Secondary: lipgloss.Color("#844AE7"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#0E1116"), // foreground
	TextMuted: lipgloss.Color("#4B535D"), // bright black - muted text
	Border:    lipgloss.Color("#0E1116"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#1B7C83"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#3192AA"), // bright cyan - active
	NodeHot:    lipgloss.Color("#86061D"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#3F2200"), // normal yellow
	Memory: lipgloss.Color("#024C1A"), // normal green

	// State.
	Success: lipgloss.Color("#055D20"),
	Warning: lipgloss.Color("#4E2C00"),
	Error:   lipgloss.Color("#86061D"),
}

// GitHub / light-colorblind.
var GithubLightColorblind = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#218BFF"), // bright blue - main accent
	Secondary: lipgloss.Color("#A475F9"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#24292F"), // foreground
	TextMuted: lipgloss.Color("#57606A"), // bright black - muted text
	Border:    lipgloss.Color("#24292F"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#1B7C83"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#3192AA"), // bright cyan - active
	NodeHot:    lipgloss.Color("#8A4600"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#4D2D00"), // normal yellow
	Memory: lipgloss.Color("#0550AE"), // normal green

	// State.
	Success: lipgloss.Color("#0969DA"),
	Warning: lipgloss.Color("#633C01"),
	Error:   lipgloss.Color("#8A4600"),
}

// GitHub / light-legacy.
var GithubLightLegacy = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#005CC5"), // bright blue - main accent
	Secondary: lipgloss.Color("#5A32A3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#586069"), // foreground
	TextMuted: lipgloss.Color("#959DA5"), // bright black - muted text
	Border:    lipgloss.Color("#24292E"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#1B7C83"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#3192AA"), // bright cyan - active
	NodeHot:    lipgloss.Color("#CB2431"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#DBAB09"), // normal yellow
	Memory: lipgloss.Color("#28A745"), // normal green

	// State.
	Success: lipgloss.Color("#22863A"),
	Warning: lipgloss.Color("#B08800"),
	Error:   lipgloss.Color("#CB2431"),
}

// Gotham / default.
var GothamDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#195466"), // bright blue - main accent
	Secondary: lipgloss.Color("#4E5166"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#99D1CE"), // foreground
	TextMuted: lipgloss.Color("#0C1014"), // bright black - muted text
	Border:    lipgloss.Color("#0C1014"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#33859E"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#33859E"), // bright cyan - active
	NodeHot:    lipgloss.Color("#C23127"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EDB443"), // normal yellow
	Memory: lipgloss.Color("#2AA889"), // normal green

	// State.
	Success: lipgloss.Color("#2AA889"),
	Warning: lipgloss.Color("#EDB443"),
	Error:   lipgloss.Color("#C23127"),
}

// Gruvbox / dark.
var GruvboxDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#83A598"), // bright blue - main accent
	Secondary: lipgloss.Color("#D3869B"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#EBDBB2"), // foreground
	TextMuted: lipgloss.Color("#928374"), // bright black - muted text
	Border:    lipgloss.Color("#282828"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#689D6A"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#8EC07C"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FB4934"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#D79921"), // normal yellow
	Memory: lipgloss.Color("#98971A"), // normal green

	// State.
	Success: lipgloss.Color("#B8BB26"),
	Warning: lipgloss.Color("#FABD2F"),
	Error:   lipgloss.Color("#FB4934"),
}

// Gruvbox / dark-hard-contrast.
var GruvboxDarkHardContrast = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#83A598"), // bright blue - main accent
	Secondary: lipgloss.Color("#D3869B"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#EBDBB2"), // foreground
	TextMuted: lipgloss.Color("#928374"), // bright black - muted text
	Border:    lipgloss.Color("#1D2021"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#689D6A"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#8EC07C"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FB4934"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#D79921"), // normal yellow
	Memory: lipgloss.Color("#98971A"), // normal green

	// State.
	Success: lipgloss.Color("#B8BB26"),
	Warning: lipgloss.Color("#FABD2F"),
	Error:   lipgloss.Color("#FB4934"),
}

// Gruvbox / dark-soft-contrast.
var GruvboxDarkSoftContrast = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#83A598"), // bright blue - main accent
	Secondary: lipgloss.Color("#D3869B"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#EBDBB2"), // foreground
	TextMuted: lipgloss.Color("#928374"), // bright black - muted text
	Border:    lipgloss.Color("#32302F"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#689D6A"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#8EC07C"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FB4934"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#D79921"), // normal yellow
	Memory: lipgloss.Color("#98971A"), // normal green

	// State.
	Success: lipgloss.Color("#B8BB26"),
	Warning: lipgloss.Color("#FABD2F"),
	Error:   lipgloss.Color("#FB4934"),
}

// Gruvbox / light.
var GruvboxLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#076678"), // bright blue - main accent
	Secondary: lipgloss.Color("#8F3F71"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#3C3836"), // foreground
	TextMuted: lipgloss.Color("#928374"), // bright black - muted text
	Border:    lipgloss.Color("#FBF1C7"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#689D6A"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#427B58"), // bright cyan - active
	NodeHot:    lipgloss.Color("#9D0006"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#D79921"), // normal yellow
	Memory: lipgloss.Color("#98971A"), // normal green

	// State.
	Success: lipgloss.Color("#79740E"),
	Warning: lipgloss.Color("#B57614"),
	Error:   lipgloss.Color("#9D0006"),
}

// Gruvbox / light-hard-contrast.
var GruvboxLightHardContrast = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#076678"), // bright blue - main accent
	Secondary: lipgloss.Color("#8F3F71"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#3C3836"), // foreground
	TextMuted: lipgloss.Color("#928374"), // bright black - muted text
	Border:    lipgloss.Color("#F9F5D7"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#689D6A"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#427B58"), // bright cyan - active
	NodeHot:    lipgloss.Color("#9D0006"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#D79921"), // normal yellow
	Memory: lipgloss.Color("#98971A"), // normal green

	// State.
	Success: lipgloss.Color("#79740E"),
	Warning: lipgloss.Color("#B57614"),
	Error:   lipgloss.Color("#9D0006"),
}

// Gruvbox / light-soft-contrast.
var GruvboxLightSoftContrast = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#076678"), // bright blue - main accent
	Secondary: lipgloss.Color("#8F3F71"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#3C3836"), // foreground
	TextMuted: lipgloss.Color("#928374"), // bright black - muted text
	Border:    lipgloss.Color("#F2E5BC"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#689D6A"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#427B58"), // bright cyan - active
	NodeHot:    lipgloss.Color("#9D0006"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#D79921"), // normal yellow
	Memory: lipgloss.Color("#98971A"), // normal green

	// State.
	Success: lipgloss.Color("#79740E"),
	Warning: lipgloss.Color("#B57614"),
	Error:   lipgloss.Color("#9D0006"),
}

// Iceberg / dark.
var IcebergDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#91ACD1"), // bright blue - main accent
	Secondary: lipgloss.Color("#ADA0D3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C6C8D1"), // foreground
	TextMuted: lipgloss.Color("#6B7089"), // bright black - muted text
	Border:    lipgloss.Color("#1E2132"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#89B8C2"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#95C4CE"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E98989"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E2A478"), // normal yellow
	Memory: lipgloss.Color("#B4BE82"), // normal green

	// State.
	Success: lipgloss.Color("#C0CA8E"),
	Warning: lipgloss.Color("#E9B189"),
	Error:   lipgloss.Color("#E98989"),
}

// Iceberg / light.
var IcebergLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#22478E"), // bright blue - main accent
	Secondary: lipgloss.Color("#6845AD"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#33374C"), // foreground
	TextMuted: lipgloss.Color("#8389A3"), // bright black - muted text
	Border:    lipgloss.Color("#DCDFE7"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#3F83A6"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#327698"), // bright cyan - active
	NodeHot:    lipgloss.Color("#CC3768"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#C57339"), // normal yellow
	Memory: lipgloss.Color("#668E3D"), // normal green

	// State.
	Success: lipgloss.Color("#598030"),
	Warning: lipgloss.Color("#B6662D"),
	Error:   lipgloss.Color("#CC3768"),
}

// Jellybeans / default.
var JellybeansDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#B1D8F6"), // bright blue - main accent
	Secondary: lipgloss.Color("#FBDAFF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#DEDEDE"), // foreground
	TextMuted: lipgloss.Color("#BDBDBD"), // bright black - muted text
	Border:    lipgloss.Color("#929292"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#00988E"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#1AB2A8"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FFA1A1"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FFBA7B"), // normal yellow
	Memory: lipgloss.Color("#94B979"), // normal green

	// State.
	Success: lipgloss.Color("#BDDEAB"),
	Warning: lipgloss.Color("#FFDCA0"),
	Error:   lipgloss.Color("#FFA1A1"),
}

// Kanagawa / wave.
var KanagawaWave = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#7FB4CA"), // bright blue - main accent
	Secondary: lipgloss.Color("#938AA9"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#DCD7BA"), // foreground
	TextMuted: lipgloss.Color("#727169"), // bright black - muted text
	Border:    lipgloss.Color("#16161D"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#6A9589"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#7AA89F"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E82424"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#C0A36E"), // normal yellow
	Memory: lipgloss.Color("#76946A"), // normal green

	// State.
	Success: lipgloss.Color("#98BB6C"),
	Warning: lipgloss.Color("#E6C384"),
	Error:   lipgloss.Color("#E82424"),
}

// Kanagawa / dragon.
var KanagawaDragon = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#7FB4CA"), // bright blue - main accent
	Secondary: lipgloss.Color("#938AA9"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C5C9C5"), // foreground
	TextMuted: lipgloss.Color("#A6A69C"), // bright black - muted text
	Border:    lipgloss.Color("#0D0C0C"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#8EA4A2"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#7AA89F"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E46876"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#C4B28A"), // normal yellow
	Memory: lipgloss.Color("#8A9A7B"), // normal green

	// State.
	Success: lipgloss.Color("#87A987"),
	Warning: lipgloss.Color("#E6C384"),
	Error:   lipgloss.Color("#E46876"),
}

// Kanagawa / lotus.
var KanagawaLotus = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#6693BF"), // bright blue - main accent
	Secondary: lipgloss.Color("#624C83"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#545464"), // foreground
	TextMuted: lipgloss.Color("#8A8980"), // bright black - muted text
	Border:    lipgloss.Color("#1F1F28"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#597B75"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#5E857A"), // bright cyan - active
	NodeHot:    lipgloss.Color("#D7474B"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#77713F"), // normal yellow
	Memory: lipgloss.Color("#6F894E"), // normal green

	// State.
	Success: lipgloss.Color("#6E915F"),
	Warning: lipgloss.Color("#836F4A"),
	Error:   lipgloss.Color("#D7474B"),
}

// Lucario / default.
var LucarioDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#D6ACFF"), // bright blue - main accent
	Secondary: lipgloss.Color("#D4A9FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#F8F8F2"), // foreground
	TextMuted: lipgloss.Color("#2F3943"), // bright black - muted text
	Border:    lipgloss.Color("#19242F"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#8BE0FD"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#B9ECFD"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF6541"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F0CC04"), // normal yellow
	Memory: lipgloss.Color("#199C4B"), // normal green

	// State.
	Success: lipgloss.Color("#72CC5A"),
	Warning: lipgloss.Color("#FFFFA5"),
	Error:   lipgloss.Color("#FF6541"),
}

// Miasma / default.
var MiasmaDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#78824B"), // bright blue - main accent
	Secondary: lipgloss.Color("#BB7744"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C2C2B0"), // foreground
	TextMuted: lipgloss.Color("#666666"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#C9A554"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#C9A554"), // bright cyan - active
	NodeHot:    lipgloss.Color("#685742"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#B36D43"), // normal yellow
	Memory: lipgloss.Color("#5F875F"), // normal green

	// State.
	Success: lipgloss.Color("#5F875F"),
	Warning: lipgloss.Color("#B36D43"),
	Error:   lipgloss.Color("#685742"),
}

// Moonfly / default.
var MoonflyDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#74B2FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#AE81FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#BDBDBD"), // foreground
	TextMuted: lipgloss.Color("#949494"), // bright black - muted text
	Border:    lipgloss.Color("#323437"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#79DAC8"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#85DC85"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF5189"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E3C78A"), // normal yellow
	Memory: lipgloss.Color("#8CC85F"), // normal green

	// State.
	Success: lipgloss.Color("#36C692"),
	Warning: lipgloss.Color("#C6C684"),
	Error:   lipgloss.Color("#FF5189"),
}

// Night Owl / dark.
var NightOwlDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#82AAFF"), // bright blue - main accent
	Secondary: lipgloss.Color("#C792EA"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#CCCCCC"), // foreground
	TextMuted: lipgloss.Color("#575656"), // bright black - muted text
	Border:    lipgloss.Color("#011627"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#21C7A8"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#7FDBCA"), // bright cyan - active
	NodeHot:    lipgloss.Color("#EF5350"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#C5E478"), // normal yellow
	Memory: lipgloss.Color("#22DA6E"), // normal green

	// State.
	Success: lipgloss.Color("#22DA6E"),
	Warning: lipgloss.Color("#FFEB95"),
	Error:   lipgloss.Color("#EF5350"),
}

// Night Owl / light.
var NightOwlLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#288ED7"), // bright blue - main accent
	Secondary: lipgloss.Color("#D6438A"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#403F53"), // foreground
	TextMuted: lipgloss.Color("#403F53"), // bright black - muted text
	Border:    lipgloss.Color("#403F53"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#2AA298"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#2AA298"), // bright cyan - active
	NodeHot:    lipgloss.Color("#DE3D3B"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E0AF02"), // normal yellow
	Memory: lipgloss.Color("#08916A"), // normal green

	// State.
	Success: lipgloss.Color("#08916A"),
	Warning: lipgloss.Color("#DAAA01"),
	Error:   lipgloss.Color("#DE3D3B"),
}

// Nightfly / default.
var NightflyDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#82AAFF"), // bright blue - main accent
	Secondary: lipgloss.Color("#AE81FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#BDC1C6"), // foreground
	TextMuted: lipgloss.Color("#7C8F8F"), // bright black - muted text
	Border:    lipgloss.Color("#1D3B53"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#7FDBCA"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#7FDBCA"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF5874"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E7D37A"), // normal yellow
	Memory: lipgloss.Color("#A1CD5E"), // normal green

	// State.
	Success: lipgloss.Color("#21C7A8"),
	Warning: lipgloss.Color("#ECC48D"),
	Error:   lipgloss.Color("#FF5874"),
}

// Nightfox / default.
var NightfoxDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#86ABDC"), // bright blue - main accent
	Secondary: lipgloss.Color("#BAA1E2"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#CDCECF"), // foreground
	TextMuted: lipgloss.Color("#575860"), // bright black - muted text
	Border:    lipgloss.Color("#393B44"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#63CDCF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#7AD5D6"), // bright cyan - active
	NodeHot:    lipgloss.Color("#D16983"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#DBC074"), // normal yellow
	Memory: lipgloss.Color("#81B29A"), // normal green

	// State.
	Success: lipgloss.Color("#8EBAA4"),
	Warning: lipgloss.Color("#E0C989"),
	Error:   lipgloss.Color("#D16983"),
}

// Nightfox / dayfox.
var NightfoxDayfox = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#4863B6"), // bright blue - main accent
	Secondary: lipgloss.Color("#8452D5"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#3D2B5A"), // foreground
	TextMuted: lipgloss.Color("#534C45"), // bright black - muted text
	Border:    lipgloss.Color("#352C24"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#287980"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#488D93"), // bright cyan - active
	NodeHot:    lipgloss.Color("#B3434E"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#AC5402"), // normal yellow
	Memory: lipgloss.Color("#396847"), // normal green

	// State.
	Success: lipgloss.Color("#577F63"),
	Warning: lipgloss.Color("#B86E28"),
	Error:   lipgloss.Color("#B3434E"),
}

// Nightfox / dawnfox.
var NightfoxDawnfox = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#2D81A3"), // bright blue - main accent
	Secondary: lipgloss.Color("#9A80B9"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#575279"), // foreground
	TextMuted: lipgloss.Color("#5F5695"), // bright black - muted text
	Border:    lipgloss.Color("#575279"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#56949F"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#5CA7B4"), // bright cyan - active
	NodeHot:    lipgloss.Color("#C26D85"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EA9D34"), // normal yellow
	Memory: lipgloss.Color("#618774"), // normal green

	// State.
	Success: lipgloss.Color("#629F81"),
	Warning: lipgloss.Color("#EEA846"),
	Error:   lipgloss.Color("#C26D85"),
}

// Nightfox / duskfox.
var NightfoxDuskfox = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#65B1CD"), // bright blue - main accent
	Secondary: lipgloss.Color("#CCB1ED"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E0DEF4"), // foreground
	TextMuted: lipgloss.Color("#47407D"), // bright black - muted text
	Border:    lipgloss.Color("#393552"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#9CCFD8"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#A6DAE3"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F083A2"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F6C177"), // normal yellow
	Memory: lipgloss.Color("#A3BE8C"), // normal green

	// State.
	Success: lipgloss.Color("#B1D196"),
	Warning: lipgloss.Color("#F9CB8C"),
	Error:   lipgloss.Color("#F083A2"),
}

// Nightfox / nordfox.
var NightfoxNordfox = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#8CAFD2"), // bright blue - main accent
	Secondary: lipgloss.Color("#C895BF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#CDCECF"), // foreground
	TextMuted: lipgloss.Color("#465780"), // bright black - muted text
	Border:    lipgloss.Color("#3B4252"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#88C0D0"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#93CCDC"), // bright cyan - active
	NodeHot:    lipgloss.Color("#D06F79"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EBCB8B"), // normal yellow
	Memory: lipgloss.Color("#A3BE8C"), // normal green

	// State.
	Success: lipgloss.Color("#B1D196"),
	Warning: lipgloss.Color("#F0D399"),
	Error:   lipgloss.Color("#D06F79"),
}

// Nightfox / terafox.
var NightfoxTerafox = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#73A3B7"), // bright blue - main accent
	Secondary: lipgloss.Color("#B97490"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E6EAEA"), // foreground
	TextMuted: lipgloss.Color("#4E5157"), // bright black - muted text
	Border:    lipgloss.Color("#2F3239"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#A1CDD8"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#AFD4DE"), // bright cyan - active
	NodeHot:    lipgloss.Color("#EB746B"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FDA47F"), // normal yellow
	Memory: lipgloss.Color("#7AA4A1"), // normal green

	// State.
	Success: lipgloss.Color("#8EB2AF"),
	Warning: lipgloss.Color("#FDB292"),
	Error:   lipgloss.Color("#EB746B"),
}

// Nightfox / carbonfox.
var NightfoxCarbonfox = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#8CB6FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#C8A5FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#F2F4F8"), // foreground
	TextMuted: lipgloss.Color("#484848"), // bright black - muted text
	Border:    lipgloss.Color("#282828"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#33B1FF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#52BDFF"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F16DA6"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#08BDBA"), // normal yellow
	Memory: lipgloss.Color("#25BE6A"), // normal green

	// State.
	Success: lipgloss.Color("#46C880"),
	Warning: lipgloss.Color("#2DC7C4"),
	Error:   lipgloss.Color("#F16DA6"),
}

// Noctis / default.
var NoctisDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#60B6EB"), // bright blue - main accent
	Secondary: lipgloss.Color("#E798B3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#B2CACD"), // foreground
	TextMuted: lipgloss.Color("#47686C"), // bright black - muted text
	Border:    lipgloss.Color("#324A4D"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#49D6E9"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#60DBEB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E97749"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E4B781"), // normal yellow
	Memory: lipgloss.Color("#49E9A6"), // normal green

	// State.
	Success: lipgloss.Color("#60EBB1"),
	Warning: lipgloss.Color("#E69533"),
	Error:   lipgloss.Color("#E97749"),
}

// Noctis / azureus.
var NoctisAzureus = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#60B6EB"), // bright blue - main accent
	Secondary: lipgloss.Color("#E798B3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#BECFDA"), // foreground
	TextMuted: lipgloss.Color("#475E6C"), // bright black - muted text
	Border:    lipgloss.Color("#28353E"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#49D6E9"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#60DBEB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E97749"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E4B781"), // normal yellow
	Memory: lipgloss.Color("#49E9A6"), // normal green

	// State.
	Success: lipgloss.Color("#60EBB1"),
	Warning: lipgloss.Color("#E69533"),
	Error:   lipgloss.Color("#E97749"),
}

// Noctis / bordo.
var NoctisBordo = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#60B6EB"), // bright blue - main accent
	Secondary: lipgloss.Color("#E798B3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#CBBEC2"), // foreground
	TextMuted: lipgloss.Color("#69545B"), // bright black - muted text
	Border:    lipgloss.Color("#47393E"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#49D6E9"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#60DBEB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E97749"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E4B781"), // normal yellow
	Memory: lipgloss.Color("#49E9A6"), // normal green

	// State.
	Success: lipgloss.Color("#60EBB1"),
	Warning: lipgloss.Color("#E69533"),
	Error:   lipgloss.Color("#E97749"),
}

// Noctis / minimus.
var NoctisMinimus = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#68A4CA"), // bright blue - main accent
	Secondary: lipgloss.Color("#C88DA2"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C5CDD3"), // foreground
	TextMuted: lipgloss.Color("#425866"), // bright black - muted text
	Border:    lipgloss.Color("#182A35"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#72B7C0"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#84C0C8"), // bright cyan - active
	NodeHot:    lipgloss.Color("#CA8468"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#C8A984"), // normal yellow
	Memory: lipgloss.Color("#72C09F"), // normal green

	// State.
	Success: lipgloss.Color("#84C8AB"),
	Warning: lipgloss.Color("#D1AA7B"),
	Error:   lipgloss.Color("#CA8468"),
}

// Noctis / uva.
var NoctisUva = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#60B6EB"), // bright blue - main accent
	Secondary: lipgloss.Color("#E798B3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C5C2D6"), // foreground
	TextMuted: lipgloss.Color("#504E65"), // bright black - muted text
	Border:    lipgloss.Color("#302F3D"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#49D6E9"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#60DBEB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E97749"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E4B781"), // normal yellow
	Memory: lipgloss.Color("#49E9A6"), // normal green

	// State.
	Success: lipgloss.Color("#60EBB1"),
	Warning: lipgloss.Color("#E69533"),
	Error:   lipgloss.Color("#E97749"),
}

// Noctis / viola.
var NoctisViola = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#60B6EB"), // bright blue - main accent
	Secondary: lipgloss.Color("#E798B3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#CCBFD9"), // foreground
	TextMuted: lipgloss.Color("#594E65"), // bright black - muted text
	Border:    lipgloss.Color("#362F3D"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#49D6E9"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#60DBEB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E97749"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E4B781"), // normal yellow
	Memory: lipgloss.Color("#49E9A6"), // normal green

	// State.
	Success: lipgloss.Color("#60EBB1"),
	Warning: lipgloss.Color("#E69533"),
	Error:   lipgloss.Color("#E97749"),
}

// Noctis / lux.
var NoctisLux = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#0FA3FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#FF6B9F"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#005661"), // foreground
	TextMuted: lipgloss.Color("#004D57"), // bright black - muted text
	Border:    lipgloss.Color("#003B42"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#00BDD6"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#00CBE6"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF4000"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F49725"), // normal yellow
	Memory: lipgloss.Color("#00B368"), // normal green

	// State.
	Success: lipgloss.Color("#00D17A"),
	Warning: lipgloss.Color("#FF8C00"),
	Error:   lipgloss.Color("#FF4000"),
}

// Noctis / lilac.
var NoctisLilac = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#0FA3FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#FF6B9F"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#0C006B"), // foreground
	TextMuted: lipgloss.Color("#0F0080"), // bright black - muted text
	Border:    lipgloss.Color("#0C006B"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#00BDD6"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#00CBE6"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF4000"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F49725"), // normal yellow
	Memory: lipgloss.Color("#00B368"), // normal green

	// State.
	Success: lipgloss.Color("#00D17A"),
	Warning: lipgloss.Color("#FF8C00"),
	Error:   lipgloss.Color("#FF4000"),
}

// Noctis / hibernus.
var NoctisHibernus = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#0FA3FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#FF6B9F"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#005661"), // foreground
	TextMuted: lipgloss.Color("#004D57"), // bright black - muted text
	Border:    lipgloss.Color("#003B42"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#00BDD6"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#00CBE6"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF4000"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F49725"), // normal yellow
	Memory: lipgloss.Color("#00B368"), // normal green

	// State.
	Success: lipgloss.Color("#00D17A"),
	Warning: lipgloss.Color("#FF8C00"),
	Error:   lipgloss.Color("#FF4000"),
}

// Noctis / obscuro.
var NoctisObscuro = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#60B6EB"), // bright blue - main accent
	Secondary: lipgloss.Color("#E798B3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#B2CACD"), // foreground
	TextMuted: lipgloss.Color("#47686C"), // bright black - muted text
	Border:    lipgloss.Color("#324A4D"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#49D6E9"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#60DBEB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E97749"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E4B781"), // normal yellow
	Memory: lipgloss.Color("#49E9A6"), // normal green

	// State.
	Success: lipgloss.Color("#60EBB1"),
	Warning: lipgloss.Color("#E69533"),
	Error:   lipgloss.Color("#E97749"),
}

// Noctis / sereno.
var NoctisSereno = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#60B6EB"), // bright blue - main accent
	Secondary: lipgloss.Color("#E798B3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#B2CACD"), // foreground
	TextMuted: lipgloss.Color("#47686C"), // bright black - muted text
	Border:    lipgloss.Color("#324A4D"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#49D6E9"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#60DBEB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E97749"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E4B781"), // normal yellow
	Memory: lipgloss.Color("#49E9A6"), // normal green

	// State.
	Success: lipgloss.Color("#60EBB1"),
	Warning: lipgloss.Color("#E69533"),
	Error:   lipgloss.Color("#E97749"),
}

// Nord / default.
var NordDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#81A1C1"), // bright blue - main accent
	Secondary: lipgloss.Color("#B48EAD"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#D8DEE9"), // foreground
	TextMuted: lipgloss.Color("#4C566A"), // bright black - muted text
	Border:    lipgloss.Color("#3B4252"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#88C0D0"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#8FBCBB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#BF616A"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EBCB8B"), // normal yellow
	Memory: lipgloss.Color("#A3BE8C"), // normal green

	// State.
	Success: lipgloss.Color("#A3BE8C"),
	Warning: lipgloss.Color("#EBCB8B"),
	Error:   lipgloss.Color("#BF616A"),
}

// Nordic / default.
var NordicDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#88C0D0"), // bright blue - main accent
	Secondary: lipgloss.Color("#BE9D88"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#BBC3D4"), // foreground
	TextMuted: lipgloss.Color("#3B4252"), // bright black - muted text
	Border:    lipgloss.Color("#191D24"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#8FBCBB"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#9FC6C5"), // bright cyan - active
	NodeHot:    lipgloss.Color("#C5727A"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EBCB8B"), // normal yellow
	Memory: lipgloss.Color("#A3BE8C"), // normal green

	// State.
	Success: lipgloss.Color("#B1C89D"),
	Warning: lipgloss.Color("#EFD49F"),
	Error:   lipgloss.Color("#C5727A"),
}

// One / dark.
var OneDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#61AFEF"), // bright blue - main accent
	Secondary: lipgloss.Color("#C678DD"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#ABB2BF"), // foreground
	TextMuted: lipgloss.Color("#5C6370"), // bright black - muted text
	Border:    lipgloss.Color("#1E2127"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#56B6C2"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#56B6C2"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E06C75"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#D19A66"), // normal yellow
	Memory: lipgloss.Color("#98C379"), // normal green

	// State.
	Success: lipgloss.Color("#98C379"),
	Warning: lipgloss.Color("#D19A66"),
	Error:   lipgloss.Color("#E06C75"),
}

// One / light.
var OneLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#2F5AF3"), // bright blue - main accent
	Secondary: lipgloss.Color("#A00095"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#2A2B33"), // foreground
	TextMuted: lipgloss.Color("#000000"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#3E953A"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#3E953A"), // bright cyan - active
	NodeHot:    lipgloss.Color("#DE3D35"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#D2B67B"), // normal yellow
	Memory: lipgloss.Color("#3E953A"), // normal green

	// State.
	Success: lipgloss.Color("#3E953A"),
	Warning: lipgloss.Color("#D2B67B"),
	Error:   lipgloss.Color("#DE3D35"),
}

// One Half / dark.
var OneHalfDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#61AFEF"), // bright blue - main accent
	Secondary: lipgloss.Color("#C678DD"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#DCDFE4"), // foreground
	TextMuted: lipgloss.Color("#282C34"), // bright black - muted text
	Border:    lipgloss.Color("#282C34"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#56B6C2"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#56B6C2"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E06C75"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E5C07B"), // normal yellow
	Memory: lipgloss.Color("#98C379"), // normal green

	// State.
	Success: lipgloss.Color("#98C379"),
	Warning: lipgloss.Color("#E5C07B"),
	Error:   lipgloss.Color("#E06C75"),
}

// One Half / light.
var OneHalfLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#61AFEF"), // bright blue - main accent
	Secondary: lipgloss.Color("#C678DD"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#383A42"), // foreground
	TextMuted: lipgloss.Color("#4F525E"), // bright black - muted text
	Border:    lipgloss.Color("#383A42"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#0997B3"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#56B6C2"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E06C75"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#C18401"), // normal yellow
	Memory: lipgloss.Color("#50A14F"), // normal green

	// State.
	Success: lipgloss.Color("#98C379"),
	Warning: lipgloss.Color("#E5C07B"),
	Error:   lipgloss.Color("#E06C75"),
}

// Panda / default.
var PandaDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#6FC1FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#FF9AC1"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#CCCCCC"), // foreground
	TextMuted: lipgloss.Color("#757575"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#B084EB"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#BCAAFE"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF2C6D"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FFB86C"), // normal yellow
	Memory: lipgloss.Color("#19F9D8"), // normal green

	// State.
	Success: lipgloss.Color("#19F9D8"),
	Warning: lipgloss.Color("#FFCC95"),
	Error:   lipgloss.Color("#FF2C6D"),
}

// Posterpole / default.
var PosterpoleDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#8A99A8"), // bright blue - main accent
	Secondary: lipgloss.Color("#CCB3C6"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C6C0B9"), // foreground
	TextMuted: lipgloss.Color("#A5A59C"), // bright black - muted text
	Border:    lipgloss.Color("#2C2C30"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#8EA4A2"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#AABBBA"), // bright cyan - active
	NodeHot:    lipgloss.Color("#BC8F8F"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#CC9166"), // normal yellow
	Memory: lipgloss.Color("#778C73"), // normal green

	// State.
	Success: lipgloss.Color("#92A38F"),
	Warning: lipgloss.Color("#D9AC8C"),
	Error:   lipgloss.Color("#BC8F8F"),
}

// Posterpole / gray.
var PosterpoleGray = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#8A99A8"), // bright blue - main accent
	Secondary: lipgloss.Color("#CCB3C6"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C6C0B9"), // foreground
	TextMuted: lipgloss.Color("#A5A59C"), // bright black - muted text
	Border:    lipgloss.Color("#2C2C30"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#8EA4A2"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#AABBBA"), // bright cyan - active
	NodeHot:    lipgloss.Color("#BC8F8F"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#CC9166"), // normal yellow
	Memory: lipgloss.Color("#778C73"), // normal green

	// State.
	Success: lipgloss.Color("#92A38F"),
	Warning: lipgloss.Color("#D9AC8C"),
	Error:   lipgloss.Color("#BC8F8F"),
}

// Rosé Pine / default.
var RosePineDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#9CCFD8"), // bright blue - main accent
	Secondary: lipgloss.Color("#C4A7E7"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E0DEF4"), // foreground
	TextMuted: lipgloss.Color("#908CAA"), // bright black - muted text
	Border:    lipgloss.Color("#26233A"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#EBBCBA"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#EBBCBA"), // bright cyan - active
	NodeHot:    lipgloss.Color("#EB6F92"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F6C177"), // normal yellow
	Memory: lipgloss.Color("#31748F"), // normal green

	// State.
	Success: lipgloss.Color("#31748F"),
	Warning: lipgloss.Color("#F6C177"),
	Error:   lipgloss.Color("#EB6F92"),
}

// Rosé Pine / moon.
var RosePineMoon = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#9CCFD8"), // bright blue - main accent
	Secondary: lipgloss.Color("#C4A7E7"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E0DEF4"), // foreground
	TextMuted: lipgloss.Color("#908CAA"), // bright black - muted text
	Border:    lipgloss.Color("#393552"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#EA9A97"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#EA9A97"), // bright cyan - active
	NodeHot:    lipgloss.Color("#EB6F92"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F6C177"), // normal yellow
	Memory: lipgloss.Color("#3E8FB0"), // normal green

	// State.
	Success: lipgloss.Color("#3E8FB0"),
	Warning: lipgloss.Color("#F6C177"),
	Error:   lipgloss.Color("#EB6F92"),
}

// Rosé Pine / dawn.
var RosePineDawn = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#56949F"), // bright blue - main accent
	Secondary: lipgloss.Color("#907AA9"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#575279"), // foreground
	TextMuted: lipgloss.Color("#797593"), // bright black - muted text
	Border:    lipgloss.Color("#F2E9E1"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#D7827E"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#D7827E"), // bright cyan - active
	NodeHot:    lipgloss.Color("#B4637A"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EA9D34"), // normal yellow
	Memory: lipgloss.Color("#286983"), // normal green

	// State.
	Success: lipgloss.Color("#286983"),
	Warning: lipgloss.Color("#EA9D34"),
	Error:   lipgloss.Color("#B4637A"),
}

// Seoul256 / dark.
var Seoul256Dark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#ADD4FB"), // bright blue - main accent
	Secondary: lipgloss.Color("#FFAFAF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#D0D0D0"), // foreground
	TextMuted: lipgloss.Color("#626262"), // bright black - muted text
	Border:    lipgloss.Color("#4E4E4E"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#87AFAF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#87D7D7"), // bright cyan - active
	NodeHot:    lipgloss.Color("#D75F87"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#D8AF5F"), // normal yellow
	Memory: lipgloss.Color("#5F865F"), // normal green

	// State.
	Success: lipgloss.Color("#87AF87"),
	Warning: lipgloss.Color("#FFD787"),
	Error:   lipgloss.Color("#D75F87"),
}

// Seoul256 / light.
var Seoul256Light = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#0087AF"), // bright blue - main accent
	Secondary: lipgloss.Color("#87025F"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#4E4E4E"), // foreground
	TextMuted: lipgloss.Color("#3A3A3A"), // bright black - muted text
	Border:    lipgloss.Color("#4E4E4E"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#5F8787"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#008787"), // bright cyan - active
	NodeHot:    lipgloss.Color("#870100"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#AF8760"), // normal yellow
	Memory: lipgloss.Color("#5F885F"), // normal green

	// State.
	Success: lipgloss.Color("#005F00"),
	Warning: lipgloss.Color("#D8865F"),
	Error:   lipgloss.Color("#870100"),
}

// Shades of Purple / default.
var ShadesOfPurpleDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#6943FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#FB94FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#FFFFFF"), // foreground
	TextMuted: lipgloss.Color("#5C5C61"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#80FCFF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#80FCFF"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E43937"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FAD000"), // normal yellow
	Memory: lipgloss.Color("#3AD900"), // normal green

	// State.
	Success: lipgloss.Color("#3AD900"),
	Warning: lipgloss.Color("#FAD000"),
	Error:   lipgloss.Color("#E43937"),
}

// Shades of Purple / super-dark.
var ShadesOfPurpleSuperDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#6943FF"), // bright blue - main accent
	Secondary: lipgloss.Color("#FB94FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#FFFFFF"), // foreground
	TextMuted: lipgloss.Color("#5C5C61"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#80FCFF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#80FCFF"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E33937"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FAD000"), // normal yellow
	Memory: lipgloss.Color("#3AD900"), // normal green

	// State.
	Success: lipgloss.Color("#3AD900"),
	Warning: lipgloss.Color("#FAD000"),
	Error:   lipgloss.Color("#E33937"),
}

// Solarized / dark.
var SolarizedDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#839496"), // bright blue - main accent
	Secondary: lipgloss.Color("#6C71C4"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#839496"), // foreground
	TextMuted: lipgloss.Color("#002B36"), // bright black - muted text
	Border:    lipgloss.Color("#073642"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#2AA198"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#93A1A1"), // bright cyan - active
	NodeHot:    lipgloss.Color("#CB4B16"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#B58900"), // normal yellow
	Memory: lipgloss.Color("#859900"), // normal green

	// State.
	Success: lipgloss.Color("#586E75"),
	Warning: lipgloss.Color("#657B83"),
	Error:   lipgloss.Color("#CB4B16"),
}

// Solarized / light.
var SolarizedLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#839496"), // bright blue - main accent
	Secondary: lipgloss.Color("#6C71C4"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#657B83"), // foreground
	TextMuted: lipgloss.Color("#002B36"), // bright black - muted text
	Border:    lipgloss.Color("#073642"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#2AA198"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#93A1A1"), // bright cyan - active
	NodeHot:    lipgloss.Color("#CB4B16"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#B58900"), // normal yellow
	Memory: lipgloss.Color("#859900"), // normal green

	// State.
	Success: lipgloss.Color("#586E75"),
	Warning: lipgloss.Color("#657B83"),
	Error:   lipgloss.Color("#CB4B16"),
}

// Sonokai / default.
var SonokaiDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#76CCE0"), // bright blue - main accent
	Secondary: lipgloss.Color("#B39DF3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E2E2E3"), // foreground
	TextMuted: lipgloss.Color("#7F8490"), // bright black - muted text
	Border:    lipgloss.Color("#181819"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#F39660"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#F39660"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FC5D7C"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E7C664"), // normal yellow
	Memory: lipgloss.Color("#9ED072"), // normal green

	// State.
	Success: lipgloss.Color("#9ED072"),
	Warning: lipgloss.Color("#E7C664"),
	Error:   lipgloss.Color("#FC5D7C"),
}

// Sonokai / atlantis.
var SonokaiAtlantis = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#72CCE8"), // bright blue - main accent
	Secondary: lipgloss.Color("#BA9CF3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E1E3E4"), // foreground
	TextMuted: lipgloss.Color("#828A9A"), // bright black - muted text
	Border:    lipgloss.Color("#181A1C"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#F69C5E"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#F69C5E"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF6578"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EACB64"), // normal yellow
	Memory: lipgloss.Color("#9DD274"), // normal green

	// State.
	Success: lipgloss.Color("#9DD274"),
	Warning: lipgloss.Color("#EACB64"),
	Error:   lipgloss.Color("#FF6578"),
}

// Sonokai / andromeda.
var SonokaiAndromeda = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#6DCAE8"), // bright blue - main accent
	Secondary: lipgloss.Color("#BB97EE"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E1E3E4"), // foreground
	TextMuted: lipgloss.Color("#7E8294"), // bright black - muted text
	Border:    lipgloss.Color("#181A1C"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#F89860"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#F89860"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FB617E"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EDC763"), // normal yellow
	Memory: lipgloss.Color("#9ED06C"), // normal green

	// State.
	Success: lipgloss.Color("#9ED06C"),
	Warning: lipgloss.Color("#EDC763"),
	Error:   lipgloss.Color("#FB617E"),
}

// Sonokai / shusia.
var SonokaiShusia = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#7ACCD7"), // bright blue - main accent
	Secondary: lipgloss.Color("#AB9DF2"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E3E1E4"), // foreground
	TextMuted: lipgloss.Color("#848089"), // bright black - muted text
	Border:    lipgloss.Color("#1A181A"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#EF9062"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#EF9062"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F85E84"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E5C463"), // normal yellow
	Memory: lipgloss.Color("#9ECD6F"), // normal green

	// State.
	Success: lipgloss.Color("#9ECD6F"),
	Warning: lipgloss.Color("#E5C463"),
	Error:   lipgloss.Color("#F85E84"),
}

// Sonokai / maia.
var SonokaiMaia = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#78CEE9"), // bright blue - main accent
	Secondary: lipgloss.Color("#BAA0F8"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E1E2E3"), // foreground
	TextMuted: lipgloss.Color("#82878B"), // bright black - muted text
	Border:    lipgloss.Color("#1C1E1F"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#F3A96A"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#F3A96A"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F76C7C"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E3D367"), // normal yellow
	Memory: lipgloss.Color("#9CD57B"), // normal green

	// State.
	Success: lipgloss.Color("#9CD57B"),
	Warning: lipgloss.Color("#E3D367"),
	Error:   lipgloss.Color("#F76C7C"),
}

// Sonokai / espresso.
var SonokaiEspresso = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#81D0C9"), // bright blue - main accent
	Secondary: lipgloss.Color("#9FA0E1"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E4E3E1"), // foreground
	TextMuted: lipgloss.Color("#90817B"), // bright black - muted text
	Border:    lipgloss.Color("#1F1E1C"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#F08D71"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#F08D71"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F86882"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F0C66F"), // normal yellow
	Memory: lipgloss.Color("#A6CD77"), // normal green

	// State.
	Success: lipgloss.Color("#A6CD77"),
	Warning: lipgloss.Color("#F0C66F"),
	Error:   lipgloss.Color("#F86882"),
}

// Srcery / default.
var SrceryDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#68A8E4"), // bright blue - main accent
	Secondary: lipgloss.Color("#FF5C8F"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#FCE8C3"), // foreground
	TextMuted: lipgloss.Color("#918175"), // bright black - muted text
	Border:    lipgloss.Color("#1C1B19"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#0AAEB3"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#2BE4D0"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F75341"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FBB829"), // normal yellow
	Memory: lipgloss.Color("#519F50"), // normal green

	// State.
	Success: lipgloss.Color("#98BC37"),
	Warning: lipgloss.Color("#FED06E"),
	Error:   lipgloss.Color("#F75341"),
}

// Tender / default.
var TenderDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#B3DEEF"), // bright blue - main accent
	Secondary: lipgloss.Color("#D3B987"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#EEEEEE"), // foreground
	TextMuted: lipgloss.Color("#1D1D1D"), // bright black - muted text
	Border:    lipgloss.Color("#282828"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#73CEF4"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#73CEF4"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F43753"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FFC24B"), // normal yellow
	Memory: lipgloss.Color("#C9D05C"), // normal green

	// State.
	Success: lipgloss.Color("#C9D05C"),
	Warning: lipgloss.Color("#FFC24B"),
	Error:   lipgloss.Color("#F43753"),
}

// Tokyo Night / default.
var TokyoNightDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#7AA2F7"), // bright blue - main accent
	Secondary: lipgloss.Color("#BB9AF7"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C0CAF5"), // foreground
	TextMuted: lipgloss.Color("#414868"), // bright black - muted text
	Border:    lipgloss.Color("#15161E"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#7DCFFF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#7DCFFF"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F7768E"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E0AF68"), // normal yellow
	Memory: lipgloss.Color("#9ECE6A"), // normal green

	// State.
	Success: lipgloss.Color("#9ECE6A"),
	Warning: lipgloss.Color("#E0AF68"),
	Error:   lipgloss.Color("#F7768E"),
}

// Tokyo Night / day.
var TokyoNightDay = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#2E7DE9"), // bright blue - main accent
	Secondary: lipgloss.Color("#9854F1"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#3760BF"), // foreground
	TextMuted: lipgloss.Color("#A1A6C5"), // bright black - muted text
	Border:    lipgloss.Color("#B4B5B9"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#007197"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#007197"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F52A65"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#8C6C3E"), // normal yellow
	Memory: lipgloss.Color("#587539"), // normal green

	// State.
	Success: lipgloss.Color("#587539"),
	Warning: lipgloss.Color("#8C6C3E"),
	Error:   lipgloss.Color("#F52A65"),
}

// Tokyo Night / storm.
var TokyoNightStorm = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#7AA2F7"), // bright blue - main accent
	Secondary: lipgloss.Color("#BB9AF7"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C0CAF5"), // foreground
	TextMuted: lipgloss.Color("#414868"), // bright black - muted text
	Border:    lipgloss.Color("#1D202F"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#7DCFFF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#7DCFFF"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F7768E"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E0AF68"), // normal yellow
	Memory: lipgloss.Color("#9ECE6A"), // normal green

	// State.
	Success: lipgloss.Color("#9ECE6A"),
	Warning: lipgloss.Color("#E0AF68"),
	Error:   lipgloss.Color("#F7768E"),
}

// Tokyo Night / moon.
var TokyoNightMoon = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#82AAFF"), // bright blue - main accent
	Secondary: lipgloss.Color("#C099FF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C8D3F5"), // foreground
	TextMuted: lipgloss.Color("#444A73"), // bright black - muted text
	Border:    lipgloss.Color("#1B1D2B"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#86E1FC"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#86E1FC"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF757F"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FFC777"), // normal yellow
	Memory: lipgloss.Color("#C3E88D"), // normal green

	// State.
	Success: lipgloss.Color("#C3E88D"),
	Warning: lipgloss.Color("#FFC777"),
	Error:   lipgloss.Color("#FF757F"),
}

// Tomorrow / night.
var TomorrowNight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#81A2BE"), // bright blue - main accent
	Secondary: lipgloss.Color("#B294BB"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C5C8C6"), // foreground
	TextMuted: lipgloss.Color("#000000"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#8ABEB7"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#8ABEB7"), // bright cyan - active
	NodeHot:    lipgloss.Color("#CC6666"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F0C674"), // normal yellow
	Memory: lipgloss.Color("#B5BD68"), // normal green

	// State.
	Success: lipgloss.Color("#B5BD68"),
	Warning: lipgloss.Color("#F0C674"),
	Error:   lipgloss.Color("#CC6666"),
}

// Tomorrow / default.
var TomorrowDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#4271AE"), // bright blue - main accent
	Secondary: lipgloss.Color("#8959A8"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#4D4D4C"), // foreground
	TextMuted: lipgloss.Color("#000000"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#3E999F"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#3E999F"), // bright cyan - active
	NodeHot:    lipgloss.Color("#C82829"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EAB700"), // normal yellow
	Memory: lipgloss.Color("#718C00"), // normal green

	// State.
	Success: lipgloss.Color("#718C00"),
	Warning: lipgloss.Color("#EAB700"),
	Error:   lipgloss.Color("#C82829"),
}

// Tomorrow / night-eighties.
var TomorrowNightEighties = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#6699CC"), // bright blue - main accent
	Secondary: lipgloss.Color("#CC99CC"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#CCCCCC"), // foreground
	TextMuted: lipgloss.Color("#000000"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#66CCCC"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#66CCCC"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F2777A"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FFCC66"), // normal yellow
	Memory: lipgloss.Color("#99CC99"), // normal green

	// State.
	Success: lipgloss.Color("#99CC99"),
	Warning: lipgloss.Color("#FFCC66"),
	Error:   lipgloss.Color("#F2777A"),
}

// Tomorrow / night-blue.
var TomorrowNightBlue = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#BBDAFF"), // bright blue - main accent
	Secondary: lipgloss.Color("#EBBBFF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#FFFFFF"), // foreground
	TextMuted: lipgloss.Color("#000000"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#99FFFF"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#99FFFF"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF9DA4"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FFEEAD"), // normal yellow
	Memory: lipgloss.Color("#D1F1A9"), // normal green

	// State.
	Success: lipgloss.Color("#D1F1A9"),
	Warning: lipgloss.Color("#FFEEAD"),
	Error:   lipgloss.Color("#FF9DA4"),
}

// Tomorrow / night-bright.
var TomorrowNightBright = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#7AA6DA"), // bright blue - main accent
	Secondary: lipgloss.Color("#C397D8"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#EAEAEA"), // foreground
	TextMuted: lipgloss.Color("#000000"), // bright black - muted text
	Border:    lipgloss.Color("#000000"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#70C0B1"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#70C0B1"), // bright cyan - active
	NodeHot:    lipgloss.Color("#D54E53"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E7C547"), // normal yellow
	Memory: lipgloss.Color("#B9CA4A"), // normal green

	// State.
	Success: lipgloss.Color("#B9CA4A"),
	Warning: lipgloss.Color("#E7C547"),
	Error:   lipgloss.Color("#D54E53"),
}

// Zenbones / zenwritten-dark.
var ZenbonesZenwrittenDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#61ABDA"), // bright blue - main accent
	Secondary: lipgloss.Color("#CF86C1"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#BBBBBB"), // foreground
	TextMuted: lipgloss.Color("#3D3839"), // bright black - muted text
	Border:    lipgloss.Color("#191919"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#66A5AD"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#65B8C1"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E8838F"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#B77E64"), // normal yellow
	Memory: lipgloss.Color("#819B69"), // normal green

	// State.
	Success: lipgloss.Color("#8BAE68"),
	Warning: lipgloss.Color("#D68C67"),
	Error:   lipgloss.Color("#E8838F"),
}

// Zenbones / zenwritten-light.
var ZenbonesZenwrittenLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#1D5573"), // bright blue - main accent
	Secondary: lipgloss.Color("#7B3B70"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#353535"), // foreground
	TextMuted: lipgloss.Color("#C6C3C3"), // bright black - muted text
	Border:    lipgloss.Color("#EEEEEE"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#3B8992"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#2B747C"), // bright cyan - active
	NodeHot:    lipgloss.Color("#94253E"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#944927"), // normal yellow
	Memory: lipgloss.Color("#4F6C31"), // normal green

	// State.
	Success: lipgloss.Color("#3F5A22"),
	Warning: lipgloss.Color("#803D1C"),
	Error:   lipgloss.Color("#94253E"),
}

// Zenbones / neobones-dark.
var ZenbonesNeobonesDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#92A0E2"), // bright blue - main accent
	Secondary: lipgloss.Color("#CF86C1"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C6D5CF"), // foreground
	TextMuted: lipgloss.Color("#263945"), // bright black - muted text
	Border:    lipgloss.Color("#0F191F"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#66A5AD"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#65B8C1"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E8838F"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#B77E64"), // normal yellow
	Memory: lipgloss.Color("#90FF6B"), // normal green

	// State.
	Success: lipgloss.Color("#A0FF85"),
	Warning: lipgloss.Color("#D68C67"),
	Error:   lipgloss.Color("#E8838F"),
}

// Zenbones / neobones-light.
var ZenbonesNeobonesLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#1D5573"), // bright blue - main accent
	Secondary: lipgloss.Color("#7B3B70"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#202E18"), // foreground
	TextMuted: lipgloss.Color("#B3C6B6"), // bright black - muted text
	Border:    lipgloss.Color("#E5EDE6"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#3B8992"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#2B747C"), // bright cyan - active
	NodeHot:    lipgloss.Color("#94253E"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#944927"), // normal yellow
	Memory: lipgloss.Color("#567A30"), // normal green

	// State.
	Success: lipgloss.Color("#3F5A22"),
	Warning: lipgloss.Color("#803D1C"),
	Error:   lipgloss.Color("#94253E"),
}

// Zenbones / vimbones.
var ZenbonesVimbones = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#1D5573"), // bright blue - main accent
	Secondary: lipgloss.Color("#7B3B70"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#353535"), // foreground
	TextMuted: lipgloss.Color("#C6C6A3"), // bright black - muted text
	Border:    lipgloss.Color("#F0F0CA"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#3B8992"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#2B747C"), // bright cyan - active
	NodeHot:    lipgloss.Color("#94253E"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#944927"), // normal yellow
	Memory: lipgloss.Color("#4F6C31"), // normal green

	// State.
	Success: lipgloss.Color("#3F5A22"),
	Warning: lipgloss.Color("#803D1C"),
	Error:   lipgloss.Color("#94253E"),
}

// Zenbones / rosebones-dark.
var ZenbonesRosebonesDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#94DAE6"), // bright blue - main accent
	Secondary: lipgloss.Color("#CEB3EF"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E1D4D4"), // foreground
	TextMuted: lipgloss.Color("#3A3651"), // bright black - muted text
	Border:    lipgloss.Color("#1A1825"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#9CCFD8"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#94DAE6"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F289A4"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F6C074"), // normal yellow
	Memory: lipgloss.Color("#317490"), // normal green

	// State.
	Success: lipgloss.Color("#358DAF"),
	Warning: lipgloss.Color("#F9CA8E"),
	Error:   lipgloss.Color("#F289A4"),
}

// Zenbones / rosebones-light.
var ZenbonesRosebonesLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#407D88"), // bright blue - main accent
	Secondary: lipgloss.Color("#855AAC"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#724341"), // foreground
	TextMuted: lipgloss.Color("#E8C48B"), // bright black - muted text
	Border:    lipgloss.Color("#FBF6F0"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#5795A0"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#407D88"), // bright cyan - active
	NodeHot:    lipgloss.Color("#A54A66"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EC9D33"), // normal yellow
	Memory: lipgloss.Color("#286A84"), // normal green

	// State.
	Success: lipgloss.Color("#1C5970"),
	Warning: lipgloss.Color("#C68223"),
	Error:   lipgloss.Color("#A54A66"),
}

// Zenbones / forestbones-dark.
var ZenbonesForestbonesDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#7AC9C0"), // bright blue - main accent
	Secondary: lipgloss.Color("#E5A7C4"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#E7DCC4"), // foreground
	TextMuted: lipgloss.Color("#45525C"), // bright black - muted text
	Border:    lipgloss.Color("#2C343A"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#83C193"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#7DD093"), // bright cyan - active
	NodeHot:    lipgloss.Color("#ED9294"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#DDBD7F"), // normal yellow
	Memory: lipgloss.Color("#A9C181"), // normal green

	// State.
	Success: lipgloss.Color("#B0CE7B"),
	Warning: lipgloss.Color("#EDC77A"),
	Error:   lipgloss.Color("#ED9294"),
}

// Zenbones / forestbones-light.
var ZenbonesForestbonesLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#297CA6"), // bright blue - main accent
	Secondary: lipgloss.Color("#CA43A3"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#4F5B62"), // foreground
	TextMuted: lipgloss.Color("#DBC988"), // bright black - muted text
	Border:    lipgloss.Color("#FAF3E1"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#36A87E"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#258C67"), // bright cyan - active
	NodeHot:    lipgloss.Color("#E6271C"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#DEA000"), // normal yellow
	Memory: lipgloss.Color("#8DA200"), // normal green

	// State.
	Success: lipgloss.Color("#758700"),
	Warning: lipgloss.Color("#B98500"),
	Error:   lipgloss.Color("#E6271C"),
}

// Zenbones / nordbones.
var ZenbonesNordbones = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#89CAC8"), // bright blue - main accent
	Secondary: lipgloss.Color("#CF97C5"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#EBEEF3"), // foreground
	TextMuted: lipgloss.Color("#475063"), // bright black - muted text
	Border:    lipgloss.Color("#2F3541"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#87BFCE"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#82CCE0"), // bright cyan - active
	NodeHot:    lipgloss.Color("#D6787F"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#CF866F"), // normal yellow
	Memory: lipgloss.Color("#A4BE8D"), // normal green

	// State.
	Success: lipgloss.Color("#A8CC86"),
	Warning: lipgloss.Color("#E09680"),
	Error:   lipgloss.Color("#D6787F"),
}

// Zenbones / tokyobones-dark.
var ZenbonesTokyobonesDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#90AFFA"), // bright blue - main accent
	Secondary: lipgloss.Color("#C6ACFA"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#C0CAF5"), // foreground
	TextMuted: lipgloss.Color("#36384D"), // bright black - muted text
	Border:    lipgloss.Color("#1A1B26"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#2BC4DE"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#74DBCB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#F98EA0"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E1B068"), // normal yellow
	Memory: lipgloss.Color("#74DBCB"), // normal green

	// State.
	Success: lipgloss.Color("#6DE5D3"),
	Warning: lipgloss.Color("#F2BA64"),
	Error:   lipgloss.Color("#F98EA0"),
}

// Zenbones / tokyobones-light.
var ZenbonesTokyobonesLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#26467A"), // bright blue - main accent
	Secondary: lipgloss.Color("#503875"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#333A57"), // foreground
	TextMuted: lipgloss.Color("#ADB0BD"), // bright black - muted text
	Border:    lipgloss.Color("#D6D7DC"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#176775"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#34645D"), // bright cyan - active
	NodeHot:    lipgloss.Color("#7E3242"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#8F5E14"), // normal yellow
	Memory: lipgloss.Color("#34645D"), // normal green

	// State.
	Success: lipgloss.Color("#26554F"),
	Warning: lipgloss.Color("#794E0D"),
	Error:   lipgloss.Color("#7E3242"),
}

// Zenbones / seoulbones-dark.
var ZenbonesSeoulbonesDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#A2C8E9"), // bright blue - main accent
	Secondary: lipgloss.Color("#B2B3DA"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#DDDDDD"), // foreground
	TextMuted: lipgloss.Color("#6C6465"), // bright black - muted text
	Border:    lipgloss.Color("#4B4B4B"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#6FBDBE"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#6BCACB"), // bright cyan - active
	NodeHot:    lipgloss.Color("#EB99B1"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#FFDF9B"), // normal yellow
	Memory: lipgloss.Color("#98BD99"), // normal green

	// State.
	Success: lipgloss.Color("#8FCD92"),
	Warning: lipgloss.Color("#FFE5B3"),
	Error:   lipgloss.Color("#EB99B1"),
}

// Zenbones / seoulbones-light.
var ZenbonesSeoulbonesLight = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#006F89"), // bright blue - main accent
	Secondary: lipgloss.Color("#7F4C7E"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#555555"), // foreground
	TextMuted: lipgloss.Color("#BFBABB"), // bright black - muted text
	Border:    lipgloss.Color("#E2E2E2"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#008586"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#006F70"), // bright cyan - active
	NodeHot:    lipgloss.Color("#BE3C6D"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#C48562"), // normal yellow
	Memory: lipgloss.Color("#628562"), // normal green

	// State.
	Success: lipgloss.Color("#487249"),
	Warning: lipgloss.Color("#A76B48"),
	Error:   lipgloss.Color("#BE3C6D"),
}

// Zenbones / duckbones.
var ZenbonesDuckbones = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#00B4E0"), // bright blue - main accent
	Secondary: lipgloss.Color("#B3A1E6"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#EBEFC0"), // foreground
	TextMuted: lipgloss.Color("#2B2F46"), // bright black - muted text
	Border:    lipgloss.Color("#0E101A"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#00A3CB"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#00B4E0"), // bright cyan - active
	NodeHot:    lipgloss.Color("#FF4821"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E39500"), // normal yellow
	Memory: lipgloss.Color("#5DCD97"), // normal green

	// State.
	Success: lipgloss.Color("#58DB9E"),
	Warning: lipgloss.Color("#F6A100"),
	Error:   lipgloss.Color("#FF4821"),
}

// Zenbones / zenburned.
var ZenbonesZenburned = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#61ABDA"), // bright blue - main accent
	Secondary: lipgloss.Color("#CF86C1"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#F0E4CF"), // foreground
	TextMuted: lipgloss.Color("#625A5B"), // bright black - muted text
	Border:    lipgloss.Color("#404040"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#66A5AD"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#65B8C1"), // bright cyan - active
	NodeHot:    lipgloss.Color("#EC8685"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#B77E64"), // normal yellow
	Memory: lipgloss.Color("#819B69"), // normal green

	// State.
	Success: lipgloss.Color("#8BAE68"),
	Warning: lipgloss.Color("#D68C67"),
	Error:   lipgloss.Color("#EC8685"),
}

// Zenbones / kanagawabones.
var ZenbonesKanagawabones = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#7BC2DF"), // bright blue - main accent
	Secondary: lipgloss.Color("#A98FD2"), // bright magenta - secondary accent

	Text:      lipgloss.Color("#DDD8BB"), // foreground
	TextMuted: lipgloss.Color("#3C3C51"), // bright black - muted text
	Border:    lipgloss.Color("#1F1F28"), // normal black - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#7EB3C9"), // normal cyan - dormant
	NodeActive: lipgloss.Color("#7BC2DF"), // bright cyan - active
	NodeHot:    lipgloss.Color("#EC818C"), // bright red - hot

	// Profiling.
	CPU:    lipgloss.Color("#E5C283"), // normal yellow
	Memory: lipgloss.Color("#98BC6D"), // normal green

	// State.
	Success: lipgloss.Color("#9EC967"),
	Warning: lipgloss.Color("#F1C982"),
	Error:   lipgloss.Color("#EC818C"),
}

//
// Hand-curated themes.
//
// The schemes below are popular editor themes that terminalcolors.com does
// not publish. Their colors come from each project's official palette (see
// the source references), choosing the color that best fits every semantic
// field instead of an ANSI-to-Theme transpose.

// Cyberdream / default — scottmckendry/cyberdream.nvim.
var CyberdreamDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#5EA1FF"), // blue - main accent
	Secondary: lipgloss.Color("#BD5EFF"), // purple - secondary accent

	Text:      lipgloss.Color("#FFFFFF"), // foreground
	TextMuted: lipgloss.Color("#7B8496"), // grey - muted text
	Border:    lipgloss.Color("#1E2124"), // bg_alt - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#5EF1FF"), // cyan - dormant
	NodeActive: lipgloss.Color("#5EA1FF"), // blue - active
	NodeHot:    lipgloss.Color("#FF6E5E"), // red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F1FF5E"), // yellow
	Memory: lipgloss.Color("#5EFF6C"), // green

	// State.
	Success: lipgloss.Color("#5EFF6C"),
	Warning: lipgloss.Color("#FFBD5E"),
	Error:   lipgloss.Color("#FF6E5E"),
}

// Bamboo / vulgaris — ribru17/bamboo.nvim.
var BambooVulgaris = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#57A5E5"), // blue - main accent
	Secondary: lipgloss.Color("#AAAAFF"), // purple - secondary accent

	Text:      lipgloss.Color("#F1E9D2"), // foreground
	TextMuted: lipgloss.Color("#5B5E5A"), // grey - muted text
	Border:    lipgloss.Color("#1C1E1B"), // bg_dark - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#70C2BE"), // cyan - dormant
	NodeActive: lipgloss.Color("#96C7EF"), // light blue - active
	NodeHot:    lipgloss.Color("#E75A7C"), // red - hot

	// Profiling.
	CPU:    lipgloss.Color("#DBB651"), // yellow
	Memory: lipgloss.Color("#8FB573"), // green

	// State.
	Success: lipgloss.Color("#8FB573"),
	Warning: lipgloss.Color("#FF9966"),
	Error:   lipgloss.Color("#E75A7C"),
}

// Melange / dark — savq/melange-nvim.
var MelangeDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#A3A9CE"), // blue - main accent
	Secondary: lipgloss.Color("#CF9BC2"), // magenta - secondary accent

	Text:      lipgloss.Color("#ECE1D7"), // foreground
	TextMuted: lipgloss.Color("#867462"), // muted tan - muted text
	Border:    lipgloss.Color("#34302C"), // float bg - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#7B9695"), // cyan - dormant
	NodeActive: lipgloss.Color("#89B3B6"), // bright cyan - active
	NodeHot:    lipgloss.Color("#D47766"), // red - hot

	// Profiling.
	CPU:    lipgloss.Color("#EBC06D"), // yellow
	Memory: lipgloss.Color("#85B695"), // green

	// State.
	Success: lipgloss.Color("#85B695"),
	Warning: lipgloss.Color("#EBC06D"),
	Error:   lipgloss.Color("#D47766"),
}

// TokyoDark / default — tiagovla/tokyodark.nvim.
var TokyoDarkDefault = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#7199EE"), // blue - main accent
	Secondary: lipgloss.Color("#A485DD"), // purple - secondary accent

	Text:      lipgloss.Color("#A0A8CD"), // foreground
	TextMuted: lipgloss.Color("#4A5057"), // comment grey - muted text
	Border:    lipgloss.Color("#212234"), // bg2 - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#38A89D"), // cyan - dormant
	NodeActive: lipgloss.Color("#7199EE"), // blue - active
	NodeHot:    lipgloss.Color("#EE6D85"), // red - hot

	// Profiling.
	CPU:    lipgloss.Color("#F6955B"), // orange
	Memory: lipgloss.Color("#95C561"), // green

	// State.
	Success: lipgloss.Color("#95C561"),
	Warning: lipgloss.Color("#D7A65F"),
	Error:   lipgloss.Color("#EE6D85"),
}

// VSCode Dark+ — Mofiqul/vscode.nvim (VS Code's built-in Dark+ theme).
var VSCodeDarkPlus = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#569CD6"), // blue - main accent
	Secondary: lipgloss.Color("#C586C0"), // purple - secondary accent

	Text:      lipgloss.Color("#D4D4D4"), // foreground
	TextMuted: lipgloss.Color("#858585"), // grey - muted text
	Border:    lipgloss.Color("#2D2D2D"), // widget border - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#4EC9B0"), // teal - dormant
	NodeActive: lipgloss.Color("#9CDCFE"), // light blue - active
	NodeHot:    lipgloss.Color("#F44747"), // red - hot

	// Profiling.
	CPU:    lipgloss.Color("#CE9178"), // orange
	Memory: lipgloss.Color("#6A9955"), // green

	// State.
	Success: lipgloss.Color("#6A9955"),
	Warning: lipgloss.Color("#DCDCAA"),
	Error:   lipgloss.Color("#F44747"),
}

// Oxocarbon / dark — nyoom-engineering/oxocarbon.nvim (IBM Carbon palette).
var OxocarbonDark = Theme{
	// Semantic reference colors.
	Primary:   lipgloss.Color("#78A9FF"), // blue - main accent
	Secondary: lipgloss.Color("#BE95FF"), // violet - secondary accent

	Text:      lipgloss.Color("#D5D5D5"), // foreground
	TextMuted: lipgloss.Color("#5C5C5C"), // grey - muted text
	Border:    lipgloss.Color("#2A2A2A"), // base01 - subtle separators

	// Runtime graph nodes.
	NodeIdle:   lipgloss.Color("#3DDBD9"), // cyan - dormant
	NodeActive: lipgloss.Color("#82CFFF"), // bright blue - active
	NodeHot:    lipgloss.Color("#FF7EB6"), // pink - hot

	// Profiling.
	CPU:    lipgloss.Color("#F1C21B"), // gold
	Memory: lipgloss.Color("#42BE65"), // green

	// State.
	Success: lipgloss.Color("#42BE65"),
	Warning: lipgloss.Color("#F1C21B"),
	Error:   lipgloss.Color("#FA4D56"),
}

// Themes registers every available theme by its slug.
var Themes = map[string]Theme{
	"apprentice-default":          ApprenticeDefault,
	"ayu-dark":                    AyuDark,
	"ayu-light":                   AyuLight,
	"ayu-mirage":                  AyuMirage,
	"bamboo-vulgaris":             BambooVulgaris,
	"catppuccin-mocha":            CatppuccinMocha,
	"catppuccin-macchiato":        CatppuccinMacchiato,
	"catppuccin-frappe":           CatppuccinFrappe,
	"catppuccin-latte":            CatppuccinLatte,
	"cobalt2-default":             Cobalt2Default,
	"cyberdream-default":          CyberdreamDefault,
	"deus-default":                DeusDefault,
	"dracula-default":             DraculaDefault,
	"dracula-soft":                DraculaSoft,
	"everforest-dark":             EverforestDark,
	"everforest-light":            EverforestLight,
	"github-dark":                 GithubDark,
	"github-dark-dimmed":          GithubDarkDimmed,
	"github-dark-high-contrast":   GithubDarkHighContrast,
	"github-dark-colorblind":      GithubDarkColorblind,
	"github-dark-legacy":          GithubDarkLegacy,
	"github-light":                GithubLight,
	"github-light-high-contrast":  GithubLightHighContrast,
	"github-light-colorblind":     GithubLightColorblind,
	"github-light-legacy":         GithubLightLegacy,
	"gotham-default":              GothamDefault,
	"gruvbox-dark":                GruvboxDark,
	"gruvbox-dark-hard-contrast":  GruvboxDarkHardContrast,
	"gruvbox-dark-soft-contrast":  GruvboxDarkSoftContrast,
	"gruvbox-light":               GruvboxLight,
	"gruvbox-light-hard-contrast": GruvboxLightHardContrast,
	"gruvbox-light-soft-contrast": GruvboxLightSoftContrast,
	"iceberg-dark":                IcebergDark,
	"iceberg-light":               IcebergLight,
	"jellybeans-default":          JellybeansDefault,
	"kanagawa-wave":               KanagawaWave,
	"kanagawa-dragon":             KanagawaDragon,
	"kanagawa-lotus":              KanagawaLotus,
	"lucario-default":             LucarioDefault,
	"miasma-default":              MiasmaDefault,
	"melange-dark":                MelangeDark,
	"moonfly-default":             MoonflyDefault,
	"night-owl-dark":              NightOwlDark,
	"night-owl-light":             NightOwlLight,
	"nightfly-default":            NightflyDefault,
	"nightfox-default":            NightfoxDefault,
	"nightfox-dayfox":             NightfoxDayfox,
	"nightfox-dawnfox":            NightfoxDawnfox,
	"nightfox-duskfox":            NightfoxDuskfox,
	"nightfox-nordfox":            NightfoxNordfox,
	"nightfox-terafox":            NightfoxTerafox,
	"nightfox-carbonfox":          NightfoxCarbonfox,
	"noctis-default":              NoctisDefault,
	"noctis-azureus":              NoctisAzureus,
	"noctis-bordo":                NoctisBordo,
	"noctis-minimus":              NoctisMinimus,
	"noctis-uva":                  NoctisUva,
	"noctis-viola":                NoctisViola,
	"noctis-lux":                  NoctisLux,
	"noctis-lilac":                NoctisLilac,
	"noctis-hibernus":             NoctisHibernus,
	"noctis-obscuro":              NoctisObscuro,
	"noctis-sereno":               NoctisSereno,
	"nord-default":                NordDefault,
	"nordic-default":              NordicDefault,
	"one-dark":                    OneDark,
	"one-light":                   OneLight,
	"one-half-dark":               OneHalfDark,
	"one-half-light":              OneHalfLight,
	"oxocarbon-dark":              OxocarbonDark,
	"panda-default":               PandaDefault,
	"posterpole-default":          PosterpoleDefault,
	"posterpole-gray":             PosterpoleGray,
	"rose-pine-default":           RosePineDefault,
	"rose-pine-moon":              RosePineMoon,
	"rose-pine-dawn":              RosePineDawn,
	"seoul256-dark":               Seoul256Dark,
	"seoul256-light":              Seoul256Light,
	"shades-of-purple-default":    ShadesOfPurpleDefault,
	"shades-of-purple-super-dark": ShadesOfPurpleSuperDark,
	"solarized-dark":              SolarizedDark,
	"solarized-light":             SolarizedLight,
	"sonokai-default":             SonokaiDefault,
	"sonokai-atlantis":            SonokaiAtlantis,
	"sonokai-andromeda":           SonokaiAndromeda,
	"sonokai-shusia":              SonokaiShusia,
	"sonokai-maia":                SonokaiMaia,
	"sonokai-espresso":            SonokaiEspresso,
	"srcery-default":              SrceryDefault,
	"tender-default":              TenderDefault,
	"tokyodark-default":           TokyoDarkDefault,
	"tokyo-night-default":         TokyoNightDefault,
	"tokyo-night-day":             TokyoNightDay,
	"tokyo-night-storm":           TokyoNightStorm,
	"tokyo-night-moon":            TokyoNightMoon,
	"tomorrow-night":              TomorrowNight,
	"tomorrow-default":            TomorrowDefault,
	"tomorrow-night-eighties":     TomorrowNightEighties,
	"tomorrow-night-blue":         TomorrowNightBlue,
	"tomorrow-night-bright":       TomorrowNightBright,
	"vscode-dark-plus":            VSCodeDarkPlus,
	"zenbones-zenwritten-dark":    ZenbonesZenwrittenDark,
	"zenbones-zenwritten-light":   ZenbonesZenwrittenLight,
	"zenbones-neobones-dark":      ZenbonesNeobonesDark,
	"zenbones-neobones-light":     ZenbonesNeobonesLight,
	"zenbones-vimbones":           ZenbonesVimbones,
	"zenbones-rosebones-dark":     ZenbonesRosebonesDark,
	"zenbones-rosebones-light":    ZenbonesRosebonesLight,
	"zenbones-forestbones-dark":   ZenbonesForestbonesDark,
	"zenbones-forestbones-light":  ZenbonesForestbonesLight,
	"zenbones-nordbones":          ZenbonesNordbones,
	"zenbones-tokyobones-dark":    ZenbonesTokyobonesDark,
	"zenbones-tokyobones-light":   ZenbonesTokyobonesLight,
	"zenbones-seoulbones-dark":    ZenbonesSeoulbonesDark,
	"zenbones-seoulbones-light":   ZenbonesSeoulbonesLight,
	"zenbones-duckbones":          ZenbonesDuckbones,
	"zenbones-zenburned":          ZenbonesZenburned,
	"zenbones-kanagawabones":      ZenbonesKanagawabones,
}
