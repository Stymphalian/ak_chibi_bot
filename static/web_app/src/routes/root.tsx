import { Outlet, Link, useLoaderData, Form, redirect, NavLink, useNavigation, useSubmit } from "react-router-dom";
import { getContacts, createContact } from "../contacts";
import {useEffect} from 'react';

export async function action() {
    const contact = await createContact();
    return redirect(`/contacts/${contact.id}/edit`);
    // return { contact };
}

export async function loader(data:any) {
    const request = data.request;
    const url = new URL(request.url);
    const q= url.searchParams.get("q");
    const contacts = await getContacts(q);
    return { contacts, q };
}

export default function Root() {
    const data: any = useLoaderData();
    const contacts = data.contacts;
    const q = data.q;
    const navigation = useNavigation();
    const submit = useSubmit();

    const searching = navigation.location &&
        new URLSearchParams(navigation.location!.search).has("q");

    useEffect(() => {
        const elem = document.getElementById("q") as HTMLInputElement;
        if (elem !== null) {
            elem.value = q;
        }
    }, [q]);

    return (
        <>
            <div id="sidebar">
                <h1>React Router Contacts</h1>
                <div>
                    <Link to="dashboard">
                        Dashboard
                    </Link>
                </div>
                <div>
                    <Form id="search-form" role="search">
                        <input
                            id="q"
                            aria-label="Search contacts"
                            placeholder="Search"
                            type="search"
                            name="q"
                            defaultValue={q}
                            className={searching ? "loading": ""}
                            onChange={(event) => {
                                const isFirstSearch = (q == null);
                                submit(event.currentTarget.form, {
                                    replace: !isFirstSearch,
                                });
                            }}
                        />
                        <div
                            id="search-spinner"
                            aria-hidden
                            hidden={!searching}
                        />
                        <div
                            className="sr-only"
                            aria-live="polite"
                        ></div>
                    </Form>
                    <Form method="post">
                        <button type="submit">New</button>
                    </Form>
                </div>
                <nav>
                    {contacts.length ? (
                        <ul>
                            {contacts.map((contact: any) => (
                                <li key={contact.id}>
                                    <NavLink
                                        to={`contacts/${contact.id}`}
                                        className={({ isActive, isPending }) =>
                                            isActive
                                              ? "active"
                                              : isPending
                                              ? "pending"
                                              : ""
                                          }
                                    >
                                        <Link to={`contacts/${contact.id}`}>
                                            {contact.first || contact.last ? (
                                                <>
                                                    {contact.first} {contact.last}
                                                </>
                                            ) : (
                                                <i>No Name</i>
                                            )}{" "}
                                            {contact.favorite && <span>★</span>}
                                        </Link>
                                    </NavLink>

                                </li>
                            ))}
                        </ul>
                    ) : (
                        <p>
                            <i>No contacts</i>
                        </p>
                    )}
                </nav>
            </div>
            <div id="detail"
                className={
                    navigation.state === "loading" ? "loading" : ""
                }
            >
                <Outlet />
            </div>
        </>
    );
}





// import App from '../App';
// import { Outlet } from 'react-router-dom';

// export default function Root() {
//     return (
//         <>
//             <App />
//             <Outlet />
//         </>
//     );
// }