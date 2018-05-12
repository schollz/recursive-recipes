import React, { Component } from 'react';
import Slider from 'rc-slider';
import Sockette from 'sockette';
// We can just import Slider or Range to reduce bundle size
// import Slider from 'rc-slider/lib/Slider';
// import Range from 'rc-slider/lib/Range';
import 'rc-slider/assets/index.css';
import './index.css'

var moment = require("moment");
var momentDurationFormatSetup = require("moment-duration-format");


class App extends Component {

    constructor(props) {
      super(props);
      this.timeout = null;
      this.ws = new Sockette('ws://localhost:8012/ws', {
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
      
      this.state = {
        websocketURL: "ws://localhost:8080/ws",
        version: "v0.0.0",
        recipe: "Chocolate Chip Cookies",
        totalCost: "$2.30",
        totalTime: "3 days, 2 hours",
        limitfactor: 0,
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
        },
        ]
      };
    }


  handleData(data) {
    console.log(data);
    let result = JSON.parse(data.data);
    console.log(result);
    // this.setState({
    //   limitfactor:10,
    // });
  }


  handleOnChange(value) {
    clearTimeout(this.timeout);
    this.timeout = setTimeout((function(){
      this.ws.send("hello there");
      this.ws.send(value);
    }).bind(this),250);
  
    this.setState({
      limitfactor: value,
    })
  }

  render() {
    const listDirections = this.state.directions.map((direction) =>
    <div>
      <div className="outsidebox">
          <h2>Make the {direction.name}</h2>
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
    <span className="display-block">{ing.name}</span>
    </h3>
    
      {ing.scratchCost !== '' &&
      <p>{ing.scratchCost}, {ing.scratchTime} to make {ing.name.toLowerCase()} from scratch.</p>
      }
    </div>
  );
return (
      <div className="App">
        <header className="padding-top-xs text-center color-white background-primary">
            <div className="container">
                <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" className="feather feather-book-open">
                    <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path>
                    <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path>
                </svg>
                <h1 className="display-title">Recursive Cookbook</h1>
                <p className="no-margin"><strong>{this.state.version}</strong></p>
            </div>
        </header>
        <main className="padding-vertical-xl color-white background-primary">

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


<h2 className="display-title margin-top-xl">Before you begin</h2>
            <p className="lead max-width-xs">These are the things to purchase before you start, which will cost <strong>{this.state.totalCost}</strong>.</p>

           

            <div className="boxes margin-top-m">
                {listItems}
            </div>


            <h2 className="display-title margin-top-xl">Directions</h2>
            <p className="lead max-width-xs">Follow these steps to make this recipe, which will take about <strong>{this.state.totalTime}</strong>.</p>
            {listDirections}


          </div>
        </main>

      </div>
    );
  }
}

export default App;
