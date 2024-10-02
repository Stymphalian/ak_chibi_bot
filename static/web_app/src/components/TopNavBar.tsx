
import { NavLink, useLocation } from "react-router-dom";
import { AuthStatus } from "../contexts/auth";

export function TopNavBar() {
    const location = useLocation();
    const isNotLoginPage = location.pathname !== "/login";
    console.log(location);
    return (
        <nav id="topnavbar" className="navbar navbar-light bg-light">
            <div className="container-fluid">
                <NavLink className="navbar-brand" to="/">Home</NavLink>
                <NavLink className="nav-link me-2" to="/docs">Docs</NavLink>
                <NavLink className="nav-link me-auto" to="/settings">Settings</NavLink>
                
                {isNotLoginPage && 
                <div className="nav-item">
                    <AuthStatus></AuthStatus>
                </div>
                }
            </div>
        </nav>
    );
}