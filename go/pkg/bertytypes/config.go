package bertytypes

// Config represents the configuration file defined in /config/config.yml
type Config struct {
	Berty struct {
		Contacts map[string]struct {
			Link        string `json:"link"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"contacts"`
		Conversations map[string]struct {
			Link        string `json:"link"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"conversations"`
	} `json:"berty"`
	P2P struct {
		RDVP []struct {
			Maddr string `json:"maddr"`
		} `json:"rdvp"`
	} `json:"p2p"`
}
