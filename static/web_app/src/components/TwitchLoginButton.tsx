import { useLocation } from "react-router-dom";
import { useCookies } from "react-cookie";

import { useAuth } from "../contexts/auth";
import './TwitchLoginButton.css';


// https://www.vectorlogo.zone/logos/twitch/
export function TwitchLoginButton(props: {
  redirect_to: string
}) {
  const auth = useAuth();
  const [cookies, setCookie, removeCookie] = useCookies(['redirect_to']);
  
  const handleClick = () => {
    console.log("@@@@ from = ", props.redirect_to);
    setCookie('redirect_to', props.redirect_to, {
      path: '/',
      expires: new Date(Date.now() + 5 * 60 * 1000),
      secure: true,
      sameSite: true,
    });
    auth.Login();
  }

  return (
    <div className="container d-flex flex-row px-0">
      <button className="btn-twitch flex-fill" onClick={handleClick}>
        <div className="btn-twitch-logo"></div>
        Login with Twitch
      </button>
    </div>
  );
}