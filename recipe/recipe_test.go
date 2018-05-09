package recipe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

	recipe := "chocolate chip cookies"
	log.Println(reactions[recipe])
	d := new(Dag)
	recursivelyAddRecipe(reactions[recipe].Product[0], d, reactions)
	log.Printf("%+v", d)
	for _, child := range d.Children {
		log.Println(child.Name, child.Measure, child.Amount)
	}
}

func recursivelyAddRecipe(recipe Element, d *Dag, reactions map[string]Reaction) {
	log.Printf("%s", recipe.Name)
	d.Reaction.Product = nil
	d.Reaction.Reactant = nil
	d.Children = []*Dag{}
	d.Element = recipe
	if _, ok := reactions[recipe.Name]; ok {
		d.Reaction = reactions[recipe.Name]
		for _, child := range reactions[recipe.Name].Reactant {
			d2 := new(Dag)
			recursivelyAddRecipe(child, d2, reactions)
			d.Children = append(d.Children, d2)
		}
	}
	return
}

type Dag struct {
	Element
	Reaction
	Children []*Dag
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
