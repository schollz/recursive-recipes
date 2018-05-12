package recipe

import (
	"io/ioutil"
	"testing"

	"github.com/BurntSushi/toml"
	log "github.com/cihub/seelog"
	"github.com/stretchr/testify/assert"
)

func BenchmarkGetRecipe(b *testing.B) {
	SetLogLevel("info")
	for i := 0; i < b.N; i++ {
		GetRecipe("chocolate chip cookies", 100000)
	}
}

func TestOpen2(t *testing.T) {
	defer log.Flush()
	err := SetLogLevel("info")
	assert.Nil(t, err)

	err = GetRecipe("chocolate chip cookies", 1)
	assert.Nil(t, err)

}

func TestPruneByIngredient(t *testing.T) {
	defer log.Flush()
	err := SetLogLevel("info")
	// collect all the possible reactions
	var r Reactions
	b, _ := ioutil.ReadFile("../recipes.toml")
	_, err = toml.Decode(string(b), &r)
	if err != nil {
		return
	}
	reactions := make(map[string]Reaction)
	for _, reaction := range r.Reactions {
		for _, product := range reaction.Product {
			if _, ok := reactions[product.Name]; ok {
				log.Debugf("uh oh, already have %s", product.Name)
			} else {
				reactions[product.Name] = reaction
				reactions[product.Name].Product[0] = product
			}
		}
	}

	// get tree based on recipe and amount
	recipe := "chocolate chip cookies"
	log.Debug(reactions[recipe])
	d := new(Dag)
	recipeToGet := reactions[recipe].Product[0]
	log.Debug(reactions[recipe].Product[0])
	recursivelyAddRecipe(Element{
		Name:    recipeToGet.Name,
		Amount:  recipeToGet.Amount * 2,
		Measure: recipeToGet.Measure,
		Price:   recipeToGet.Price,
	}, d, reactions)

	// TODO: prune tree by time or price
	ingredientsToMake := make(map[string]struct{})
	ingredientsToMake[recipe] = struct{}{}
	ingredientsToMake["oatmeal"] = struct{}{}
	pruneTreeByIngredients(d, ingredientsToMake)
	log.Info(printDag(d))
}
