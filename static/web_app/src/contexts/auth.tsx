import React from "react"
import { Button, Container } from "react-bootstrap"
import { Navigate, redirect, useLocation, useNavigate } from "react-router-dom"
import { LoaderBlock } from "../components/LoaderBlock"
import { TwitchLoginButton } from "../components/TwitchLoginButton"


interface AuthDataContext {
    isAuthenticated: boolean
    loading: boolean
    userName: string
    isAdmin: boolean
    Login: () => void
    Logout: (callback: VoidFunction) => void
}
export const AuthContext = React.createContext<AuthDataContext>({
    isAuthenticated: false,
    loading: true,
    userName: "",
    isAdmin: false,
    Login: () => {},
    Logout: (callback: VoidFunction) => {},
})

export const AuthProvider = (props: {
    children: React.ReactNode
}) => {
    const [isAuthenticated, setIsAuthenticated] = React.useState(false)
    const [loading, setLoading] = React.useState(true)
    const [userName, setUserName] = React.useState("")
    const [isAdmin, setIsAdmin] = React.useState(false)
    
    const checkAuthenticated = async () => {
        try {
            const url = "/auth/check/"
            const options = {method: "GET"}
            const response = await fetch(url, options);
            if (response.status != 200) {
                setIsAuthenticated(false);
                setLoading(false);
                return;
            }
            
            const jsonBody = await response.json();
            setIsAuthenticated(jsonBody.authenticated);
            setUserName(jsonBody.username);
            setIsAdmin(jsonBody.is_admin);
            setLoading(false);
        } catch (err) {
            console.log(err);
        }
    }
    React.useEffect(() => {checkAuthenticated();}, [])

    let Login = () => {
        window.location.assign("/auth/login/twitch/");
    };

    let Logout = (callback: VoidFunction) => {
        fetch("/auth/logout/", {method: "GET"})
        .catch((err) => console.log(err))
        .then(() => {
            setIsAuthenticated(false);
            callback();
        })
    };

    let value = {isAuthenticated, loading, userName, isAdmin, Login, Logout}
    return (
        <AuthContext.Provider value={value}>
            {props.children}
        </AuthContext.Provider>
    )
}

export function AuthStatus() {
    const auth = React.useContext(AuthContext)
    const navigate = useNavigate()
    const location = useLocation();
    const from = location.state?.from?.pathname || location.pathname;

    if (auth.isAuthenticated) {
        return (
            <div>
                <Button onClick={() => {
                    auth.Logout(() => navigate("/"))
                }}>
                    Logout
                </Button>
            </div>
        )
    } else {
        return (
            <TwitchLoginButton redirect_to={from} />
        )       
    }
}

export function RequireAuth({ children, checkAdmin }: { children: JSX.Element, checkAdmin ?: boolean }) {
    const auth = React.useContext(AuthContext)
    const location = useLocation();

    if (auth.loading) {
        return <div>Loading...</div>
    } else if (auth.isAuthenticated) {
        if (checkAdmin && !auth.isAdmin) {
            <Navigate to="/login" state={{from: location }} replace />
        } else {
            return children
        }
    }
    return <Navigate to="/login" state={{from: location }} replace />
}

export function useAuth() {
    return React.useContext(AuthContext)
}