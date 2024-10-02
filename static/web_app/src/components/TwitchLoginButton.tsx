import { useAuth } from "../contexts/auth";
import './TwitchLoginButton.css';


// https://www.vectorlogo.zone/logos/twitch/
export function TwitchLoginButton() {
    const auth = useAuth();
    const handleClick = () => {auth.Login();}

    return (
      <div className="container d-flex flex-row px-0">
        <button className="btn-twitch flex-fill" onClick={handleClick}>
          <div className="btn-twitch-logo"></div>
          Login with Twitch
        </button>
      </div>
    );
}