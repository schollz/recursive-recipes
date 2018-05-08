package recipe

import (
	"io/ioutil"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	b, _ := ioutil.ReadFile("../recipes.toml")
	var r Recipes
	_, err := toml.Decode(string(b), &r)
	assert.Nil(t, err)
	// bJson, _ := json.MarshalIndent(r, "", " ")
	// fmt.Println(string(bJson))
}
