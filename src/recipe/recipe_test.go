package recipe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/BurntSushi/toml"
	log "github.com/cihub/seelog"
	"github.com/stretchr/testify/assert"
)

func TestFormatString(t *testing.T) {
	defer log.Flush()
	assert.Equal(t, "5 days, 3 hours", FormatDuration(123))
	assert.Equal(t, "5 days", FormatDuration(120))
	assert.Equal(t, "30 minutes", FormatDuration(0.5))
	assert.Equal(t, "-$1.10", FormatCost(-1.1))
	assert.Equal(t, "21 ⅛", FormatCookingRational(21.12))
	assert.Equal(t, "1 ⅜ cups", FormatMeasure(1.4, "cup"))
	assert.Equal(t, "1 ⅝ tablespoons", FormatMeasure(0.1, "cup"))
}

func TestGetRecipe1(t *testing.T) {
	m := make(map[string]struct{})
	m["sugar"] = struct{}{}
	m["chocolate chip cookies"] = struct{}{}
	m["salt"] = struct{}{}
	m["vanilla"] = struct{}{}
	m["alcohol"] = struct{}{}
	payload, err := GetRecipe("chocolate chip cookies", 0, 1, m)
	assert.Nil(t, err)
	p, _ := json.MarshalIndent(payload, "", " ")
	fmt.Println(string(p))
}

func TestGetRecipe2(t *testing.T) {
	defer log.Flush()
	m := make(map[string]struct{})
	m["apple pie"] = struct{}{}

	payload, err := GetRecipe("apple pie", 0, 1, m)
	assert.Nil(t, err)
	p, _ := json.MarshalIndent(payload, "", " ")
	fmt.Println(string(p))
}

func TestGetRecipe3(t *testing.T) {
	defer log.Flush()
	payload, err := GetRecipe("noodles", 0, 1, make(map[string]struct{}))
	assert.Nil(t, err)
	p, _ := json.MarshalIndent(payload, "", " ")
	fmt.Println(string(p))
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
	// ingredientsToMake[recipe] = struct{}{}
	// ingredientsToMake["oatmeal"] = struct{}{}
	log.Info(pruneTreeByTimeAndIngredients(d, 0, 100, ingredientsToMake))
	log.Info(printDag(d))

	getGraphviz(d)

}
