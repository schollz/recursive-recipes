package recipe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	b, _ := ioutil.ReadFile("../recipes.toml")
	fmt.Println(string(b))
	var r Recipes
	r.Recipes = []Recipe{{
		Time:       "alskdjf",
		Directions: " alskdjf",
		Product: []Element{
			{Name: "food", Cups: 0.5},
		},
	}}
	fmt.Println(r)
	var firstBuffer bytes.Buffer
	e := toml.NewEncoder(&firstBuffer)
	err := e.Encode(r)
	assert.Nil(t, err)
	fmt.Println(firstBuffer.String())

	_, err = toml.Decode(string(b), &r)
	assert.Nil(t, err)
	bJson, _ := json.MarshalIndent(r, "", " ")
	fmt.Println(string(bJson))
}
