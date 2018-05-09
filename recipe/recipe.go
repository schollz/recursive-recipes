package recipe

import "time"

type Reactions struct {
	Reactions []Reaction `toml:"reaction"`
}

// Reaction is the reaction that takes place, with reactants
// and products in either their theoretical OR practical
// lowest irreducible
type Reaction struct {
	// ParallelHours it takes to create the products from reactants, in parallel.
	// For example, trees grow at the same rate no matter how many there are,
	// so there time cost would be almost entirely "parallel". Parallel time
	// does not scale when scaling the recipe
	ParallelHours float64 `toml:"p_hours" json:"p_hours,omitempty"`

	// SerialHours is the serial amount of time needed, time that scales
	// proportional to the quantities.
	SerialHours float64 `toml:"s_hours" json:"s_hours,omitempty"`

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

	// Price is the cost per amount+measure, specified on products.
	Price float64 `toml:"price" json:"price,omitempty"`

	// Notes are for references
	Notes string `toml:"notes" json:"notes,omitempty"`
}

type RecipeDag struct {
	// Node maps the name of a recipe to the reaction, but the reaction
	// only contains a single product
	Node map[string]Reaction

	// Children maps the name of the ingredient to its children
	Children map[string][]string

	// All elements
	All map[string]struct{}
}

func Open(tomlfile string) (r Reaction, err error) {

	return
}
