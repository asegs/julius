import React, {Component} from 'react'
import CreateServer from "./CreateServer";
import ViewServer from "./ViewServer";

class DualRouter extends Component{
    constructor(props) {
        super(props);
        console.log("CREATED")
        this.state = {
            url:this.props.match.params.location,
            child:undefined,
            createNew:true,
            cleanText:"",
        }
    }

    async componentDidMount() {
        let data = {};
        data['url'] = this.state.url;
        let url = window.location.hostname === "localhost" ? 'http://localhost:3005/access' : 'http://softsort.org:3009/access';
        const response = await fetch(url,{
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data),
        })
        const body = await response.json();

        if (!body){
            this.setState({child:<CreateServer url={this.state.url} toggle={this.toggle} update={this.updateCleanText}/>,createNew:true});
        }else{
            this.setState({child : <ViewServer url={this.state.url} cleanText={body} toggle={this.toggle} update={this.updateCleanText}/>,createNew:false,cleanText:body});
        }
    }

    toggle=()=>{
        if (this.state.createNew){
            this.setState({child : <ViewServer url={this.state.url} cleanText={this.state.cleanText} toggle={this.toggle} update={this.updateCleanText}/>,createNew:false,cleanText:this.state.cleanText});
        }else{
            this.setState({child:<CreateServer url={this.state.url} toggle={this.toggle} update={this.updateCleanText}/>,createNew:true});
        }
    }

    updateCleanText=(word)=>{
        this.setState({cleanText:word});
    }

    render() {
        return (
            <div>
                {this.state.child}
            </div>
        );
    }
}
export default DualRouter;
