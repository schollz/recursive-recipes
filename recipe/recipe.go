package recipe

type Recipes struct {
	Prices  []Price  `toml:"price"`
	Recipes []Recipe `toml:"recipe"`
}

type Price struct {
	Name         string  `tomlL"name"`
	BeforeMarket bool    `toml:"beforemarket"`
	Price        float64 `toml:"price"`
	Measure      string  `toml:"measure"`
	Year         int64   `toml:"year"`
	Notes        string  `toml:"notes"`
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
	Name string `toml:"name"`

	// Cups if it can be converted to volume
	Cups float64 `toml:"cups"`
	// Whole is to be used otherwise
	Whole float64 `toml:"whole"`

	// Acres if it is a farm product
	Acres float64 `toml:"acres"`

	// Notes are for references
	Notes string `toml:"notes"`
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
