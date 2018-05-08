package recipe

type Recipes struct {
	Recipes []Recipe `toml:"recipe"`
}

type Recipe struct {
	Time string `toml:"time"`
	// If it scales, then do not add time when changing volume
	Scales     bool      `toml:"scales"`
	Directions string    `toml:"directions"`
	Product    []Element `toml:"product"`
	Reactant   []Element `toml:"reactant"`
}

type Element struct {
	Name string  `toml:"name"`
	Cups float64 `toml:"cups"`
}

func Open(tomlfile string) (r Recipe, err error) {

	return
}
