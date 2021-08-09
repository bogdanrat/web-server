import React from 'react';
import './App.css'
import 'bootstrap/dist/css/bootstrap.min.css';
import {BrowserRouter, Route, Switch} from 'react-router-dom';
import Dashboard from '../Dashboard/Dashboard';
import Login from '../Login/Login';
import useToken from './useToken';

function App() {
    const { token, setToken } = useToken();
    if (!token) {
        return <Login setToken={setToken}/>
    }

    return (
        <div className="wrapper">
            <BrowserRouter>
                <Switch>
                    <Route path="/dashboard">
                        <Dashboard token={token}/>
                    </Route>
                </Switch>
            </BrowserRouter>
        </div>
    );
}

export default App;
