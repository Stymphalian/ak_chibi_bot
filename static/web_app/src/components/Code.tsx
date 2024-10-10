import './Code.css';

interface CodeBlockProps {
    children: string;
}

export function Code({ children }: CodeBlockProps) {
    return (
        <span className="code-block user-select-all">{children}</span>
    );
}
