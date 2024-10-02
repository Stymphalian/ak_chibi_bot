
import { useLocation } from "react-router-dom";
import { AuthStatus } from "../contexts/auth";

export function TopNavBar() {
    const location = useLocation();
    const isNotLoginPage = location.pathname !== "/login";
    console.log(location);
    return (
        <nav id="topnavbar" className="navbar navbar-light bg-light">
            <div className="container-fluid">
                <a className="navbar-brand" href="/">Home</a>
                <a className="nav-link me-2" href="/docs">Docs</a>
                <a className="nav-link me-auto" href="/settings">Settings</a>
                
                {isNotLoginPage && 
                <div className="nav-item">
                    <AuthStatus></AuthStatus>
                </div>
                }
            </div>
        </nav>
    );
}