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

	subset, timelimit := rd.getSubset(40, "chocolate chip cookies", make(map[string][]string), rd.Children, 40)
	fmt.Println(subset, timelimit)

	// generate graphviz
	s := "digraph G {\n"
	for node := range subset {
		for _, child := range subset[node] {
			s += fmt.Sprintf(`"%s" -> "%s"`+"\n", node, child)
		}
	}
	s += "}\n"
	fmt.Println(s)

	fmt.Println(rd.subsetCost(subset))
}

func (rd RecipeDag) subsetCost(subset map[string][]string) float64 {
	for node := range subset {
		for _, child := range subset[node] {
			if _, ok := subset[child]; !ok {
				// child is not a node, so it is a leaf
				log.Println(child)
			}
		}
	}
	return 0
}

func (rd RecipeDag) getSubset(amount float64, node string, subset map[string][]string, all map[string][]string, timelimit float64) (map[string][]string, float64) {
	if len(all[node]) == 0 {
		return subset, timelimit
	}
	scaling := amount / rd.Node[node].Product[0].Amount
	log.Printf("%s amount needed: %2.3f, default: %2.3f, scaling = %2.3f", node, amount, rd.Node[node].Product[0].Amount, scaling)
	if _, ok := rd.Node[node]; ok {
		timetaken := scaling*rd.Node[node].SerialHours + rd.Node[node].ParallelHours
		if timelimit-timetaken < 0 {
			return subset, timelimit
		}
		timelimit -= timetaken
	}
	subset[node] = all[node]
	for _, child := range all[node] {
		amountneeded := 0.0
		for _, reactant := range rd.Node[node].Reactant {
			if reactant.Name == child {
				amountneeded = reactant.Amount
				break
			}
		}
		log.Printf("for %s scaled %2.3f, needs %s amount needed: %2.3f", node, scaling, child, scaling*amountneeded)
		subset, timelimit = rd.getSubset(scaling*amountneeded, child, subset, all, timelimit)
	}
	return subset, timelimit
}
