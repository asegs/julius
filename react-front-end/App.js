import './App.css';
import React from 'react';
import DualRouter from "./Components/DualRouter";
import {BrowserRouter as Router,Route, Switch} from 'react-router-dom';


// class App extends Component{
//   constructor(props) {
//     super(props);
//     this.state = {
//       url:"",
//       showingRouter:false
//     }
//   }
//
//   changeURL =e=>{
//     this.setState({url:e.target.value});
//   }
//
//   toggle=()=>{
//     this.setState({showingRouter:!this.state.showingRouter});
//   }
//
//   render() {
//
//     if (this.state.showingRouter) {
//       return (
//           <div className={"App"}>
//
//             <DualRouter url={this.state.url}></DualRouter>
//           </div>
//
//       )
//     } else {
//
//       return (
//           <div className={"App"}>
//             <h2 style={{fontFamily: 'courier', alignText: 'left', paddingLeft: '15%',color:"white"}}>julius</h2>
//             <label htmlFor={"url-finder"} style={{color:"white"}}>Server location: </label>
//             <input type={"text"} id={"url-finder"} onChange={this.changeURL}/>
//             <button onClick={this.toggle}>Visit</button>
//           </div>
//       );
//     }
//   }
// }

export default function App(){
    return (
        <div className={"App"}>
           <Router>
               <Switch>
                   <Route path='/server/:location' component={DualRouter}/>
               </Switch>
           </Router>
        </div>
    );
}

