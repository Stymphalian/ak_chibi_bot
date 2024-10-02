
import github_mark from "../assets/github-mark/github-mark.svg"

export default function Footer() {

    return (
        <footer className="fixed bottom-0 right-0 p-4">
            <div className="container d-flex align-items-center justify-content-center">
                <div>
                    <a href="https://github.com/Stymphalian/ak_chibi_bot" 
                        target="_blank" rel="noopener noreferrer">
                        <img src={github_mark} alt="Github" className="h-6" width="40px"/>
                    </a>
                </div>                
            </div>
        </footer>
    )

}