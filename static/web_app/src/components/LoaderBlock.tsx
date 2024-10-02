
export interface LoaderBlockProps {
    children: React.ReactNode;
    loading: boolean;
}

export function LoaderBlock(props:LoaderBlockProps) {
    if (props.loading) {
        return (
            <div className="loader">
                <div className="spinner">   
                    Loading...
                </div>
            </div>
        );
    } else {
        return <>{props.children}</>;
    }
}