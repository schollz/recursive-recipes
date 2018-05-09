package recipe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	for _, reaction := range r.Reactions {
		for _, product := range reaction.Product {
			reaction1 := reaction
			reaction1.Product = []Element{product}
			rd.Node[reaction1.Product[0].Name] = reaction1
			rd.Children[reaction1.Product[0].Name] = make([]string, len(reaction1.Reactant))
			for i, reactant := range reaction.Reactant {
				rd.Children[reaction1.Product[0].Name][i] = reactant.Name
			}
		}
	}
	bJson, _ = json.MarshalIndent(rd, "", " ")
	fmt.Println(string(bJson))
}
