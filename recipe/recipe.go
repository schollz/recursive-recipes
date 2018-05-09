package recipe

type Recipes struct {
	Recipes []Recipe `toml:"recipe"`
}

// Recipe is the reaction that takes place, with reactants
// and products in either their theoretical OR practical
// lowest irreducible
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
	// Name is the name of the product/reactant
	Name string `toml:"name"`

	// Amount is the amount
	Amount float64 `toml:"amount"`

	// Measure is the type of measurement, either
	// "grams" (weight), "cups" (volume), "acres" (sq ft),
	// or "whole"
	Measure string `toml:"measure"`

	// Price is the cost per amount+measure
	Price float64 `toml:"price"`

	// LastUpdated is the year it was last updated
	// (refers to the price)
	Year int64 `toml:"year"`

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
