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

	subset, elements, timelimit := rd.getSubset(8, "english muffin", make(map[string][]string), make(map[string]Element), 10000000)
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
	var scaling float64
	subset[node] = []string{}
	if len(rd.Children[node]) == 0 {
		// define the scaled ingredient
		return subset, elements, timelimit
	}
	if _, ok := rd.Node[node]; ok {
		scaling = amount / rd.Node[node].Product[0].Amount
		log.Printf("%s amount needed: %2.3f, default: %2.3f, scaling = %2.3f", node, amount, rd.Node[node].Product[0].Amount, scaling)

		timetaken := scaling*rd.Node[node].SerialHours + rd.Node[node].ParallelHours
		if timelimit-timetaken < 0 {
			elements[node] = Element{
				Name:    rd.Node[node].Product[0].Name,
				Amount:  scaling * rd.Node[node].Product[0].Amount,
				Measure: rd.Node[node].Product[0].Measure,
				Price:   scaling * rd.Node[node].Product[0].Price,
			}
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
		previousTimeLimit := timelimit
		subset, elements, timelimit = rd.getSubset(scaling*amountneeded, child, subset, elements, timelimit)
		if timelimit == previousTimeLimit {
			// did not add, this is a leaf
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
	}
	return subset, elements, timelimit
}
