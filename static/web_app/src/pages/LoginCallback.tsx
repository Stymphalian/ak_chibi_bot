
import { Navigate, useLocation } from "react-router-dom";
import { TwitchLoginButton } from "../components/TwitchLoginButton";
import Cookies from "js-cookie";
export function LoginCallbackPage() {
    const location = useLocation();
    const query = new URLSearchParams(location.search);
    const status = query.get('status') || "unknown";
    const redirect_to = Cookies.get('redirect_to') || "/";
    
    if (status !== "success") {
        return (
            <div className="container-fluid pt-5 pb-2 my-2 bg-light border rounded-3">
                <h1 className="display-5 fw-semibold">Failed to login</h1>
                <p className="lead">Please try again.</p>
                <div>
                    <TwitchLoginButton redirect_to={redirect_to}/>
                </div>
            </div>
        )
    } else {
        Cookies.remove('redirect_to', {
            path: '/',
            secure: true,
            sameSite : 'Strict'
        });
        return (
            <Navigate to={redirect_to} replace/>
        )
    }
}