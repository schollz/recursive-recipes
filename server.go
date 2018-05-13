package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/schollz/recursive-recipes/recipe"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(middleWareHandler(), gin.Recovery())
	router.GET("/ws/:recipe", wshandler)
	router.GET("/recipe/:recipe", func(c *gin.Context) {
		recipeName := c.Param("recipe")
		log.Println("got recipe", recipeName)
		if recipeName == "" {
			// TODO: handle if no recipe, do redirect
		}
		// check if recipe exists
		c.File("./scratch/app/build/index.html")

	})
	router.Static("/asset-manifest.json", "./scratch/app/build/asset-manifest.json")
	router.Static("/service-worker.js", "./scratch/app/build/service-worker.js")
	// router.Static("/a", "./scratch/app/build/")
	router.Static("/static", "./scratch/app/build/static")
	log.Println("running on ", ":8012")
	router.Run(":" + "8012")
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
	serverPayload, err := recipe.GetRecipe(recipeToGet, 1, make(map[string]struct{}))
	if err != nil {
		log.Println(err)
		return
	}
	serverPayloadBytes, _ := json.Marshal(serverPayload)
	log.Println(string(serverPayloadBytes))
	err = c.WriteMessage(1, serverPayloadBytes)
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
		serverPayload, err := recipe.GetRecipe(clientPayload.Recipe, clientPayload.MinutesToBuild/60, clientPayload.IngredientsToBuild)
		if err != nil {
			log.Println(err)
			continue
		}
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
