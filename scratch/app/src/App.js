import React, { Component } from 'react';
import Slider from 'rc-slider';
import Sockette from 'sockette';
import 'rc-slider/assets/index.css';
import './index.css'

var moment = require("moment");
var momentDurationFormatSetup = require("moment-duration-format");


class App extends Component {

    constructor(props) {
      super(props);
      this.timeout = null;
      let websocketURL ="ws"+window.origin.substring(4,window.origin.length)+window.location.pathname.replace("/recipe/","/ws/");
      this.ws = new Sockette(websocketURL, {
        timeout: 5e3,
        maxAttempts: 10,
        onopen: e => console.log('Connected!', e),
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
        websocketURL: websocketURL,
        version: "v0.0.0",
        recipe: recipe,
        totalCost: "$2.30",
        totalTime: "3 days, 2 hours",
        limitfactor: 6,
        ingredientsToBuild: {},
        ingredients: [
          {
            amount: "1 1/2 cup",
            name: "Flour",
            cost: "$1.00",
            scratchTime: "+2 hours",
            scratchCost: "-$1.00",
          },
          {
            amount: "1 cup",
            name: "Chocolate Chips",
            cost: "$1.34",
            scratchTime: "+1 hours",
            scratchCost: "-$0.30",
          }

        ],
        directions: [
          {
            name:'Milk',
            texts: ['Milk the cow.','Make milk'],
            totalTime: "1 day",
        },
        ]
      };
    }


  handleData(data) {
    console.log(data);
    let result = JSON.parse(data.data);
    console.log(result.ingredients);
    this.setState({
      version: result.version,
      ingredients: result.ingredients,
      directions: result.directions,
      totalCost: result.totalCost,
      totalTime: result.totalTime,
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

  requestFromServer() {
    let payload = JSON.stringify({
      recipe: this.state.recipe.toLowerCase(),
      ingredientsToBuild: this.state.ingredientsToBuild,
      minutes: Math.pow(1.8,this.state.limitfactor),
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
    const linkStyle = {
      textDecoration:'none',
    }
    const listDirections = this.state.directions.map((direction) =>
    <div className="boxwrapper">
      <div className="outsidebox">
          <h2>Make the {direction.name} ({direction.totalTime})</h2>
           <ol>
             {direction.texts.map((text) => <li>{text}</li> )}
          </ol>
      </div>
    </div>
  );
    const listItems = this.state.ingredients.map((ing) =>
    <div className="box">
    <h3>
    <span className="small-caps">{ing.amount}{ing.cost !== '' &&
    <span> / {ing.cost}</span> 
      }</span>
    <span className="display-block"><a href="#" onClick={this.handleClick.bind(this,ing.name)} style={linkStyle}>{ing.name.toTitleCase()}</a>
    </span>
    </h3>
    
      {ing.scratchCost !== '' &&
      <p>{ing.scratchCost}, {ing.scratchTime} to make {ing.name.toLowerCase()} from scratch.</p>
      }
    </div>
  );
return (
      <div className="App">
      
        <header className="padding-top-xs text-center color-white backgroundblue">
            <div className="container">
                <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" className="feather feather-book-open">
                    <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path>
                    <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path>
                </svg>
                <h1 className="display-title"><a href="/" className="nounderline">Recursive Cookbook</a></h1>
                <p className="no-margin"><strong>{this.state.version}</strong></p>
                
            </div>
        </header>
        <main className="padding-vertical-xl color-white backgroundblue">

        <div className="container">
            <h2 className="hero-text">
    <span>{this.state.recipe}</span>
    <small>{this.state.totalCost} | </small>
    <small>{this.state.totalTime}</small>
    </h2>

<span className="hero-text2">
Time limit:  {moment.duration(Math.pow(1.8,this.state.limitfactor), "minutes").format("Y [years], M [months], w [weeks], d [days], h [hrs], m [min]")}
<div className="slider">
<Slider max="30" step="0.01" value={this.state.limitfactor} onChange={this.handleOnChange.bind(this)} />
</div>
</span>

<div className="flex-grid">
            <div className="col pr1 margin-top-m">

<h2 className="display-title margin-top-xl">Before you begin</h2>
            <p className="lead max-width-xs">These are the things to purchase before you start, which will cost <strong>{this.state.totalCost}</strong>.</p>

           

            <div className="boxes">
                {listItems}
            </div>
</div>
<div className="col margin-top-m">

            <h2 className="display-title margin-top-xl">Directions</h2>
            <p className="lead max-width-xs">Follow these steps to make this recipe, which will take about <strong>{this.state.totalTime}</strong>.</p>
            {listDirections}
            <div>
              <div className="outsidebox">
                  <h2>Enjoy!</h2>
              </div>
            </div>

            </div>
            </div>

          </div>
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
                <a href="https://github.com/schollz/recursive-recipes/issues">Issues</a>
            </nav>
        </div>
    </footer>
      </div>
    );
  }
}

export default App;
