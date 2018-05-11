package recipe

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestOpen2(t *testing.T) {
	// collect all the possible reactions
	var r Reactions
	b, _ := ioutil.ReadFile("../recipes.toml")
	_, err := toml.Decode(string(b), &r)
	if err != nil {
		return
	}
	reactions := make(map[string]Reaction)
	for _, reaction := range r.Reactions {
		for _, product := range reaction.Product {
			if _, ok := reactions[product.Name]; ok {
				log.Printf("uh oh, already have %s", product.Name)
			} else {
				reactions[product.Name] = reaction
				reactions[product.Name].Product[0] = product
			}
		}
	}

	// get tree based on recipe and amount
	recipe := "chocolate chip cookies"
	log.Println(reactions[recipe])
	d := new(Dag)
	recipeToGet := reactions[recipe].Product[0]
	log.Println(reactions[recipe].Product[0])
	recursivelyAddRecipe(Element{
		Name:    recipeToGet.Name,
		Amount:  recipeToGet.Amount * 2,
		Measure: recipeToGet.Measure,
		Price:   recipeToGet.Price,
	}, d, reactions)

	// TODO: prune tree by time or price
	pruneTreeByTime(d, 0, 0)
	printDag(d)

	// parse tree for ingredients to build and the ingredients to buy
	ingredientsToBuild, ingredientsToBuy := getIngredientsToBuild(d, []Element{}, []Element{})
	fmt.Println("\nIngredients to build:")
	for _, ing := range ingredientsToBuild {
		fmt.Println("-", ing.Name, ing.Amount)
	}
	fmt.Println("\nIngredients to buy:")
	for _, ing := range ingredientsToBuy {
		fmt.Println("-", ing.Name, ing.Amount)
	}

	// collect the roots
	roots := getDagRoots(d, []*Dag{})
	rootMap := make(map[string]*Dag)
	for _, root := range roots {
		if _, ok := rootMap[root.Product.Name]; ok {
			log.Println(root.Product.Name)
		}
		rootMap[root.Product.Name] = root
	}

	// log.Println(pathExists(rootMap["cow milk"], rootMap["cow milk"])) // true
	// log.Println(pathExists(rootMap["milk"], rootMap["cow milk"]))     // true
	// log.Println(pathExists(rootMap["cow milk"], rootMap["milk"]))     // false

	// DETERMINE THE BEST ORDERING
	// find ingredients to build that don't depend on any ingredients to build
	ingredientsToBuildMap := make(map[string]struct{})
	for _, ing := range ingredientsToBuild {
		ingredientsToBuildMap[ing.Name] = struct{}{}
	}
	directionsOrder := []string{}
	for {
		if len(ingredientsToBuildMap) == 0 {
			break
		}
		thingsThatCanBeBuiltNow := make(map[string]struct{})
		for ing1 := range ingredientsToBuildMap {
			var ing1DependsOnIng2 bool
			for ing2 := range ingredientsToBuildMap {
				if ing1 == ing2 {
					continue
				}
				// make sure ing1 doesn't depend on ing2
				if pathExists(rootMap[ing1], rootMap[ing2]) {
					ing1DependsOnIng2 = true
					break
				}
			}
			if !ing1DependsOnIng2 {
				thingsThatCanBeBuiltNow[ing1] = struct{}{}
			}
		}
		if len(ingredientsToBuildMap) == 1 {
			for ing1 := range ingredientsToBuildMap {
				thingsThatCanBeBuiltNow[ing1] = struct{}{}
			}
		}

		// find the one that takes the longest
		longestTime := 0.0
		currentThing := ""
		for ing := range thingsThatCanBeBuiltNow {
			timeTaken := rootMap[ing].SerialHours + rootMap[ing].ParallelHours
			if timeTaken > longestTime {
				longestTime = timeTaken
				currentThing = ing
			}
		}

		directionsOrder = append(directionsOrder, currentThing)
		// delete it from things to build, and iterate
		delete(ingredientsToBuildMap, currentThing)
	}
	log.Println(directionsOrder)
	printDag(d)

}

type Dag struct {
	ParallelHours float64   `toml:"p_hours" json:"p_hours,omitempty"`
	SerialHours   float64   `toml:"s_hours" json:"s_hours,omitempty"`
	Directions    string    `toml:"directions" json:"directions,omitempty"`
	Notes         string    `toml:"notes" json:"notes,omitempty"`
	Product       Element   `toml:"product" json:"product,omitempty"`
	Reactant      []Element `toml:"reactant" json:"reactant,omitempty"`
	Children      []*Dag
}

func pruneTreeByTime(d *Dag, currentTime float64, maxTime float64) {
	currentTime += d.SerialHours + d.ParallelHours
	if currentTime > maxTime {
		d.Children = []*Dag{}
	} else {
		for _, child := range d.Children {
			pruneTreeByTime(child, currentTime, maxTime)
		}
	}
}

func printDag(d *Dag) {
	printDagRecursively(d, 0)
}

func printDagRecursively(d *Dag, in int) {
	for i := 0; i < in; i++ {
		fmt.Print("\t")
	}
	fmt.Println(d.Product.Name, d.Product.Amount, d.Product.Measure)
	for _, child := range d.Children {
		printDagRecursively(child, in+1)
	}
}

func pathExists(fromNode *Dag, toNode *Dag) bool {
	if fromNode.Product.Name == toNode.Product.Name {
		return true
	}
	for _, child := range fromNode.Children {
		if pathExists(child, toNode) {
			return true
		}
	}
	return false
}

func getDagRoots(d *Dag, roots []*Dag) []*Dag {
	roots = append(roots, d)
	for _, child := range d.Children {
		roots = getDagRoots(child, roots)
	}
	return roots
}

func getIngredientsToBuild(d *Dag, ingredientsToBuild []Element, ingredientsToBuy []Element) ([]Element, []Element) {
	if len(d.Children) == 0 {
		i := -1
		for j, e := range ingredientsToBuy {
			if e.Name == d.Product.Name {
				i = j
				break
			}
		}
		if i == -1 {
			ingredientsToBuy = append(ingredientsToBuy, d.Product)
		} else {
			ingredientsToBuy[i].Amount += d.Product.Amount
			ingredientsToBuy[i].Price += d.Product.Price
		}
		return ingredientsToBuild, ingredientsToBuy
	}
	i := -1
	for j, e := range ingredientsToBuild {
		if e.Name == d.Product.Name {
			i = j
			break
		}
	}
	if i == -1 {
		ingredientsToBuild = append(ingredientsToBuild, d.Product)
	} else {
		ingredientsToBuild[i].Amount += d.Product.Amount
		ingredientsToBuild[i].Price += d.Product.Price
	}
	for _, child := range d.Children {
		ingredientsToBuild, ingredientsToBuy = getIngredientsToBuild(child, ingredientsToBuild, ingredientsToBuy)

	}
	return ingredientsToBuild, ingredientsToBuy
}

func recursivelyAddRecipe(recipe Element, d *Dag, reactions map[string]Reaction) {
	// add basic element
	d.Product = Element{
		Name:    recipe.Name,
		Notes:   recipe.Notes,
		Measure: recipe.Measure,
		Amount:  recipe.Amount,
		Price:   recipe.Price,
	}
	d.Product = recipe
	// if recipe.Measure == "whole" {
	// 	d.Product.Amount = math.Ceil(d.Product.Amount)
	// }

	// add children, if any
	d.Children = []*Dag{}
	if _, ok := reactions[recipe.Name]; ok {
		// determine the scaling from the baseline reaction
		scaling := recipe.Amount / reactions[recipe.Name].Product[0].Amount
		log.Println("A:", recipe.Name, scaling, recipe.Amount, reactions[recipe.Name].Product[0].Amount)

		d.Directions = reactions[recipe.Name].Directions
		d.Notes = reactions[recipe.Name].Notes
		d.ParallelHours = reactions[recipe.Name].ParallelHours
		d.SerialHours = scaling * reactions[recipe.Name].SerialHours          // scale the time
		d.Product.Amount = scaling * reactions[recipe.Name].Product[0].Amount // scale the amount
		d.Product.Price = scaling * reactions[recipe.Name].Product[0].Price   // scale the price
		d.Reactant = make([]Element, len(reactions[recipe.Name].Reactant))
		for i, r := range reactions[recipe.Name].Reactant {
			d.Reactant[i].Amount = r.Amount * scaling // scale the amount
			// if d.Reactant[i].Measure == "whole" {
			// 	d.Reactant[i].Amount = math.Ceil(d.Reactant[i].Amount)
			// }
			d.Reactant[i].Price = r.Price * scaling // scale the price (though there shouldn't be one)
			d.Reactant[i].Measure = r.Measure
			d.Reactant[i].Name = r.Name
			d.Reactant[i].Notes = r.Notes
		}

		// add the reactants as children to the tree
		for _, child := range d.Reactant {
			d2 := new(Dag)
			recursivelyAddRecipe(child, d2, reactions)
			d.Children = append(d.Children, d2)
		}
	}
	return
}
