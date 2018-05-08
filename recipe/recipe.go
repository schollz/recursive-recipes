package recipe

type Recipes struct {
	Recipes []Recipe `toml:"recipe"`
}

type Recipe struct {
	// Hours it takes to create the products from reactants
	Hours float64 `toml:"hours"`
	// If it scales, then do not add time when changing volume
	Scales     bool      `toml:"scales"`
	Directions string    `toml:"directions"`
	Notes      string    `toml:"notes"`
	Product    []Element `toml:"product"`
	Reactant   []Element `toml:"reactant"`
}

type Element struct {
	Name  string  `toml:"name"`
	Cups  float64 `toml:"cups"`
	Acres float64 `toml:"acres"`
	Notes string  `toml:"notes"`
}

type SingleRecipe struct {
	Product     string
	Cups        float64
	Hours       float64
	Scales      bool
	Directions  string
	Ingredients []string
}

func Open(tomlfile string) (r Recipe, err error) {

	return
}
