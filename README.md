# recursive-reicipes

This is an attempt to visualize the recursive nature of recipes. You can view the time it takes and the cost needed to make different recipes, even when you subtitute the ingredients by recipes themselves.

Try it out at https://recursiverecipes.schollz.com, or run it yourself.

## Run yourself 

**Requirements:**

- [ ] Node + Yarn
- [ ] Go
- [ ] Graphviz

**Build:**

```
$ go get -u github.com/schollz/recursive-recipe
$ cd $GOPATH/src/github.com/schollz/recursive-recipes/scratch/app
$ yarn install
$ yarn build
```

**Run:**

```
$ cd $GOPATH/src/github.com/schollz/recursive-recipes
$ go build -v
$ ./recursive-recipes
```

Now open up `localhost:8031`.

## Recipes

The recipes themselves are in the [recipes.toml](https://github.com/schollz/recursive-recipes/blob/master/recipes.toml) file. You can add/delete/edit recipes here, and then the app will automatically update.

# License

MIT


