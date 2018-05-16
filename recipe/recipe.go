package recipe

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/cihub/seelog"
)

type Reactions struct {
	Reactions []Reaction `toml:"reaction"`
}

// Reaction is the reaction that takes place, with reactants
// and products in either their theoretical OR practical
// lowest irreducible. This is the format of the toml file for all
// the recipes.
type Reaction struct {
	// ParallelHours it takes to create the products from reactants, in parallel.
	// For example, trees grow at the same rate no matter how many there are,
	// so there time cost would be almost entirely "parallel". Parallel time
	// does not scale when scaling the recipe
	ParallelHours float64 `toml:"p_hours" json:"p_hours,omitempty"`

	// SerialHours is the serial amount of time needed, time that scales
	// proportional to the quantities.
	SerialHours float64 `toml:"s_hours" json:"s_hours,omitempty"`

	Directions string    `toml:"directions" json:"directions,omitempty"`
	Notes      string    `toml:"notes" json:"notes,omitempty"`
	Product    []Element `toml:"product" json:"product,omitempty"`
	Reactant   []Element `toml:"reactant" json:"reactant,omitempty"`

	// LastUpdated is the year it was last updated
	// (refers to the price)
	LastUpdated time.Time `toml:"updated" json:"updated,omitempty"`
}

type Element struct {
	// Name is the name of the product/reactant
	Name string `toml:"name" json:"name,omitempty"`

	// Amount is the amount
	Amount float64 `toml:"amount" json:"amount,omitempty"`

	// Measure is the type of measurement, either
	// "grams" (weight), "cups" (volume), "acres" (sq ft),
	// or "whole"
	Measure string `toml:"measure" json:"measure,omitempty"`

	// Price is the cost per amount+measure, specified on products.
	Price float64 `toml:"price" json:"price,omitempty"`

	// Notes are for references
	Notes string `toml:"notes" json:"notes,omitempty"`
}

// Dag is the format that the reactions are parsed into. Each root only
// corresponds to a single Product (which is different than a Reaction). The
// children correspond to dags of the reactions of the reactants. Everything
// else is pretty much carried over from the reaction.
type Dag struct {
	ParallelHours float64   `toml:"p_hours" json:"p_hours,omitempty"`
	SerialHours   float64   `toml:"s_hours" json:"s_hours,omitempty"`
	Directions    string    `toml:"directions" json:"directions,omitempty"`
	Notes         string    `toml:"notes" json:"notes,omitempty"`
	Product       Element   `toml:"product" json:"product,omitempty"`
	Reactant      []Element `toml:"reactant" json:"reactant,omitempty"`
	Children      []*Dag
}

type UpdateApp struct {
	Graph       string                 `json:"graph"`
	Version     string                 `json:"version"`
	Recipe      string                 `json:"recipe"`
	Amount      float64                `json:"amount"`
	Measure     string                 `json:"measure"`
	TotalCost   string                 `json:"totalCost"`
	TotalTime   string                 `json:"totalTime"`
	Ingredients []UpdateAppIngredients `json:"ingredients"`
	Directions  []UpdateAppDirections  `json:"directions"`
}

type UpdateAppIngredients struct {
	Amount      string `json:"amount"`
	Name        string `json:"name"`
	Cost        string `json:"cost"`
	ScratchTime string `json:"scratchTime"`
	ScratchCost string `json:"scratchCost"`
}

type UpdateAppDirections struct {
	Name      string   `json:"name"`
	TotalTime string   `json:"totalTime"`
	Texts     []string `json:"texts"`
}

type RequestFromApp struct {
	Amount             float64             `json:"amount"`
	Measure            string              `json:"measure"`
	Recipe             string              `json:"recipe"`
	IngredientsToBuild map[string]struct{} `json:"ingredientsToBuild"`
	MinutesToBuild     float64             `json:"minutes"`
}

func GetRecipe(recipe string, amountSpecified float64, hours float64, ingredientsToInclude map[string]struct{}) (payload UpdateApp, err error) {
	payload.Version = "v0.0.0"
	payload.Recipe = recipe

	// collect all the possible reactions
	var r Reactions
	b, err := ioutil.ReadFile("recipes.toml")
	if err != nil {
		return
	}
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
				reactions[product.Name] = Reaction{
					Directions:    reaction.Directions,
					LastUpdated:   reaction.LastUpdated,
					Notes:         reaction.Notes,
					ParallelHours: reaction.ParallelHours,
					SerialHours:   reaction.SerialHours,
					Reactant:      reaction.Reactant,
					Product: []Element{{
						Name:    product.Name,
						Amount:  product.Amount,
						Measure: product.Measure,
						Notes:   product.Notes,
						Price:   product.Price,
					}},
				}

			}
		}
	}

	// get tree based on recipe and amount
	log.Debug(reactions[recipe])
	d := new(Dag)
	recipeToGet := reactions[recipe].Product[0]
	log.Debug(reactions[recipe].Product[0])
	recipeToBuildFrom := Element{
		Name:    recipeToGet.Name,
		Amount:  amountSpecified,
		Measure: recipeToGet.Measure,
		Price:   recipeToGet.Price,
	}
	if recipeToBuildFrom.Amount == 0 {
		recipeToBuildFrom.Amount = recipeToGet.Amount
	}
	recursivelyAddRecipe(recipeToBuildFrom, d, reactions)
	payload.Amount = recipeToBuildFrom.Amount
	payload.Measure = recipeToBuildFrom.Measure

	// get graphviz for full graph
	payload.Graph, err = getGraphviz(d)
	if err != nil {
		log.Error(err)
		return
	}

	totalTime := pruneTreeByTimeAndIngredients(d, 0, hours, ingredientsToInclude)
	log.Info("totalTime", totalTime, FormatDuration(totalTime))
	payload.TotalTime = FormatDuration(totalTime)
	if payload.TotalTime == "" {
		payload.TotalTime = "No time"
	}

	// parse tree for ingredients to build and the ingredients to buy
	ingredientsToBuild, ingredientsToBuy := getIngredientsToBuild(d, []Element{}, []Element{})
	log.Debug("\nIngredients to build:")
	for _, ing := range ingredientsToBuild {
		log.Debug("-", ing.Name, ing.Amount)
	}
	log.Debug("\nIngredients to buy:")
	payload.Ingredients = make([]UpdateAppIngredients, len(ingredientsToBuy))
	totalCost := 0.0
	for i, ing := range ingredientsToBuy {
		log.Debug("ingredientsToBuy", ing.Name, ing.Amount, ing.Price)
		totalCost += ing.Price
		payload.Ingredients[i].Name = ing.Name
		payload.Ingredients[i].Amount = FormatMeasure(ing.Amount, ing.Measure)
		payload.Ingredients[i].Cost = fmt.Sprintf("$%2.2f", ing.Price)
		priceDifference, timeDifference, errScratch := scratchReplacement(reactions, ing.Name, ing.Amount)
		if errScratch != nil {
			log.Warn(errScratch)
			continue
		}
		log.Info(ing.Name, priceDifference, timeDifference)
		payload.Ingredients[i].ScratchCost = FormatCost(priceDifference)
		payload.Ingredients[i].ScratchTime = FormatDuration(timeDifference)
	}
	log.Debug("totalCost", totalCost)
	payload.TotalCost = FormatCost(totalCost)
	if len(payload.TotalCost) > 1 {
		payload.TotalCost = payload.TotalCost[1:]
	} else {
		payload.TotalCost = "$0"
	}

	// collect the roots
	log.Debug("collect the roots")
	roots := getDagRoots(d, []*Dag{})
	rootMap := make(map[string]*Dag)
	for _, root := range roots {
		if _, ok := rootMap[root.Product.Name]; ok {
			log.Debug(root.Product.Name)
		}
		rootMap[root.Product.Name] = root
	}

	// log.Debug(pathExists(rootMap["cow milk"], rootMap["cow milk"])) // true
	// log.Debug(pathExists(rootMap["milk"], rootMap["cow milk"]))     // true
	// log.Debug(pathExists(rootMap["cow milk"], rootMap["milk"]))     // false

	// DETERMINE THE BEST ORDERING
	// find ingredients to build that don't depend on any ingredients to build
	log.Debug("determine the best ordering")
	ingredientsToBuildMap := make(map[string]struct{})
	for _, ing := range ingredientsToBuild {
		ingredientsToBuildMap[ing.Name] = struct{}{}
	}
	directionsOrder := []string{}
	for {
		if len(ingredientsToBuildMap) == 0 {
			break
		}
		thingsThatCanBeBuiltNow := make(map[string]struct{})
		for ing1 := range ingredientsToBuildMap {
			var ing1DependsOnIng2 bool
			for ing2 := range ingredientsToBuildMap {
				if ing1 == ing2 {
					continue
				}
				// make sure ing1 doesn't depend on ing2
				if pathExists(rootMap[ing1], rootMap[ing2]) {
					ing1DependsOnIng2 = true
					break
				}
			}
			if !ing1DependsOnIng2 {
				thingsThatCanBeBuiltNow[ing1] = struct{}{}
			}
		}
		if len(ingredientsToBuildMap) == 1 {
			for ing1 := range ingredientsToBuildMap {
				thingsThatCanBeBuiltNow[ing1] = struct{}{}
			}
		}

		// find the one that takes the longest
		longestTime := -1.0
		currentThing := ""
		for ing := range thingsThatCanBeBuiltNow {
			timeTaken := rootMap[ing].SerialHours + rootMap[ing].ParallelHours
			if timeTaken > longestTime {
				longestTime = timeTaken
				currentThing = ing
			}
		}

		// log.Debug("thingsThatCanBeBuiltNow", thingsThatCanBeBuiltNow)
		// log.Debug("ingredientsToBuildMap", ingredientsToBuildMap)
		// log.Debug("currentThing", currentThing)
		directionsOrder = append(directionsOrder, currentThing)
		// delete it from things to build, and iterate
		delete(ingredientsToBuildMap, currentThing)
	}
	log.Debug(directionsOrder)
	log.Debug(printDag(d))
	payload.Directions = make([]UpdateAppDirections, len(directionsOrder))
	for i, direction := range directionsOrder {
		payload.Directions[i].Name = direction
		payload.Directions[i].TotalTime = FormatDuration(rootMap[direction].SerialHours + rootMap[direction].ParallelHours)
		payload.Directions[i].Texts = []string{}
		for _, text := range strings.Split(rootMap[direction].Directions, "\n") {
			text = strings.TrimSpace(text)
			if len(text) == 0 {
				continue
			}
			payload.Directions[i].Texts = append(payload.Directions[i].Texts, text)
		}
	}

	log.Info(scratchReplacement(reactions, "milk", 1))

	return
}

func scratchReplacement(reactions map[string]Reaction, ing string, amount float64) (priceDifference float64, timeDifference float64, err error) {
	if _, ok := reactions[ing]; !ok {
		err = errors.New("no such reaction for " + ing)
		return
	}
	if len(reactions[ing].Reactant) == 0 {
		err = errors.New("no such reaction for " + ing)
		return
	}
	scaling := amount / reactions[ing].Product[0].Amount
	priceToBuy := reactions[ing].Product[0].Price * scaling
	timeToBuild := reactions[ing].SerialHours*scaling + reactions[ing].ParallelHours

	priceToBuild := 0.0
	for _, child := range reactions[ing].Reactant {
		if _, ok := reactions[child.Name]; !ok {
			continue
		}
		childScaling := child.Amount / reactions[child.Name].Product[0].Amount
		priceToBuild += reactions[child.Name].Product[0].Price * scaling * childScaling
		log.Info("scratchReplacement", child.Name, reactions[child.Name].Product[0].Price*scaling)
	}
	log.Info(ing, amount, priceToBuy, timeToBuild, priceToBuild)
	priceDifference = priceToBuild - priceToBuy
	timeDifference = timeToBuild - 0
	return
}

// func scratchReplacement(d *Dag) (priceDifference float64, timeDifference float64) {
// 	priceToBuy := d.Product.Price
// 	priceToBuild := 0.0
// 	timeToBuild := d.SerialHours + d.ParallelHours
// 	log.Info(d.Product.Name, d.Product.Price)
// 	for _, child := range d.Children {
// 		log.Info(child.Product.Name)
// 		if len(child.Children) > 0 {
// 			log.Info(child.Product.Name, child.Product.Price)
// 		}
// 		priceToBuild += child.Product.Price
// 	}
// 	priceDifference = priceToBuild - priceToBuy
// 	timeDifference = timeToBuild - 0.0
// 	return
// }

func pruneTreeByTimeAndIngredients(d *Dag, currentTime float64, maxTime float64, ingredientsToMake map[string]struct{}) float64 {
	_, ingredientToMake := ingredientsToMake[d.Product.Name]
	if currentTime+d.SerialHours+d.ParallelHours > maxTime && !ingredientToMake {
		d.Children = []*Dag{}
	} else {
		currentTime += d.SerialHours + d.ParallelHours
		for _, child := range d.Children {
			currentTime = pruneTreeByTimeAndIngredients(child, currentTime, maxTime, ingredientsToMake)
		}
	}
	return currentTime
}

func pruneTreeByIngredients(d *Dag, ingredientsToMake map[string]struct{}) {
	if _, ok := ingredientsToMake[d.Product.Name]; !ok {
		d.Children = []*Dag{}
	} else {
		for _, child := range d.Children {
			pruneTreeByIngredients(child, ingredientsToMake)
		}
	}
}

func pruneTreeByTime(d *Dag, currentTime float64, maxTime float64) {
	currentTime += d.SerialHours + d.ParallelHours
	if currentTime > maxTime {
		d.Children = []*Dag{}
	} else {
		for _, child := range d.Children {
			pruneTreeByTime(child, currentTime, maxTime)
		}
	}
}

func printDag(d *Dag) string {
	return printDagRecursively(d, 0)
}

func printDagRecursively(d *Dag, in int) string {
	s := ""
	for i := 0; i < in; i++ {
		s += "---"
	}
	s += fmt.Sprintf("%s %2.3f %s\n", d.Product.Name, d.Product.Amount, d.Product.Measure)
	for _, child := range d.Children {
		s += printDagRecursively(child, in+1)
	}
	return s
}

func generateDagGraphviz(d *Dag, dots map[string]struct{}) map[string]struct{} {
	dots[fmt.Sprintf(`"%s" [color="white", fontcolor="white"];`, d.Product.Name)] = struct{}{}
	for _, child := range d.Children {
		dots[fmt.Sprintf(`"%s" -> "%s" [style="filled", color="white"];`, child.Product.Name, d.Product.Name)] = struct{}{}
		dots = generateDagGraphviz(child, dots)
	}
	return dots
}

func pathExists(fromNode *Dag, toNode *Dag) bool {
	if fromNode.Product.Name == toNode.Product.Name {
		return true
	}
	for _, child := range fromNode.Children {
		if pathExists(child, toNode) {
			return true
		}
	}
	return false
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
	// add basic element
	d.Product = Element{
		Name:    recipe.Name,
		Notes:   recipe.Notes,
		Measure: recipe.Measure,
		Amount:  recipe.Amount,
		Price:   recipe.Price,
	}
	d.Product = recipe
	// if recipe.Measure == "whole" {
	// 	d.Product.Amount = math.Ceil(d.Product.Amount)
	// }

	// add children, if any
	d.Children = []*Dag{}
	if _, ok := reactions[recipe.Name]; ok {
		// determine the scaling from the baseline reaction
		scaling := recipe.Amount / reactions[recipe.Name].Product[0].Amount
		log.Debug("A:", recipe.Name, scaling, recipe.Amount, reactions[recipe.Name].Product[0].Amount)

		d.Directions = reactions[recipe.Name].Directions
		d.Notes = reactions[recipe.Name].Notes
		d.ParallelHours = reactions[recipe.Name].ParallelHours
		d.SerialHours = scaling * reactions[recipe.Name].SerialHours          // scale the time
		d.Product.Amount = scaling * reactions[recipe.Name].Product[0].Amount // scale the amount
		d.Product.Price = scaling * reactions[recipe.Name].Product[0].Price   // scale the price
		d.Reactant = make([]Element, len(reactions[recipe.Name].Reactant))
		for i, r := range reactions[recipe.Name].Reactant {
			d.Reactant[i].Amount = r.Amount * scaling // scale the amount
			// if d.Reactant[i].Measure == "whole" {
			// 	d.Reactant[i].Amount = math.Ceil(d.Reactant[i].Amount)
			// }
			d.Reactant[i].Price = r.Price * scaling // scale the price (though there shouldn't be one)
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

// SetLogLevel determines the log level
func SetLogLevel(level string) (err error) {

	// https://en.wikipedia.org/wiki/ANSI_escape_code#3/4_bit
	// https://github.com/cihub/seelog/wiki/Log-levels
	appConfig := `
	<seelog minlevel="` + level + `">
	<outputs formatid="stdout">
	<filter levels="debug,trace">
		<console formatid="debug"/>
	</filter>
	<filter levels="info">
		<console formatid="info"/>
	</filter>
	<filter levels="critical,error">
		<console formatid="error"/>
	</filter>
	<filter levels="warn">
		<console formatid="warn"/>
	</filter>
	</outputs>
	<formats>
		<format id="stdout"   format="%Date %Time [%LEVEL] %File %FuncShort:%Line %Msg %n" />
		<format id="debug"   format="%Date %Time %EscM(37)[%LEVEL]%EscM(0) %File %FuncShort:%Line %Msg %n" />
		<format id="info"    format="%Date %Time %EscM(36)[%LEVEL]%EscM(0) %File %FuncShort:%Line %Msg %n" />
		<format id="warn"    format="%Date %Time %EscM(33)[%LEVEL]%EscM(0) %File %FuncShort:%Line %Msg %n" />
		<format id="error"   format="%Date %Time %EscM(31)[%LEVEL]%EscM(0) %File %FuncShort:%Line %Msg %n" />
	</formats>
	</seelog>
	`
	logger, err := log.LoggerFromConfigAsBytes([]byte(appConfig))
	if err != nil {
		return
	}
	log.ReplaceLogger(logger)
	return
}

func getGraphviz(d *Dag) (graphvizFileName string, err error) {
	dotsMap := generateDagGraphviz(d, make(map[string]struct{}))
	dots := make([]string, len(dotsMap))
	i := 0
	for key := range dotsMap {
		dots[i] = key
		i++
	}
	sort.Strings(dots)
	log.Info(dots)
	graphvizData := fmt.Sprintf(`digraph G {
color="#FFFFFF"
bgcolor="#357EDD00" # RGBA (with alpha)

%s
}`, strings.Join(dots, "\n"))
	os.MkdirAll("graphviz", 0644)
	graphvizFileName = path.Join("graphviz", GetMD5Hash(graphvizData)+".png")
	if _, err = os.Stat(graphvizFileName); err == nil {
		log.Infof("already generated %s", graphvizFileName)
		return
	}
	log.Info(graphvizData, graphvizFileName)

	content := []byte(graphvizData)
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		return
	}

	defer os.Remove(tmpfile.Name()) // clean up

	if _, err = tmpfile.Write(content); err != nil {
		return
	}
	if err = tmpfile.Close(); err != nil {
		return
	}
	cmd := exec.Command("dot", "-Tpng", tmpfile.Name(), "-o"+graphvizFileName)
	_, err = cmd.CombinedOutput()
	return
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
