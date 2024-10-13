
import { NavLink, useLocation } from "react-router-dom";
import { AuthStatus, useAuth } from "../contexts/auth";
import "./TopNavBar.css";

export function TopNavBar() {
    const auth = useAuth();
    const location = useLocation();
    const isNotLoginPage = !location.pathname.startsWith("/login");
    return (
        <nav id="topnavbar" className="navbar navbar-light bg-light">
            <div className="container-fluid">
                <NavLink className="navbar-brand" to="/">Home</NavLink>
                <NavLink className="nav-link me-2" to="/docs">Docs</NavLink>
                <NavLink className="nav-link me-auto" to="/settings">Settings</NavLink>

                {auth.isAdmin &&
                    <NavLink className="nav-link me-2" to="/admin">Admin</NavLink>
                }
                
                {isNotLoginPage && 
                <div className="nav-item">
                    <AuthStatus></AuthStatus>
                </div>
                }
            </div>
        </nav>
    );
}