
import { useState, useEffect} from 'react';
import './App.css';
import { Button } from 'react-bootstrap';

function LoginStatus() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    fetch('/login/in/')
      .then(response => setIsLoggedIn(false))
      // .then(data => setIsLoggedIn(true));
  }, []);

  return (
    <div>
      {isLoggedIn ? (
        <p>You are logged in</p>
      ) : (
        <Button variant="primary" href="/login/twitch/">
          Login with Twitch
        </Button>
      )}
    </div>
  );
}
export default LoginStatus;
