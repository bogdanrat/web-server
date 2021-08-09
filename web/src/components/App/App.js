import React from 'react';
import './App.css'
import 'bootstrap/dist/css/bootstrap.min.css';
import {BrowserRouter, Route, Switch} from 'react-router-dom';
import Dashboard from '../Dashboard/Dashboard';
import Login from '../Login/Login';
import useToken from './useToken';

function App() {
    const {token, setToken} = useToken();
    if (!token) {
        return <Login setToken={setToken}/>
    }

    return (
        <div className="wrapper">
            <BrowserRouter>
                <Switch>
                    <Route path="/dashboard">
                        <Dashboard token={token} setToken={setToken}/>
                    </Route>
                </Switch>
            </BrowserRouter>
        </div>
    );
}

// async function fetchConfig() {
//     return fetch('http://localhost:8080/config/auth', {
//         method: 'GET',
//     }).then(res => res.json())
//         .catch(err => console.log(err));
// }

export default App;
