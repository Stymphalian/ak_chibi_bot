import { NavLink } from "react-router-dom";
import { Code } from "../components/Code"
import { useAuth } from "../contexts/auth";

export function DocsPage() {
    const auth = useAuth();
    let channelName = auth.userName || "REPLACE_ME";
    let url = `https://akchibibot.stymphalian.top/rooms?channelName=${channelName}`

    return (
        <>
            <div className="display-5 fw-semibold">Docs</div>
            <table className="table">
                <thead>
                    <tr>
                        <th scope="col">Commands</th>
                        <th scope="col">Description</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>!chibi help</td>
                        <td>Displays a help message in chat for how to use the bot</td>
                    </tr>
                    <tr>
                        <td>!chibi Amiya</td>
                        <td>Changes the User's current chibi to Amiya. The bot does fuzzy matching on the operator's names.</td>
                    </tr>
                    <tr>
                        <td>!chibi skins</td>
                        <td>List the available skins for the chibi. The default skin is <code>default</code>. Other skins are identified by the fashion brand (i.e <code>epoque</code>, <code>winter</code>, etc)</td>
                    </tr>
                    <tr>
                        <td>!chibi anims</td>
                        <td>List the available animations for the chibi. Normal idle animations are <code>Relax</code>, <code>Idle</code>, <code>Sit</code> and <code>Move</code>. Operators with skins also have a <code>Special</code> animation that can be played.</td>
                    </tr>
                    <tr>
                        <td>!chibi face [front or back]</td>
                        <td>Only works in battle stance. Make the chibi face forward or backwards. <Code>!chibi face back</Code></td>
                    </tr>
                    <tr>
                        <td>!chibi stance [base or battle]</td>
                        <td>Set the chibi to their base model (ie. the one which walks around the factory/dorms) or the battle map model</td>
                    </tr>
                    <tr>
                        <td>!chibi skin epoque</td>
                        <td>Change the skin to epoque or any other skin name.</td>
                    </tr>
                    <tr>
                        <td>!chibi play Attack</td>
                        <td>Change the animation the chibi currently plays. The animation loops forever.</td>
                    </tr>
                    <tr>
                        <td>!chibi walk</td>
                        <td>Have the chibi walk back and forth across the screen. Optionally specify a number between 0 and 1.0 (<Code>!chibi walk 0.7</Code>) for the chibi to walk to that part of the screen.</td>
                    </tr>
                    <tr>
                        <td>!chibi enemy The Last Steam Knight</td>
                        <td>Change the chibi into one of the enemies in the game. You can also use the enemies "code". !chibi enemy B4 is equivalent to !chibi enemy Originium Slug. You can find all the enemy codes from <a href="https://arknights.wiki.gg/wiki/Enemy/Normal">Arknights Wiki</a>.</td>
                    </tr>
                    <tr>
                        <td>!chibi who Rock</td>
                        <td>Ask the bot for operators/enemies which match the given name.</td>
                    </tr>
                    <tr>
                        <td>!chibi size 0.4</td>
                        <td>Change the size/scale of the chibi. (min 0.5, max 1.5). There is still a maximum pixel size limitation on the chibi (350px)</td>
                    </tr>
                    <tr>
                        <td>!chibi velocity 1.2</td>
                        <td>For when the chibi is walking around change the movement speed of the chibi (min 0.1, max 2). Use <code>!chibi velocity default</code> to change back to a default speed.</td>
                    </tr>
                </tbody>
            </table>

            <div className="m-2"></div>

            <div className="display-5 fw-semibold mb-3">Additional Notes</div>
            <ul className="list-group">
                <li className="list-group-item">
                    For cases where your OBS viewport size is not exactly <Code>1920x1080</Code>. 
                    You can pass additional query arguments to the URL to set the width and height of the Browser Source. <br />
                    For example if your monitor is <Code>1080x720</Code> you can use this URL instead: <br />
                    <Code>{url + "&width=1080&height=720"}</Code>
                </li>
                <li className="list-group-item">
                    You can change settings related to how your bot handles 
                    (<Code>!chibi size, !chibi speed, !chibi velocity</Code>) commands 
                    by going to <NavLink to="/settings">Settings</NavLink>. On that page you can modify the min/max values which are accepted by the commands. 
                    The default values on that page are the normal defaults. Be sure to refresh the Browser Source to have the changes take effect in chat.
                </li>
            </ul>
        </>
    )
}