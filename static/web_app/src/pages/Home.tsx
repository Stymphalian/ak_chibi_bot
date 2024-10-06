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
                            A full list of commands can be seen in the <NavLink to="/docs">Docs</NavLink> page.
                        </li>
                    </ul>
                    
                </div>
            </div>

            <div className="card mt-3">
                <div className="card-body">
                    <h3 className="card-header">Disclaimer</h3>
                    <ul className="list-group list-group-flush">
                        <li className="list-group-item">
                            All the art assets/chibis are owned by Arknights/Hypergryph. <br />
                            I claim no ownership and the use of their assets <br />
                            are purely for personal enjoyment and entertainment.
                        </li>
                        <li className="list-group-item">
                            The software used for rendering the chibis (i.e Spine models) <br />
                            use the Esoteric runtime libraries which is under <a href="http://esotericsoftware.com/spine-editor-license">License</a>. <br />
                            Strictly speaking use of this software requires each individual to have purchased their own software license.
                        </li>
                    </ul>
                </div>
            </div>

        </Container>
    );
}
