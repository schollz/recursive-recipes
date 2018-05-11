import React, { Component } from 'react';
import ReactDOM from 'react-dom';
import Slider, { Range } from 'rc-slider';
// We can just import Slider or Range to reduce bundle size
// import Slider from 'rc-slider/lib/Slider';
// import Range from 'rc-slider/lib/Range';
import 'rc-slider/assets/index.css';
import './index.css'

const marks = {
  0: '1 minute',
  1: '10 minutes',
  2: '20 minutes',
  3: '30 minutes',
  6: '1 hour',
  12: '2 hours',
}
class App extends Component {

    constructor(props) {
      super(props);
      this.state = {
        version: "v0.0.0",
        recipe: "Chocolate Chip Cookies",
        totalCost: "$2.30",
        limitfactor: 0,
        ingredients: [
          {
            amount: "1 1/2 cup",
            name: "Flour",
            cost: "$1.00",
            scratchTime: "+2 hours",
            scratchCost: "-$1.00",
          }
        ]
      };
    }

handleOnChange(value) {
    console.log(value);
    this.setState({
      volume: value,
      ingredients: [
        {
          name:"ice",
          amount: "1 1/2 cup",
          cost: "",
        }
      ]
    })
  }

  render() {
    const listItems = this.state.ingredients.map((ing) =>
    <div class="box">
    <h3>
    <span className="small-caps">{ing.amount}</span>
    <span className="display-block">{ing.name} {ing.cost != '' &&
    <small>({ing.cost})</small> 
      }</span>
    </h3>
    
      {ing.scratchCost != '' &&
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
    <small>3 days, 2 hours</small>
    </h2>
    <div>
<div>
Time limit:
<div className="slider">
 <Slider className="slider" min={0} max={12} marks={marks} step={null} onChange={this.handleOnChange.bind(this)} defaultValue={20} />
 </div>
   </div>

      </div>

<h2 className="display-title margin-top-xl">Before you begin</h2>
            <p className="lead max-width-xs">These are the things to purchase before you start, which will cost <strong>{this.state.totalCost}</strong>.</p>

           

            <div className="boxes margin-top-m">
                {listItems}
                <div className="box">
                    <h3>
        <span className="small-caps">1 1/2 cup</span>
        <span className="display-block">Flour</span>
        </h3>
                    <p>
                      alskdjf
                    </p>
                </div>
                <div className="box">
                    <h3>
        <span className="small-caps">1 1/2 cup</span>
        <span className="display-block">Chocolate Chips</span>
        </h3>
                    <p>
                    </p>
                </div>
                <div className="box">
                    <h3>
        <span className="small-caps">1 whole</span>
        <span className="display-block">Egg Laying Chicken</span>
        </h3>
                </div>
                <div className="box">
                    <h3>
        <span className="small-caps">1.3 acre</span>
        <span className="display-block">Soil</span>
        </h3>
                </div>
            </div>



          </div>
        </main>
      </div>
    );
  }
}

export default App;
