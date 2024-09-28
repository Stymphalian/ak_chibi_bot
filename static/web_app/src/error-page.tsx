import { useRouteError } from "react-router-dom";

export default function ErrorPage() {
  const error: any = useRouteError();
  console.error(error);

  return (
    <div id="error-page" className="container">
      <h1 className="mt-5">Oops!</h1>
      <p className="lead">Sorry, an unexpected error has occurred.</p>
      <p className="text-muted">
        <i>{error.statusText || error.message}</i>
      </p>
    </div>
  );
}