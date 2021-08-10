import React from 'react';
import axios from "axios";
import Button from 'react-bootstrap/Button';
import {Form} from "react-bootstrap";

class Dashboard extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            isFetching: false,
            files: [],
            token: props.token,
            setToken: props.setToken,
        };
    }

    render() {
        return (
            <div>
                <h2>Dashboard</h2>
                {this.state.isFetching ? 'Fetching files...' :
                    <div>
                        <Form method="post" action="http://localhost:8080/api/files" id="files-form">
                            <Form.Group controlId="formFileMultiple" className="mb-3">
                                <Form.Control type="file" multiple/>
                            </Form.Group>
                            <Button variant="success" type="submit" onClick={this.submitForm}>
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
                                    <Button variant="warning" className="btn-sm"
                                            onClick={this.deleteFile(file.key)}>Delete</Button>
                                </div>
                            )}
                        </ul>
                        <div style={{maxWidth: '250px', display: "flex", justifyContent: "space-between"}}>
                            <Button variant="primary" className="btn-sm" onClick={this.downloadCSV}>Download
                                CSV</Button>
                            <Button variant="info" className="btn-sm" onClick={this.downloadExcel}>Download
                                Excel</Button>
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
                "Authorization": `Bearer ${this.props.token?.access_token}`
            },
        }).then(response => {
            this.setState({files: response.data, isFetching: false});
        }).catch(e => {
            if (e.response.status === 401) {
                this.refreshToken();
            }
            this.setState({...this.state, isFetching: false});
        });
    }

    downloadCSV = () => {
        axios.get('http://localhost:8080/api/files/csv', {
            headers: {
                "Authorization": `Bearer ${this.props.token?.access_token}`
            },
        }).then(response => {
            const headers = response.headers;
            const body = response.data;
            const contentDisposition = headers['content-disposition'];
            let fileName = "unnamed.csv"

            if (contentDisposition) {
                const fileNameRegex = new RegExp(`(filename=)(.*)`, 'g');
                const match = contentDisposition.match(fileNameRegex);
                if (match && match.length) {
                    fileName = match[0].split("=")[1];
                }
            }

            const a = document.createElement('a');
            let binaryData = [];
            binaryData.push(body);
            const objectURL = window.URL.createObjectURL(new Blob(binaryData, {type: "text/csv"}))

            try {
                a.style.display = 'none';
                a.href = objectURL;
                a.download = fileName;
                document.body.appendChild(a);
                a.click();
            } finally {
                window.URL.revokeObjectURL(objectURL);
                document.body.removeChild(a);
            }

        }).catch(e => {
            console.log(e);
        });
    }

    downloadExcel = () => {
    }

    deleteFile = (fileKey) => {
        return () => {
            axios.delete(`http://localhost:8080/api/file`, {
                headers: {
                    "Authorization": `Bearer ${this.props.token?.access_token}`
                },
                data: {
                    "key": fileKey,
                },
            }).then(res => {
                if (res.status === 200) {
                    window.location.reload();
                }
            }).then(err => console.log(err));
        }
    }

    refreshToken = () => {
        return axios.post('http://localhost:8080/token/refresh', {
            'refresh_token': this.props.token?.refresh_token,
        }, {
            headers: {
                "Authorization": `Bearer ${this.props.token?.access_token}`
            },
        }).then(res => {
            const token = res.data;
            this.props.setToken(token);
            window.location.reload();
        })
    }
}


export default Dashboard;
