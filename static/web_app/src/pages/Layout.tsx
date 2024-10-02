import { Outlet } from "react-router-dom";
import { TopNavBar } from "../components/TopNavBar";
import Footer from "../components/Footer";

export async function loader() {
    return null
}
export async function action() {
    return null
}
export function Layout() {
    return (
        <div className="container">
            <TopNavBar />
            <Outlet />
            <Footer />
        </div>
    )
}