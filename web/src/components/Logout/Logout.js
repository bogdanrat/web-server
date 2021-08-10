import Button from "react-bootstrap/Button";
import React from "react";
import axios from "axios";

class Logout extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            token: props.token,
            setToken: props.setToken,
            filesToUpload: [],
        };
    }

    handleLogout = () => {
        axios({
            url: `${process.env.REACT_APP_API_URL}/logout`,
            method: 'post',
            headers: {
                "Authorization": `Bearer ${this.props.token?.access_token}`
            }
        }).then(async res => {
            if (res.status === 200) {
                sessionStorage.removeItem('token');
                this.setState({setToken: undefined});
                window.location.reload();
            }
        }).then(err => console.log(err));
    }

    render() {
        return (
            <div>
                <Button variant={"dark"} onClick={this.handleLogout} className={"mb-3"}>Logout</Button>
            </div>
        );
    }
}

export default Logout;
