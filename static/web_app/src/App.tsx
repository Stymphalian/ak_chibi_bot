import React from 'react';
import logo from './logo.svg';
import './App.css';
import { Container, Navbar, Nav, NavDropdown } from 'react-bootstrap';
import RoomSettings from './RoomSettings';
import LoginStatus from './LoginStatus';

function App() {
  return (
    <>
      <Navbar expand="lg" className="bg-body-tertiary">
        <Container>
          <Navbar.Brand href="#home">AK Chibi Bot</Navbar.Brand>
          <Navbar.Toggle aria-controls="basic-navbar-nav" />
          <Navbar.Collapse className="justify-content-end" id="basic-navbar-nav">
            <Navbar.Text>
              <LoginStatus></LoginStatus>
            </Navbar.Text>
            
          </Navbar.Collapse>
        </Container>
      </Navbar>
      <Container className="p-3">
        <Container className="p-5 mb-4 bg-light rounded-3">
          <h1 className="header">Arknights Chibi Bot</h1>
          <p>A twitch bot and browser source overlay to show AK chibis walking on your stream</p>
        </Container>
      </Container>
    </>
  );
}

export default App;
