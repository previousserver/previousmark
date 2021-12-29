import React, { useState, useEffect, Component } from 'react';
import ReactDOM from 'react-dom';
import { BrowserRouter, Route, Routes, Link } from 'react-router-dom';
import {button} from 'bootstrap/dist/css/bootstrap.min.css';
import {Modal, Button} from 'react-bootstrap';
import logo from './logo.svg';
import user from './user.svg';
import borgar from './borgar.svg';
import './App.css';


const back = "localhost:8080";

function getWindowDimensions() {
  const { innerWidth: width, innerHeight: height } = window;
  return {
    width,
    height
  };
}

function useWindowDimensions() {
  const [windowDimensions, setWindowDimensions] = useState(getWindowDimensions());

  useEffect(() => {
    function handleResize() {
      setWindowDimensions(getWindowDimensions());
    }

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return windowDimensions;
}


// Header

function Header(props) {
	const {height, width} = useWindowDimensions();
  	if (width <= height) {
  		return (ResponsiveHeader(props));
  	}
  	return (StandardHeader(props));
}

function StandardHeader(props) {
	if (props.isReg) {
		return (
			<div className="App-header">
				<p>previousmark	<Link className="App-header-link" to='/benchmarks'>Benchmarks</Link>
				<Link className="App-header-link" to='/users'>Users</Link>
				<Link className="App-header-link" to='/me'><button type="button" className="App-button"><img src={user} className="App-header-logo" alt="User"/></button></Link>
				<Link className="App-header-link" to='/logout'><button type="button" className="App-button">Logout</button></Link></p>
			</div>
		);
	}
	return (
			<div className="App-header">
				<p>previousmark	<Link className="App-header-link" to='/benchmarks'>Benchmarks</Link>
				<Link className="App-header-link" to='/users'>Users</Link>
				<Link className="App-header-link" to='/login'><button type="button" className="App-button">Login</button></Link>
				<Link className="App-header-link" to='/register'><button type="button" className="App-button">Register</button></Link></p>
			</div>
	);
}

function ResponsiveBurger(props) {
	if (props.isReg) {
		return (
			<div id="header">
			previousmark	<button className="App-button" onClick={ResponsiveBurgerReg}><img src={borgar} className="App-header-menu" alt="Menu"/></button>
			</div>
		);
	}
	return (
		<div id="header">
		previousmark	<button className="App-button" onClick={ResponsiveBurgerGuest}><img src={borgar} className="App-header-menu" alt="Menu"/></button>
		</div>
	);
}

function ResponsiveBurgerReg() {
	return (
		ReactDOM.render(
		<React.Fragment>
		<BrowserRouter>
		<StandardHeader isReg={true} />
		</BrowserRouter>
		</React.Fragment>,
		document.getElementById('header')
	));
}

function ResponsiveBurgerGuest() {
	return (
		ReactDOM.render(
		<React.Fragment>
		<BrowserRouter>
		<StandardHeader />
		</BrowserRouter>
		</React.Fragment>,
		document.getElementById('header')
	));
}
		
function ResponsiveHeader(props) {
	if (props.isReg) {
		return (
			<div className="App-header">
				<ResponsiveBurger isReg={true} />
			</div>
		);
	}
	return (
		<div className="App-header">
			<ResponsiveBurger isReg={false} />
		</div>
	);
}


// Content

class ContentDefault extends Component {
  render() {
	return (
		<div>
			<img src={logo} className="App-logo" alt="logo" />
			<p>Welcome to <b>previousmark</b>, the alternative benchmark leaderboard!</p>
		</div>
	);
  }
}

class Login extends Component {
  constructor(props) {
	super(props);
	this.nickname = "";
	this.password = "";
	this.windowDimensions = this.windowDimensions.bind(this);
  }
  windowDimensions = () => {
  	const { innerWidth: width, innerHeight: height } = window;
  	return {width, height};
  }
  componentDidMount() {
    const requestOptions = {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Accept': '' }
    };
//    fetch('http://localhost:8080/api/auth')
//        .then(response => response.json())
//        .then(data => this.setState({}));
  }
  render() {
	let {height, width} = this.windowDimensions();
	if (width / height >= 4/3) {
		return (
			<form type="submit">
				<label for="nick"><b>Nickname:</b></label> <input type="text" id="nick" name="nick" /> <label for="pass"><b>Password:</b></label> <input type="password" id="pass" name="pass" /> <LoginModal/>
			</form>
		);
	}
	return (
		<form type="submit">
		<label for="nick"><b>Nickname:</b></label>‏‏‎ ‎<input type="text" id="nick" name="nick" />
		<br/><label for="pass"><b>Password:</b></label>‏‏‎‏‏‎ ‎‏‏‎ ‎ ‎<input type="password" id="pass" name="pass" />
		<br/><LoginModal/>
		</form>
	);
  }
}

function LoginModal(props) {
  const [show, setShow] = useState(false);

  const handleClose = () => setShow(false);
  const handleShow = () => setShow(true);

  return (
    <>
      <Button variant="dark" size="sm" onClick={handleShow}>
	Login
      </Button>

      <Modal show={show} onHide={handleClose}>
        <Modal.Header closeButton>
          <Modal.Title>Error</Modal.Title>
        </Modal.Header>
        <Modal.Body>You are not logged in, your nickname and/or password is invalid, or your token is invalid. Your session might have expired. Please log in or modify request to continue!</Modal.Body>
        <Modal.Footer>
          <Button variant="primary" onClick={handleClose}>
            OK
          </Button>
        </Modal.Footer>
      </Modal>
    </>
  );
}


// Footer

function Footer() {
  return (
    <footer className="App-footer">
      <p>
	{new Date().getUTCFullYear()} All rights belong to their respective owners. Font: <a href="http://www.vlkk.lt/palemonas" className="Link">Palemonas</a>
      </p>
    </footer>
  );
}


// Entry

function HeaderContent() {
  return (
	<BrowserRouter>
		<div>
        		<Header isReg={true} />
        	</div>
		<div className="App-content">
			<Routes>
				<Route path='/' element={<ContentDefault/>} />
      				<Route path='/benchmarks' />
				<Route path='users' />
				<Route path='me' />
				<Route path='logout' />
				<Route path='login' element={<Login/>} />
				<Route path='register' />
			</Routes>
		</div>
    	</BrowserRouter>
  );
}


class App extends Component {
  render() {
  return (
	<div className="App">
		<div>
			<HeaderContent />
		</div>
    		<div><p><Footer /></p></div>
	</div>
  );
  }
}

export default App;
