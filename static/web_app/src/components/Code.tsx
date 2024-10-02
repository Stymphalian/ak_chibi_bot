import './Code.css';

interface CodeBlockProps {
    children: string;
}

export function Code({ children }: CodeBlockProps) {
    return (
        <div className="code-block">{children}</div>
    );
}
