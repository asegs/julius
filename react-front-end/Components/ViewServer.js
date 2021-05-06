import React,{Component} from 'react';
import CryptoJS from 'crypto-js';
import { animateScroll } from "react-scroll";

class ViewServer extends Component{
    constructor(props) {
        super(props);
        this.state = {
            url : this.props.url,
            cleanText : this.props.cleanText,
            socket: undefined,
            cipherText : "",
            messages:[],
            message:"",
            key:"",
            username:"",
            password:"",
            gettingOlder:false
        }
        this.connectToServer = this.connectToServer.bind(this);
        this.getOlderMessages = this.getOlderMessages.bind(this);
    }

    encrypt=(word,key)=>{
        return CryptoJS.AES.encrypt(word,key).toString();
    }

    decrypt=(word,key)=>{
        return CryptoJS.AES.decrypt(word,key).toString(CryptoJS.enc.Utf8);
    }

    setCode=(e)=>{
        this.setState({cipherText:this.g(this.state.cleanText,e.target.value,true),key:e.target.value});
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


    async connectToServer(){
        if (this.state.socket === undefined){
            let added = '/ws/'+this.state.url+'/'+this.state.cipherText;
            let url = window.location.hostname === "localhost" ? 'ws://localhost:3005'+added : 'ws://softsort.org:3009'+added;
            this.setState({socket:new WebSocket(url)},()=>{
                this.state.socket.addEventListener('message',(e)=>{
                    let msg = JSON.parse(e.data);
                    if (Array.isArray(msg)){
                        let newMessages = [];
                        for (let i = 0;i<msg.length;i++){
                            let m = msg[i];
                            if (m['message'] !== ""){
                                newMessages.push(this.decryptMessage(m));
                            }else{
                                break;
                            }
                        }
                        this.setState({messages:newMessages});
                    }else{
                        let oldMessages = this.state.messages;
                        oldMessages.push(this.decryptMessage(msg));
                        this.setState({messages:oldMessages});
                    }
                })
            });
        }
    }

    sendChatAction=()=>{
        if (this.state.message.length===0){
            return ;
        }
        let data = {
            username:this.state.username,
            message:{
                username:this.state.username,
                message: this.encrypt(this.state.message,this.state.key),
                server:this.state.url
            },
            password:this.state.password
        }
        this.state.socket.send(
            JSON.stringify(data)
        );
        this.setState({message:""});
    }

    changeKey=(e)=>{
        this.setState({key:e.target.value});
    }

    changeMsg=(e)=>{
        this.setState({message:e.target.value});
    }

    changeUser=(e)=>{
        this.setState({username:e.target.value});
    }

    hash = (word)=>{
        return CryptoJS.MD5(word).toString();
    }

    changePass=(e)=>{
        this.setState({password:this.hash(e.target.value)});
    }

    sendIfEnter=(e)=>{
        if (e.key==="Enter"){
            this.sendChatAction();
        }
    }

    componentDidMount() {
        this.scrollToBottom();
    }
    // componentDidUpdate() {
    //     if (this.state.gettingOlder){
    //        this.state.gettingOlder = false;
    //     }else{
    //         this.scrollToBottom();
    //     }
    // }
    componentDidUpdate(prevProps, prevState, snapshot) {
        if (prevState.messages !== this.state.messages){
            if (this.state.gettingOlder){
                this.setState({gettingOlder:false});
            }else{
                this.scrollToBottom();
            }
        }
    }

    scrollToBottom() {
        animateScroll.scrollToBottom({
            containerId: "scroller",
            delay:0,
            duration:150
        });
    }

    decryptMessage=(msg)=>{
        let text = this.decrypt(msg['message'],this.state.key);
        return {'message':text,'username':msg['username'],'server':msg['server']}
    }

    async getOlderMessages(){
        let data = {};
        data['server'] = this.state.url;
        data['cipher_text'] = this.state.cipherText;
        data['offset'] = this.state.messages.length;
        let url = window.location.hostname === "localhost" ? 'http://localhost:3005/older' : 'http://softsort.org:3009/older';
        const response = await fetch(url,{
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data),
        })
        const body = await response.json();
        for (let i = 0;i<body.length;i++){
            body[i] = this.decryptMessage(body[i]);
        }
        let oldMessages = this.state.messages;
        body.push(...oldMessages);
        this.setState({messages:body,gettingOlder:true});
    }

    render() {
        if (this.state.socket === undefined){
            return(
                <div>
                    <h2 style={{fontFamily: 'courier', alignText: 'left', paddingLeft: '15%',"cursor":"pointer",color:"white"}}>julius</h2>
                    <label htmlFor={"key_space"} style={{color:"white"}}>Enter the private key you received: </label>
                    <input type={"text"} id={"key_space"} style={{"width":"20%",borderRadius:"2px"}} onChange={this.setCode} /><br/>
                    <button onClick={this.connectToServer}>Load messages</button>
                    <div id={"scroller"}></div>
                </div>
            );

        }else{
            let b= <button style={{marginLeft:"43%"}} onClick={this.getOlderMessages}>Get older messages...</button>
            let messageText = this.state.messages.length>=25 ? [b] : [];
            for (let i=0;i<this.state.messages.length;i++){
                let m = this.state.messages[i];
                if (m['message']===""){
                    break;
                }
                let color = this.state.username === m['username'] ? "dodgerblue" : "gray"
                messageText.push(<p style={{"backgroundColor":color,"maxWidth":"60%","margin":"auto","marginBottom":"5px","marginTop":"5px","borderRadius":"2px"}}>{m['username']+": "+m['message']}</p>)
            }
            return (
                <div>
                    <h2 style={{fontFamily: 'courier', alignText: 'left', paddingLeft: '15%',"cursor":"pointer",color:"white"}}>julius</h2>
                    <label htmlFor={"key-box"} style={{color:"white"}}>Decryption key: </label>
                    <input type={"text"} id={"key-box"} onChange={this.changeKey} value={this.state.key} style={{borderRadius:"2px"}}/>
                    <label htmlFor={"user"} style={{color:"white"}}>Username: </label>
                    <input type={"text"} id={"user"} onChange={this.changeUser} maxLength={32} style={{borderRadius:"2px"}}/>
                    <label htmlFor={"password"} style={{color:"white"}}>Password: </label>
                    <input type={"password"} id={"password"} onChange={this.changePass} maxLength={64} style={{borderRadius:"2px"}}/><br/>
                    <div id={"scroller"} className={"scroll"}>
                        {messageText}
                    </div>
                    <label htmlFor={"message-box"} style={{color:"white"}}>Message:</label>
                    <input type={"text"} style={{"width":"50%",borderRadius:"2px"}} id={"message-box"} onChange={this.changeMsg} value={this.state.message} onKeyDown={this.sendIfEnter} maxLength={2048}/>
                    <button onClick={this.sendChatAction}>Send</button>
                </div>
            )
        }
    }
}
export default ViewServer;
