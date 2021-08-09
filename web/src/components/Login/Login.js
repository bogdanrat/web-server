import React from 'react';
import PropTypes from 'prop-types';
import './Login.css'

function Login({setToken}) {
    const handleSubmit = async e => {
        e.preventDefault();
        const body = new URLSearchParams(new FormData(e.target));
        const res = await loginUser(body);
        if (res.status_code !== 401) {
            setToken(res);
        }
    }
    //
    // let config;
    // useEffect(() => {
    //     async function fetchConfig() {
    //         config = await getConfig();
    //     }
    //     fetchConfig();
    //
    // }, []);
    // console.log("config:", config);


    return (
        <div className="login-wrapper">
            <h1>Please Log In</h1>
            <form onSubmit={handleSubmit}>
                <label htmlFor="email">Email:</label>
                <input type="text" name="email" id="email" required/>
                <br/>
                <label htmlFor="password">Password:</label>
                <input type="password" name="password" id="password" required/>
                <br/>
                <label htmlFor="qr_code">QR Code:</label>
                <input type="text" name="qr_code" id="qr_code" required/>

                <div>
                    <button type="submit">Submit</button>
                </div>
            </form>
        </div>
    )
}

async function loginUser(body) {
    return fetch('http://localhost:8080/login', {
        method: 'POST',
        body: body
    }).then(data => data.json()).catch(err => console.log("err:", err))
}

Login.propTypes = {
    setToken: PropTypes.func.isRequired
}

export default Login;
