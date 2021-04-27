package scrape

type Config struct {
	Login    string `toml:"login"`
	Password string `toml:"password"`
	API      string `toml:"api"`
	WebSite  string `toml:"website"`
	Medium   string `toml:"medium"`
}

func NewConfig() *Config {
	return &Config{}
}
