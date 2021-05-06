import React,{Component} from 'react';
import CryptoJS from 'crypto-js';
class CreateServer extends Component{
    constructor(props) {
        super(props);
        this.state = {
            url : this.props.url,
            nickname:"",
            sample:"",
            username:"",
            password:"",
            key:"",
        }
        this.createNewServer = this.createNewServer.bind(this);
    }
    changeNickname = (e)=>{
        this.setState({nickname:e.target.value});
    }

    changeSample = (e)=>{
        this.setState({sample:e.target.value});
    }

    changeUsername = (e)=>{
        this.setState({username:e.target.value});
    }

    changePassword = (e)=>{
        this.setState({password:e.target.value});
    }

    changeKey = (e)=>{
        this.setState({key:e.target.value});
    }


    allFilled=()=>{
        let valid = true;
        if (this.state.nickname.length===0){
            valid = false;
        }
        if (this.state.username.length===0){
            valid = false
        }
        if (this.state.password.length<6){
            valid = false;
        }
        if (this.state.sample.length===0){
            valid = false;
        }
        if (this.state.key.length<8){
            valid = false;
        }
        return valid;
    }

    hashPassword = ()=>{
        return CryptoJS.MD5(this.state.password);
    }

    async createNewServer(){
        let data = {};
        data['url'] = this.state.url;
        data['name'] = this.state.nickname;
        data['username'] = this.state.username;
        data['password'] = this.hashPassword().toString();
        data['clean_text'] = this.state.sample;
        data['cipher_text'] = this.g(this.state.sample,this.state.key,true);
        console.log(data)
        let url = window.location.hostname === "localhost" ? 'http://localhost:3005/create' : 'http://softsort.org:3009/create';
        const response = await fetch(url,{
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data),
        })
        const body = response.json();
        if (body){
            this.props.update(this.state.sample);
            this.props.toggle();
        }
    }

    toHex =num=>{
        return num.toString(16);
    }

    fromHex =hex=>{
        return parseInt(hex,16);
    }

    g =(word,key,e)=>{
        let result = "";
        for (let i=0;i<word.length;i+=e ? 1 : 2){
            let charWord = e ? word.charCodeAt(i) : this.fromHex(word.substring(i,i+2));
            let charKey = key.charCodeAt((e ? i : i/2)%(key.length-1));
            let num = e ? charWord+charKey : charWord-charKey;
            let fmtVersion = e ? this.toHex(num) : String.fromCharCode(num);
            result+=fmtVersion;
        }
        return result;
    }


    render() {
        let b = this.allFilled() ? <button onClick={this.createNewServer}>Create!</button> : "";
        return(
            <div>
                <h2 style={{fontFamily: 'courier', alignText: 'left', paddingLeft: '15%',"cursor":"pointer",color:"white"}} >julius</h2>
                <h3>Create a private server at this URL:</h3>
                <h4>{window.location.href+this.state.url}</h4><br/>
                <label htmlFor={"server_name"} style={{color:"white"}}>Server nickname: </label><br/>
                <input id={"server_name"} style={{"width":"20%","marginBottom":"20px"}} type={"text"} onChange={this.changeNickname} value={this.state.nickname}/><br/>
                <label htmlFor={"server_username"} style={{color:"white"}} >Your username on this server: </label><br/>
                <input id={"server_username"} style={{"width":"20%","marginBottom":"20px"}} type={"text"} onChange={this.changeUsername} value={this.state.username}/><br/>
                <label htmlFor={"server_password"} style={{color:"white"}} >Your password on this server: </label><br/>
                <input id={"server_password"} style={{"width":"20%","marginBottom":"20px"}} type={"password"} onChange={this.changePassword} value={this.state.password}/><br/>
                <label htmlFor={"sample_word"} style={{color:"white"}}>Sample word to encrypt (used for validating new users)</label><br/>
                <input id={"sample_word"} style={{"width":"20%","marginBottom":"20px"}} type={"text"} onChange={this.changeSample} value={this.state.sample}/><br/>
                <label htmlFor={"server_key"} style={{color:"white"}}>Server encryption key (only used locally): </label><br/>
                <input id={"server_key"} style={{"width":"30%","marginBottom":"20px"}} type={"text"} onChange={this.changeKey} value={this.state.key}/><br/>
                {b}
            </div>
        );
    }
}
export default CreateServer;
