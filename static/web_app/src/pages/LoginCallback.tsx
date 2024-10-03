import { useCookies } from "react-cookie";
import { Navigate, useLocation } from "react-router-dom";
import { TwitchLoginButton } from "../components/TwitchLoginButton";

export function LoginCallbackPage() {
    const location = useLocation();
    const [cookies, setCookie, removeCookie] = useCookies(['redirect_to']);
    const redirect_to = cookies.redirect_to || "/";    
    const query = new URLSearchParams(location.search);
    const status = query.get('status') || "unknown";
    
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
        removeCookie('redirect_to');
        return (
            <Navigate to={redirect_to} replace/>
        )
    }
}