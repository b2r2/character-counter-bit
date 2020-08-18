package scrape

// Config ...
type Config struct {
	Login    string `toml:"login"`
	Password string `toml:"password"`
	API      string `toml:"api"`
	WebSite  string `toml:"website"`

	MIToken       string `toml:"mIToken"`
	MClientID     string `toml:"mClientID"`
	MClientSecret string `toml:"mClientSecret"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{}
}
