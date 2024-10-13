import React from 'react';
import ReactDOM from 'react-dom/client';
import {
  createBrowserRouter,
  createRoutesFromElements,
  Route,
  RouterProvider,
} from "react-router-dom";
import reportWebVitals from './reportWebVitals';
import 'bootstrap/dist/css/bootstrap.min.css';
import './index.css';

import ErrorPage from './error-page';
import { Layout, loader as rootLoader, action as rootAction } from './pages/Layout';
import { HomePage } from './pages/Home';
import { SettingsPage} from './pages/Settings';
import { AuthProvider, RequireAuth } from './contexts/auth';
import { LoginPage } from './pages/Login';
import { DocsPage } from './pages/Docs';
import { LoginCallbackPage } from './pages/LoginCallback';
import { AdminPage } from './pages/Admin'

const router = createBrowserRouter(
  createRoutesFromElements(
    <>
      <Route
        path="/"
        element={<Layout />}
        errorElement={<ErrorPage />}
        loader={rootLoader}
        action={rootAction}
      >
        <Route
          index
          element={<HomePage />}
        />
        <Route
          path="/docs"
          element={<DocsPage />}
        />
        <Route
          path="/settings"
          element={
            <RequireAuth>
              <SettingsPage />
            </RequireAuth>
          }
        />
        <Route
          path="/admin"
          element={
            <RequireAuth checkAdmin>
              <AdminPage />
            </RequireAuth>
          }
          errorElement={<ErrorPage />}
        />
        <Route
          path="/login"
          element={<LoginPage />}
          errorElement={<ErrorPage />}
        />
        <Route
          path="/login/callback"
          element={<LoginCallbackPage />}
          errorElement={<ErrorPage />}
        />
      </Route>
    </>
  )
)

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <React.StrictMode>
    <AuthProvider>
      <RouterProvider router={router} />
    </AuthProvider>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
