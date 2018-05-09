package recipe

import "time"

type Recipes struct {
	Recipes []Recipe `toml:"recipe"`
}

// Recipe is the reaction that takes place, with reactants
// and products in either their theoretical OR practical
// lowest irreducible
type Recipe struct {
	// Hours it takes to create the products from reactants
	Hours float64 `toml:"hours" json:"hours,omitempty"`
	// If it scales, then do not add time when changing volume
	Scales     bool      `toml:"scales" json:"scales,omitempty"`
	Directions string    `toml:"directions" json:"directions,omitempty"`
	Notes      string    `toml:"notes" json:"notes,omitempty"`
	Product    []Element `toml:"product" json:"product,omitempty"`
	Reactant   []Element `toml:"reactant" json:"reactant,omitempty"`

	// LastUpdated is the year it was last updated
	// (refers to the price)
	LastUpdated time.Time `toml:"updated" json:"updated,omitempty"`
}

type Element struct {
	// Name is the name of the product/reactant
	Name string `toml:"name" json:"name,omitempty"`

	// Amount is the amount
	Amount float64 `toml:"amount" json:"amount,omitempty"`

	// Measure is the type of measurement, either
	// "grams" (weight), "cups" (volume), "acres" (sq ft),
	// or "whole"
	Measure string `toml:"measure" json:"measure,omitempty"`

	// Price is the cost per amount+measure
	Price float64 `toml:"price" json:"price,omitempty"`

	// Notes are for references
	Notes string `toml:"notes" json:"notes,omitempty"`
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
