import React from 'react';
import axios from "axios";
import Button from 'react-bootstrap/Button';
import {Form} from "react-bootstrap";

class Dashboard extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            isFetching: false,
            files: []
        };
    }

    render() {
        return (
            <div>
                <h2>Dashboard</h2>
                {this.state.isFetching ? 'Fetching files...' :
                    <div>
                        <Form method="post" action="http://localhost:8080/api/files">
                            <Form.Group controlId="formFileMultiple" className="mb-3">
                                <Form.Control type="file" multiple />
                            </Form.Group>
                            <Button variant="success" type="submit">
                                Submit
                            </Button>
                        </Form>
                        <ul>
                            {this.state.files.map(file =>
                                <div style={{
                                    maxWidth: '400px',
                                    display: "flex",
                                    justifyContent: "space-between",
                                    marginTop: '20px'
                                }}>
                                    <li key={file.key}>{file.key}</li>
                                    <Button variant="warning" className="btn-sm">Delete</Button>
                                </div>
                            )}
                        </ul>
                        <div style={{maxWidth: '250px', display: "flex", justifyContent: "space-between"}}>
                            <Button variant="primary" className="btn-sm">Download CSV</Button>
                            <Button variant="info" className="btn-sm">Download Excel</Button>
                        </div>
                        <Button variant="danger" className="btn-sm mt-2">Delete All</Button>

                    </div>
                }
            </div>
        )
    }

    componentDidMount() {
        this
            .fetchFilesAsync()
    }

    fetchFilesAsync() {
        this.setState({...this.state, isFetching: true});

        axios.get('http://localhost:8080/api/files', {
            headers: {
                "Authorization": `Bearer ${this.props.token}`
            },
        }).then(response => {
            this.setState({files: response.data, isFetching: false});
        }).catch(e => {
            console.log(e);
            this.setState({...this.state, isFetching: false});
        });
    }
}

export default Dashboard;
