package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/schollz/recursive-recipes/recipe"
)

const version = "v0.1.0"

var finishedRecipes = []string{
	"refried beans",
	"chocolate chip cookies",
	"pancakes",
	"yogurt",
	"eggs benedict",
	"english muffin",
	"tortilla",
	"noodles",
	"cheese",
	"vanilla ice cream",
	"apple pie",
}

var finishedRecipesMap map[string]struct{}

func init() {
	finishedRecipesMap = make(map[string]struct{})
	for _, r := range finishedRecipes {
		finishedRecipesMap[r] = struct{}{}
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(middleWareHandler(), gin.Recovery())
	router.SetFuncMap(template.FuncMap{
		"slugify": slugify,
		"totitle": totitle,
	})
	router.LoadHTMLGlob("templates/*")
	router.GET("/ws/:recipe", wshandler)
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "main.html", gin.H{
			"Version": version,
			"Recipes": finishedRecipes,
		})
	})
	router.GET("/recipe/:recipe", func(c *gin.Context) {
		recipeName := unslugify(c.Param("recipe"))
		log.Println("got recipe", recipeName)
		_, ok := finishedRecipesMap[recipeName]
		if recipeName == "" || !ok {
			c.HTML(http.StatusOK, "main.html", gin.H{
				"Version": version,
				"Recipes": finishedRecipes,
			})
			return
		}
		c.File("./scratch/app/build/index.html")
	})
	router.Static("/asset-manifest.json", "./scratch/app/build/asset-manifest.json")
	router.Static("/service-worker.js", "./scratch/app/build/service-worker.js")
	// router.Static("/a", "./scratch/app/build/")
	router.Static("/static", "./scratch/app/build/static")
	router.Static("/graphviz", "./graphviz")
	log.Println("running on ", ":8031")
	router.Run(":" + "8031")
}

func slugify(s string) string {
	return strings.ToLower(strings.Join(strings.Split(strings.TrimSpace(s), " "), "-"))
}

func totitle(s string) string {
	return strings.Title(s)
}

func unslugify(s string) string {
	return strings.TrimSpace(strings.ToLower(strings.Join(strings.Split(s, "-"), " ")))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func wshandler(cg *gin.Context) {
	recipeToGet := strings.Replace(cg.Param("recipe"), "-", " ", -1)
	if recipeToGet == "" {
		cg.String(404, "")
		return
	}
	log.Println(recipeToGet)

	var w http.ResponseWriter = cg.Writer
	var r *http.Request = cg.Request

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	// TODO: send a recipe
	// a := "chocolate"
	// bPayload, _ := json.Marshal(a)
	// err = c.WriteMessage(1, bPayload)
	// serverPayload, err := recipe.GetRecipe(recipeToGet, 0, 1, make(map[string]struct{}))
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// serverPayload.Version = version
	// serverPayloadBytes, _ := json.Marshal(serverPayload)
	// log.Println(string(serverPayloadBytes))
	// err = c.WriteMessage(1, serverPayloadBytes)
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		var clientPayload recipe.RequestFromApp
		err = json.Unmarshal(message, &clientPayload)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("clientPayload", clientPayload)
		if len(clientPayload.IngredientsToBuild) > 0 {
			clientPayload.IngredientsToBuild[clientPayload.Recipe] = struct{}{}
		}
		serverPayload, err := recipe.GetRecipe(clientPayload.Recipe, clientPayload.Amount, clientPayload.MinutesToBuild/60, clientPayload.IngredientsToBuild)
		if err != nil {
			log.Println(err)
			continue
		}
		serverPayload.Version = version
		serverPayloadBytes, _ := json.Marshal(serverPayload)
		log.Println(string(serverPayloadBytes))
		err = c.WriteMessage(mt, serverPayloadBytes)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func addCORS(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
}

func middleWareHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// t := time.Now()
		// Add base headers
		addCORS(c)
		// Run next function
		c.Next()
		// // Log request
		// log.Infof("%v %v %v %s", c.Request.RemoteAddr, c.Request.Method, c.Request.URL, time.Since(t))
	}
}
