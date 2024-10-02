import { Container } from "react-bootstrap";
import { NavLink } from "react-router-dom";
import { Banner } from "../components/Banner";
import { Code } from "../components/Code";
import { useAuth } from "../contexts/auth";

import demo_img from '../assets/demo1.png';

export function HomePage() {
    const auth = useAuth();
    let channelName = auth.userName || "REPLACE_ME";
    let url = `https://akchibibot.stymphalian.top/rooms?channelName=${channelName}`

    return (
        <Container className="px-3">
            <Banner />

            <div className="card">
                {/* <img src={demo_img} className="card-img-top" width="600px" height="400px" /> */}
                <div className="container d-flex justify-content-center card-img-top">
                    <img src={demo_img} className="col-6" width="100%" />
                </div>

                <div className="card-body">
                    <h3 className="card-header">Getting Started</h3>
                    
                    <ul className="list-group list-group-flush">
                        <li className="list-group-item">
                            Open up OBS and add a <Code>Browser Source</Code> to your stream.
                        </li>
                        <li className="list-group-item">
                            Set the URL to <Code>{url}</Code>
                        </li>
                        {!auth.isAuthenticated &&
                            <li className="list-group-item">
                                Change the <Code>REPLACE_ME</Code> part of the URL to your own channel name (use lowercase). <br />
                                For example something like <Code>?channelName=stymphalian2__</Code>
                            </li>
                        }
                        <li className="list-group-item">
                            Set the <Code>width</Code> and <Code>height</Code> to <Code>1920x1080</Code>.  <br />
                            Also set the <Code>Shutdown source when not visible</Code> option to true.
                        </li>
                        <li className="list-group-item">
                            The OBS preview window should now show a chibi sitting on your stream.
                        </li>
                        <li className="list-group-item">
                            Open up your twitch chat and start typing <Code>!chibi</Code> commands. <br />
                            You can use <Code>!chibi help</Code> to see the list of commands. <br />
                            A full list of command can be seen in the <NavLink to="/docs">Docs</NavLink> page.
                        </li>
                    </ul>
                    
                </div>
            </div>

        </Container>
    );
}
