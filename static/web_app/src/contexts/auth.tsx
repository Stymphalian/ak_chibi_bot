import React from "react"
import { Button, Container } from "react-bootstrap"
import { Navigate, redirect, useLocation, useNavigate } from "react-router-dom"
import { LoaderBlock } from "../components/LoaderBlock"
import { TwitchLoginButton } from "../components/TwitchLoginButton"
import axios from "axios"

const isTokenExpired = (token: string) => {
    if (token == "") return true;
    const dataPart = token.split(".")[1]
    if (dataPart == "") return true;
    const data = JSON.parse(atob(dataPart))
    return data.exp < (Date.now() / 1000);
}

interface AuthDataContext {
    isAuthenticated: boolean
    loading: boolean
    userName: string
    isAdmin: boolean
    getAccessToken: () => Promise<string>,
    Login: () => void
    Logout: (callback: VoidFunction) => void
}
export const AuthContext = React.createContext<AuthDataContext>({
    isAuthenticated: false,
    loading: true,
    userName: "",
    isAdmin: false,
    getAccessToken: () => Promise.resolve(""),
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
    const [accessToken, setAccessToken] = React.useState("")
    
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
            setUserName(jsonBody.user_name);
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
        fetch("/auth/logout/", {method: "POST"})
        .catch((err) => console.log(err))
        .then(() => {
            setIsAuthenticated(false);
            callback();
        })
    };

    async function getAccessToken(): Promise<string> {
        if (!isTokenExpired(accessToken)) {
            return Promise.resolve(accessToken);
        } else {
            try {
                let response = await axios.get("/auth/token/");
                setAccessToken(response.data.token);
                return response.data.token;
            } catch (err) {
                console.log(err);
                return "";
            }
        }
    }

    let value = {isAuthenticated, loading, userName, isAdmin, getAccessToken, Login, Logout}
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
    const auth = useAuth();
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