import React from "react";
import { Button, Container, Form } from "react-bootstrap";
import { TwitchLoginButton } from "../components/TwitchLoginButton";
import './Login.css';
import yato_img from './../assets/yato.png';
import { Banner } from "../components/Banner";

export function LoginPage() {
  return (
    <>
    <Banner />
    <div className="container-fluid d-flex justify-content-center bg-light align-items-center justify-content-center">

      <div className="flex-fill col-2"></div>

      <div className="container d-flex col-2 justify-content-center">
        <img src={yato_img} width="104px" height="147px"className="yato-image" />
      </div>

      <div className="container col-3 d-block bg-white m-2 p-4 shadow-sm rounded">
        <p className="center-me">
          Login to Twitch to customize the Bot setttings 
          specific to your own channel. Sign in to confirm that you are the 
          owner of the channel. We don't need your email just your login name.
        </p>
        <hr className="opacity-10"/>
        <div>
          <TwitchLoginButton />
        </div>
      </div>

      <div className="flex-fill col-2"></div>
  
    </div>
    </>
  );
}
