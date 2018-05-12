package recipe

import (
	"fmt"
	"io/ioutil"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/cihub/seelog"
	"github.com/stretchr/testify/assert"
)

func TestFormatString(t *testing.T) {
	defer log.Flush()
	assert.Equal(t, "5 days, 3 hours", FormatDuration(123))
	assert.Equal(t, "5 days", FormatDuration(120))
	assert.Equal(t, "-$1.10", FormatCost(-1.1))
	assert.Equal(t, "21 ⅛", FormatCookingRational(21.12))
	assert.Equal(t, "1 ⅜ cups", FormatMeasure(1.4, "cup"))
	assert.Equal(t, "1 ⅝ tablespoons", FormatMeasure(0.1, "cup"))
}

func FormatCookingRational(num float64) (s string) {
	// round to nearest eight
	wholeNum := math.Floor(num)
	// log.Debug((num - wholeNum) / 8)
	fractionNum := (math.Round((num-wholeNum)*8) / 8) / .125
	// log.Debug(wholeNum, fractionNum)
	if wholeNum > 0 {
		s = fmt.Sprintf("%2.0f", wholeNum)
	}
	switch fractionNum {
	case 1:
		s += " ⅛"
	case 2:
		s += " ¼"
	case 3:
		s += " ⅜"
	case 4:
		s += " ½"
	case 5:
		s += " ⅝"
	case 6:
		s += " ¾"
	case 7:
		s += " ⅞"
	}
	return
}

func convertCups(cups float64) (amount float64, measure string) {
	amount = cups
	if cups > 0.125 {
		measure = "cup"
	} else if cups > 0.0625 {
		measure = "tablespoon"
		amount *= 16
	} else {
		measure = "teaspoon"
		amount *= 48
	}
	return
}

func FormatMeasure(amount float64, measure string) (s string) {
	if measure == "cup" {
		amount, measure = convertCups(amount)
	}
	s = fmt.Sprintf("%s %s", FormatCookingRational(amount), measure)
	if amount > 0 {
		s += "s"
	}
	s = strings.TrimSpace(s)
	return
}

func FormatCost(cost float64) (s string) {
	if cost < 0 {
		s = fmt.Sprintf("-$%2.2f", math.Abs(cost))
	} else if cost > 0 {
		s = fmt.Sprintf("+$%2.2f", math.Abs(cost))
	}
	return
}

const Year = (365 * 24 * time.Hour)
const Week = (7 * 24 * time.Hour)
const Day = (24 * time.Hour)

func FormatDuration(hours float64) (s string) {
	if hours == 0 {
		return ""
	}
	s = formatDurationRecursively(time.Duration(hours) * time.Hour)
	s = strings.TrimSpace(s)
	s = s[:len(s)-1]
	return
}

func formatDurationRecursively(t time.Duration) (s string) {
	// log.Debug(t)
	if t.Seconds() == 0 {
		return
	}
	timesStrings := []string{"year", "week", "day", "hour", "minute", "second"}
	times := []time.Duration{Year, Week, Day, 1 * time.Hour, 1 * time.Minute, 0 * time.Minute}
	i := 0
	for {
		if t.Seconds() >= times[i].Seconds() {
			break
		}
		i++
	}
	if i == len(times)-1 {
		return
	}
	t2 := int64(t.Seconds() / times[i].Seconds())
	// log.Debug(t.Seconds(), times[i].Seconds())
	t = time.Duration(math.Mod(t.Seconds(), times[i].Seconds())) * time.Second
	if t2 > 1 {
		timesStrings[i] += "s"
	}
	return fmt.Sprintf("%d %s, ", t2, timesStrings[i]) + formatDurationRecursively(t)
}

func TestGetRecipe1(t *testing.T) {
	payload, err := GetRecipe("chocolate chip cookies", 1, make(map[string]struct{}))
	assert.Nil(t, err)
	fmt.Printf("%+v\n", payload)
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
}
