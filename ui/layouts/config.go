package layouts

// FooterLink is a configurable link in the footer
type FooterLink struct {
	Label string
	URL   string
}

// PanelConfig contains the admin panel configuration
// Filament style - centralized configuration
type PanelConfig struct {
	Name              string    // Panel name (ex: "SublimeGo")
	Path              string    // Base path (ex: "/admin")
	Logo              string    // Logo URL (optional)
	Favicon           string    // Favicon URL (optional)
	PrimaryColor      string    // Primary color (default: green)
	DarkMode          bool      // Enable dark mode by default
	Registration      bool      // Enable registration
	EmailVerification bool      // Enable email verification
	PasswordReset     bool      // Enable password reset
	Profile           bool      // Enable profile page
	Navigation        []NavItem // Navigation items

	// Topbar features
	SearchBar         bool   // Show global search bar in topbar (default: true)
	Notifications     bool   // Show notifications bell in topbar (default: true)
	SearchPlaceholder string // Placeholder text for search bar

	// Footer
	FooterEnabled bool         // Show footer (default: true)
	FooterLinks   []FooterLink // Configurable footer links
}

// DefaultPanelConfig returns the default configuration
func DefaultPanelConfig() *PanelConfig {
	return &PanelConfig{
		Name:              "SublimeGo",
		Path:              "/admin",
		PrimaryColor:      "green",
		DarkMode:          false,
		Registration:      true,
		EmailVerification: false,
		PasswordReset:     true,
		Profile:           true,
		Navigation:        []NavItem{},
		SearchBar:         true,
		Notifications:     true,
		SearchPlaceholder: "Search...",
		FooterEnabled:     true,
		FooterLinks: []FooterLink{
			{Label: "Documentation", URL: "#"},
			{Label: "Support", URL: "#"},
		},
	}
}

// panelConfig stores the global panel configuration
var panelConfig = DefaultPanelConfig()

// SetPanelConfig sets the panel configuration
func SetPanelConfig(config *PanelConfig) {
	if config != nil {
		panelConfig = config
	}
}

// GetPanelConfig returns the current configuration
func GetPanelConfig() *PanelConfig {
	return panelConfig
}
