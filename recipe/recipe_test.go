package recipe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func TestOpen2(t *testing.T) {
	b, _ := ioutil.ReadFile("../recipes.toml")
	var r Reactions
	_, err := toml.Decode(string(b), &r)
	assert.Nil(t, err)

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

	// prune tree based on recipe + time
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

	// find ingredients to build that don't depend on any ingredients to build
	ingredientsToBuildMap := make(map[string]struct{})
	for _, ing := range ingredientsToBuild {
		ingredientsToBuildMap[ing.Name] = struct{}{}
	}

	for ii := 0; ii < 3; ii++ {

		thingsThatCanBeBuiltNow := make(map[string]struct{})
		for ing := range ingredientsToBuildMap {
			for _, reactant := range reactions[ing].Reactant {
				if _, ok := ingredientsToBuildMap[reactant.Name]; ok {
					continue
				}
				thingsThatCanBeBuiltNow[ing] = struct{}{}
			}
		}
		log.Println(thingsThatCanBeBuiltNow)

		// find the one that takes the longest
		roots := getDagRoots(d, []*Dag{})
		rootMap := make(map[string]*Dag)
		for _, root := range roots {
			if _, ok := rootMap[root.Product.Name]; ok {
				log.Println(root.Product.Name)
			}
			rootMap[root.Product.Name] = root
		}
		log.Println(rootMap)
		longestTime := 0.0
		currentThing := ""
		for ing := range thingsThatCanBeBuiltNow {
			timeTaken := rootMap[ing].SerialHours + rootMap[ing].ParallelHours
			if timeTaken > longestTime {
				longestTime = timeTaken
				currentThing = ing
				log.Println(longestTime, currentThing)
			}
		}
		log.Println(currentThing, "takes the longest")

		amountNeeded := 0.0
		for _, ing := range ingredientsToBuild {
			if ing.Name == currentThing {
				amountNeeded = ing.Amount
			}
		}
		log.Println("needs", amountNeeded)

		// scale up/down the reaction for the specified amount
		reactionAmount := reactions[currentThing].Product[0].Amount
		log.Println("reaction produces", reactionAmount)
		scaling := amountNeeded / reactionAmount
		log.Println("multiply base reaction by", scaling)
		// prepend the reaction for the one that takes the longest to the queue

		// delete it from things to build, and iterate
		delete(ingredientsToBuildMap, currentThing)

	}

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
	d.Product = Element{
		Name:    recipe.Name,
		Notes:   recipe.Notes,
		Measure: recipe.Measure,
		Amount:  recipe.Amount,
		Price:   recipe.Price,
	}
	d.Children = []*Dag{}
	d.Product = recipe
	if _, ok := reactions[recipe.Name]; ok {
		// determine the scaling from the baseline reaction
		scaling := recipe.Amount / reactions[recipe.Name].Product[0].Amount
		log.Println("A:", recipe.Name, scaling, recipe.Amount, reactions[recipe.Name].Product[0].Amount)

		d.Directions = reactions[recipe.Name].Directions
		d.Notes = reactions[recipe.Name].Notes
		d.ParallelHours = reactions[recipe.Name].ParallelHours
		d.SerialHours *= scaling * reactions[recipe.Name].SerialHours // scale the time
		d.Product.Amount = scaling * reactions[recipe.Name].Product[0].Amount
		d.Product.Price = scaling * reactions[recipe.Name].Product[0].Price
		d.Reactant = make([]Element, len(reactions[recipe.Name].Reactant))
		for i, r := range reactions[recipe.Name].Reactant {
			d.Reactant[i].Amount = r.Amount * scaling
			if d.Reactant[i].Measure == "whole" {
				d.Reactant[i].Amount = math.Ceil(d.Reactant[i].Amount)
			}
			d.Reactant[i].Price = r.Price * scaling
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

func TestOpen(t *testing.T) {
	b, _ := ioutil.ReadFile("../recipes.toml")
	var r Reactions
	_, err := toml.Decode(string(b), &r)
	assert.Nil(t, err)
	bJson, _ := json.MarshalIndent(r, "", " ")
	fmt.Println(string(bJson))

	var rd RecipeDag
	rd.Children = make(map[string][]string)
	rd.Node = make(map[string]Reaction)
	rd.All = make(map[string]struct{})
	for _, reaction := range r.Reactions {
		for _, product := range reaction.Product {
			reaction1 := reaction
			reaction1.Product = []Element{product}
			rd.Node[reaction1.Product[0].Name] = reaction1
			rd.All[reaction1.Product[0].Name] = struct{}{}
			rd.Children[reaction1.Product[0].Name] = make([]string, len(reaction1.Reactant))
			if reaction1.Product[0].Price == 0 {
				log.Printf("missing price for %s", reaction1.Product[0].Name)
			}
			for i, reactant := range reaction.Reactant {
				rd.Children[reaction1.Product[0].Name][i] = reactant.Name
				rd.All[reactant.Name] = struct{}{}
			}
		}
	}
	// bJson, _ = json.MarshalIndent(rd, "", " ")
	// // fmt.Println(string(bJson))
	for name := range rd.All {
		if _, ok := rd.Node[name]; !ok {
			log.Printf("missing recipe for %s\n", name)
		}
	}

	subset, elements, timelimit := rd.getSubset(32, "chocolate chip cookies", make(map[string][]string), make(map[string]Element), 50)
	fmt.Println(subset, timelimit)
	bJson, _ = json.MarshalIndent(elements, "", " ")
	fmt.Println(string(bJson))

	// generate graphviz
	s := "digraph G {\n"
	for node := range subset {
		for _, child := range subset[node] {
			s += fmt.Sprintf(`"%s" -> "%s"`+"\n", node, child)
		}
	}
	s += "}\n"
	fmt.Println(s)

	// get ingredients
	for node := range subset {
		if len(subset[node]) != 0 {
			// skip non-leafs
			continue
		}
		fmt.Println(node)
	}

	// list in order from most time to least time
	fmt.Println("\n\n\nordered")
	for node := range subset {
		for _, next := range subset[node] {
			if _, ok := subset[next]; !ok {
				continue
			}
			if len(subset[next]) == 0 {
				continue
			}
			if _, ok := rd.Node[next]; !ok {
				continue
			}
			fmt.Println(next, rd.Node[next].SerialHours)
		}
	}
	// fmt.Println(rd.subsetCost(subset))
}

// func (rd RecipeDag) subsetCost(subset map[string][]string) float64 {
// 	for node := range subset {
// 		for _, child := range subset[node] {
// 			if _, ok := subset[child]; !ok {
// 				// child is not a node, so it is a leaf
// 				log.Println(child)
// 			}
// 		}
// 	}
// 	return 0
// }

func (rd RecipeDag) getSubset(amount float64, node string, subset map[string][]string, elements map[string]Element, timelimit float64) (map[string][]string, map[string]Element, float64) {
	var scaling, timetaken float64
	subset[node] = []string{}
	if _, ok := rd.Node[node]; ok {
		scaling = amount / rd.Node[node].Product[0].Amount
		log.Printf("%s amount needed: %2.3f, default: %2.3f, scaling = %2.3f", node, amount, rd.Node[node].Product[0].Amount, scaling)
		timetaken = scaling*rd.Node[node].SerialHours + rd.Node[node].ParallelHours
		elements[node] = Element{
			Name:    rd.Node[node].Product[0].Name,
			Amount:  scaling * rd.Node[node].Product[0].Amount,
			Measure: rd.Node[node].Product[0].Measure,
			Price:   scaling * rd.Node[node].Product[0].Price,
		}
	}
	if len(rd.Children[node]) == 0 {
		// define the scaled ingredient
		return subset, elements, timelimit
	}
	if _, ok := rd.Node[node]; ok {
		if timelimit-timetaken < 0 {
			return subset, elements, timelimit
		}
		timelimit -= timetaken
	}
	// get the possible children of the current node
	subset[node] = rd.Children[node]

	for _, child := range rd.Children[node] {
		amountneeded := 0.0
		for _, reactant := range rd.Node[node].Reactant {
			if reactant.Name == child {
				amountneeded = reactant.Amount
				break
			}
		}
		log.Printf("for %s scaled %2.3f, needs %s amount needed: %2.3f", node, scaling, child, scaling*amountneeded)
		subset, elements, timelimit = rd.getSubset(scaling*amountneeded, child, subset, elements, timelimit)

		for _, reactant := range rd.Node[node].Reactant {
			if reactant.Name != child {
				continue
			}
			var price float64
			if _, ok := rd.Node[reactant.Name]; ok {
				price = rd.Node[reactant.Name].Product[0].Price / rd.Node[reactant.Name].Product[0].Amount * reactant.Amount
			}
			if _, ok := elements[reactant.Name]; !ok {
				elements[reactant.Name] = Element{
					Name:    reactant.Name,
					Amount:  scaling * reactant.Amount,
					Measure: reactant.Measure,
					Price:   scaling * price,
				}
			} else {
				elements[reactant.Name] = Element{
					Name:    reactant.Name,
					Amount:  scaling*reactant.Amount + elements[reactant.Name].Amount,
					Measure: reactant.Measure,
					Price:   scaling*price + elements[reactant.Name].Price,
				}
			}
			break

		}
	}
	return subset, elements, timelimit
}
