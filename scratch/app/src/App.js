import React, { Component } from 'react';
import ReactDOM from 'react-dom';
import Slider from 'rc-slider';
import Sockette from 'sockette';
import createHistory from "history/createBrowserHistory"
// import HoverableBox from './HoverableBox'
import 'rc-slider/assets/index.css';
import './index.css'

var moment = require("moment");
var momentDurationFormatSetup = require("moment-duration-format");
const history = createHistory()
const queryString = require('query-string');

class App extends Component {

    constructor(props) {
      super(props);
      this.timeout = null;
      // PRODUCTION
      let websocketURL ="ws"+window.origin.substring(4,window.origin.length)+window.location.pathname.replace("/recipe/","/ws/");

      // DEBUG
      // let websocketURL = "ws://127.0.0.1:8012/ws/chocolate-chip-cookies";
      this.ws = new Sockette(websocketURL, {
        timeout: 5e3,
        maxAttempts: 10,
        onopen: e =>  this.requestFromServer(),
        onmessage: e => this.handleData(e),
        onreconnect: e => console.log('Reconnecting...', e),
        onmaximum: e => console.log('Stop Attempting!', e),
        onclose: e => console.log('Closed!', e),
        onerror: e => console.log('Error:', e)
      });
      // this.ws.send('Hello, world!');
      
      // // Reconnect 10s later
      // setTimeout(this.ws.reconnect, 10e3);
      let recipe = window.location.pathname.replace("/recipe/","").replace(/-/g,' ').replace(/\//g,' ').trim();
      console.log("websocketURL:"+websocketURL);
      this.state = {
        loading: true,
        websocketURL: websocketURL,
        // version: "v0.0.0",
        // totalCost: "$2.30",
        // totalTime: "3 days, 2 hours",
        version: "",
        totalCost: "",
        totalTime: "",
        amount: 0.0,
        measure: "",
        recipe: recipe,
        limitfactor: 0,
        graph: "",
        ingredientsToBuild: {},
        ingredients: [
        //   {
        //     amount: "1 1/2 cup",
        //     name: "Flour",
        //     cost: "$1.00",
        //     scratchTime: "+2 hours",
        //     scratchCost: "-$1.00",
        //   },
        //   {
        //     amount: "1 cup",
        //     name: "Chocolate Chips",
        //     cost: "$1.34",
        //     scratchTime: "+1 hours",
        //     scratchCost: "-$0.30",
        //   }

        ],
        directions: [
        //   {
        //     name:'Milk',
        //     texts: ['Milk the cow.','Make milk'],
        //     totalTime: "1 day",
        // },
        ]
      };
      let queries = queryString.parseUrl(window.location.href);
      if ('amount' in queries.query) {
        this.state.amount = Number(queries.query.amount);
      }
      if ('timelimit' in queries.query) {
        this.state.limitfactor = Math.round(Math.log10(queries.query.timelimit)/Math.log10(1.8));
      }
      console.log(queries);
      if ('ingredientsToBuild' in queries.query) {
        queries.query.ingredientsToBuild.split(",").forEach(function(e) {
          this.state.ingredientsToBuild[e] = {};
        }.bind(this));
      }
    }



  handleData(data) {
    console.log(data);
    let result = JSON.parse(data.data);
    console.log(result.ingredients);
    console.log(result.minutes);
    let limitfactor =Math.log10(result.minutes)/Math.log10(1.8);
    console.log(limitfactor);
    this.setState({
      loading:false,
      limitfactor: limitfactor,
      graph: "/"+result.graph,
      recipe: result.recipe,
      version: result.version,
      ingredients: result.ingredients,
      directions: result.directions,
      totalCost: result.totalCost,
      totalTime: result.totalTime,
      amount: result.amount,
      measure: result.measure,
    })
    // this.setState({
    //   limitfactor:10,
    // });
  }

  handleClick = (data,e) => {
    e.preventDefault();
    console.log(data);
    this.state.ingredientsToBuild[(""+data).toLowerCase()] = {};
    console.log(this.state.ingredientsToBuild);
    this.setState({
      ingredientsToBuild: this.state.ingredientsToBuild,
    });
    this.requestFromServer();
}



handleClick2 = (data,e) => {
  e.preventDefault();
  console.log(data);
  delete(this.state.ingredientsToBuild[(""+data).toLowerCase()]);
  console.log(this.state.ingredientsToBuild);
  this.setState({
    ingredientsToBuild: this.state.ingredientsToBuild,
  });
  this.requestFromServer();
}

  requestFromServer() {
    let payload = JSON.stringify({
      recipe: this.state.recipe.toLowerCase(),
      ingredientsToBuild: this.state.ingredientsToBuild,
      minutes: Math.pow(1.8,this.state.limitfactor),
      amount: this.state.amount,
    });
    console.log("sending"+payload);
    this.ws.send(payload);
  }

  handleOnChange(value) {
    clearTimeout(this.timeout);
    this.timeout = setTimeout((function(){
      this.requestFromServer();
    }).bind(this),250);
  
    this.setState({
      limitfactor: value,
    })
  }

  handleOnChange2(value) {
    clearTimeout(this.timeout);
    this.timeout = setTimeout((function(){
      this.requestFromServer();
    }).bind(this),250);
  
    this.setState({
      amount: value,
    })
  }

  onBoxMouseover(ing,e) {
    console.log(ing);
    var i;
    for (i=0; i < this.state.ingredients.length; i++) {
      if (ing.name === this.state.ingredients[i].name) {
        break;
      }
    }
    console.log(i);
    this.state.ingredients[i].show = true;
    this.setState({
      ingredients: this.state.ingredients,
    })
  }

  onBoxMouseOut(ing,e) {
    console.log(ing);
    var i;
    for (i=0; i < this.state.ingredients.length; i++) {
      if (ing.name === this.state.ingredients[i].name) {
        break;
      }
    }
    console.log(i);
    this.state.ingredients[i].show = false;
    this.setState({
      ingredients: this.state.ingredients,
    })
  }


  render() {
    String.prototype.toTitleCase = function(){
      var smallWords = /^(a|an|and|as|at|but|by|en|for|if|in|nor|of|on|or|per|the|to|vs?\.?|via)$/i;
    
      return this.replace(/[A-Za-z0-9\u00C0-\u00FF]+[^\s-]*/g, function(match, index, title){
        if (index > 0 && index + match.length !== title.length &&
          match.search(smallWords) > -1 && title.charAt(index - 2) !== ":" &&
          (title.charAt(index + match.length) !== '-' || title.charAt(index - 1) === '-') &&
          title.charAt(index - 1).search(/[^\s-]/) < 0) {
          return match.toLowerCase();
        }
    
        if (match.substr(1).search(/[A-Z]|\../) > -1) {
          return match;
        }
    
        return match.charAt(0).toUpperCase() + match.substr(1);
      });
    };

  var listDirections;
    if (this.state.directions.length === 0) {
      listDirections =  <div className="outsidebox">
      <h2>Make the {this.state.recipe}</h2>
      <ol>
        <li>Go and buy it.</li>
        </ol>
  </div>

    } else {
      listDirections = this.state.directions.map((direction) =>
      <div className="boxwrapper">
        <div className="outsidebox">
            <h2>Make the {direction.name} ({direction.totalTime})</h2>
             <ol>
               {direction.texts.map((text) => <li>{text}</li> )}
            </ol>
        </div>
      </div>
      );
    }
    const listItems = this.state.ingredients.map((ing) =>
    <div onMouseEnter={this.onBoxMouseover.bind(this,ing)} onMouseLeave={this.onBoxMouseOut.bind(this,ing)} className={"box " + (ing.scratchCost !== '' ? 'clickable' : '')} onClick={this.handleClick.bind(this,ing.name)}>
    <h3>
    <span className="small-caps">{ing.amount}{ing.cost !== '' &&
    <span> / {ing.cost}</span> 
      }</span>
    <span className="display-block">
    {ing.scratchCost === '' ? (ing.name.toTitleCase()) :(
      <span>{ing.name.toTitleCase()}</span>
    )} 
    </span>
    </h3>
    
      {ing.scratchCost !== '' && ing.show &&
      <p>{ing.scratchCost} by making {ing.name.toLowerCase()} from scratch in {ing.scratchTime}.</p>
      }
	    </div>
	  );

    const numIngredientsToBuild = Object.keys(this.state.ingredientsToBuild).length;
    let ListIngredientsToBuildSpan = <span></span>
    if (numIngredientsToBuild > 0) {
      const ingredientList = Object.keys(this.state.ingredientsToBuild).map((ing,i) => 
        <span>{i !== 0 && <span>,</span> } <a href="#"  onClick={this.handleClick2.bind(this,ing)} className="pr0 nounderline"><em>{ing}</em></a></span>
      );
      ListIngredientsToBuildSpan = <div className="hero-text2">
      Ingredients to make from scratch <small>(click to remove)</small>: {ingredientList}
      </div>
    }


    history.push(window.location.pathname + "?amount="+this.state.amount + "&timelimit=" + Math.round(Math.pow(1.8,this.state.limitfactor)) + "&ingredientsToBuild="+Object.keys(this.state.ingredientsToBuild).join(","));
    return (
      <div className="App">

        <header className="padding-top-xs text-center color-white backgroundblue">
            <div className="container">
                <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" className="feather feather-book-open">
                    <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path>
                    <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path>
                </svg>
                <h1 className="display-title"><a href="/" className="nounderline">Recursive Recipes</a></h1>
                <p className="no-margin"><strong>{this.state.version}</strong></p>
                
            </div>
        </header>
        <main className="padding-vertical-xl color-white backgroundblue">
{this.state.version !== '' ? (
        <div className="container">
            <h2 className="hero-text">
    <span>{this.state.recipe}</span>
    <small>{this.state.totalCost.split(" ")[1]} | </small>
    <small>{this.state.totalTime}</small>
    </h2>

<div className="flex-grid">
<div className="col">

<span className="hero-text2">
<span className="firstStep">Amount:</span> {this.state.amount} {this.state.measure}
<div className="slider">
<Slider max="100" step="1" value={this.state.amount} onChange={this.handleOnChange2.bind(this)} />
</div>
</span>

</div>

<div className="col">
<span className="hero-text2">
Time limit:  {moment.duration(Math.pow(1.8,this.state.limitfactor), "minutes").format("Y [years], M [months], w [weeks], d [days], h [hrs], m [min]")}
<div className="slider">
<Slider max="30" step="0.01" value={this.state.limitfactor} onChange={this.handleOnChange.bind(this)} />
</div>
</span>
</div>
</div>

{/* <ListIngredientsToBuild ingredientsToBuild={this.state.ingredientsToBuild}/> */}
{ListIngredientsToBuildSpan}

<div className="flex-grid">
            <div className="col pr1 margin-top-m">

<h2 className="display-title margin-top-xl">Ingredients</h2>
            <p className="lead max-width-xs">These are the ingredients to purchase before you start, which will cost <strong className="second-step">{this.state.totalCost}</strong>. <em>Click on an ingredient to make it from scratch.</em></p>

           

            <div className="boxes">
                {listItems}
            </div>

</div>
<div className="col margin-top-m">

            <h2 className="display-title margin-top-xl">Directions</h2>
            <p className="lead max-width-xs">Follow these steps to make this recipe, which will take about <strong>{this.state.totalTime}</strong>.</p>
            {listDirections}

            </div>
            </div>

            <h2 className="display-title margin-top-xl">Recipe dependency graph</h2>
            <img src={this.state.graph} style={{paddingTop:'1em'}} />

          </div>

) : (
<div style={{height:'60vh',margin:'auto',textAlign:'center'}}>
<img src="/static/loader.svg" />
</div>

)}
        </main>
    <footer className="footer padding-vertical-m border-top backgroundblue color-white">
        <div className="container">
            <p>
                Designed and built by <a href="https://twitter.com/yakczar">yakczar</a> at <a href="https://schollz.github.io">schollz.github.io</a>.
            </p>
            <nav>
                {this.state.version} &middot;
                <a href="https://twitter.com/yakczar">Twitter</a> &middot;
                <a href="https://github.com/schollz">GitHub</a> &middot;
                <a href="https://github.com/schollz/recursive-recipes/issues">Comments</a>
            </nav>
        </div>
    </footer>
      </div>
    );
  }
}

export default App;
