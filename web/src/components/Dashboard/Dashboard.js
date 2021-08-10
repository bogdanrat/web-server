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
            filesToUpload: [],
        };
    }

    render() {
        return (
            <div>
                <h2>Dashboard</h2>
                {this.state.isFetching ? 'Fetching files...' :
                    <div>
                        <Form method="post" action="/api/files" id="files-form">
                            <Form.Group controlId="formFileMultiple" className="mb-3">
                                <Form.Control type="file"
                                              multiple
                                              size={"sm"}
                                              style={{width: '30%'}}
                                              onChange={e => this.handleFilesChanged(e.target.files)}/>
                            </Form.Group>
                            <Button variant="success" type="submit" onClick={e => this.handleFormSubmit(e)} size={"sm"}>
                                Upload
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
                        <Button variant="danger" className="btn-sm mt-2" onClick={this.deleteAllFiles}>Delete
                            All</Button>

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

        axios.get('api/files', {
            headers: {
                "Authorization": `Bearer ${this.props.token?.access_token}`
            },
        }).then(response => {
            this.setState({files: response.data, isFetching: false});
        }).catch(e => {
            if (e.response.status === 401) {
                this.refreshToken();
                window.location.reload();
            }
            this.setState({...this.state, isFetching: false});
        });
    }

    downloadCSV = () => {
        axios.get('/api/files/csv', {
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
        axios({
            url: '/api/files/excel',
            method: 'get',
            responseType: 'blob',
            headers: {
                "Authorization": `Bearer ${this.props.token?.access_token}`
            }
        }).then((response) => {
            const contentDisposition = response.headers['content-disposition'];
            let fileName = "unnamed.xlsx"

            if (contentDisposition) {
                const fileNameRegex = new RegExp(`(filename=)(.*)`, 'g');
                const match = contentDisposition.match(fileNameRegex);
                if (match && match.length) {
                    fileName = match[0].split("=")[1];
                }
            }

            const url = window.URL.createObjectURL(new Blob([response.data]));
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', fileName);
            document.body.appendChild(link);
            link.click();
        });
    }

    deleteFile = (fileKey) => {
        return () => {
            axios.delete(`/api/file`, {
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

    deleteAllFiles = () => {
        axios({
            url: '/api/files',
            method: 'delete',
            headers: {
                "Authorization": `Bearer ${this.props.token?.access_token}`,
            }
        }).then(res => {
            if (res.status === 200) {
                window.location.reload();
            } else if (res.status === 401) {
                this.refreshToken();
                window.location.reload();
            }
        }).then(err => console.log(err));
    }

    handleFilesChanged = (files) => {
        this.setState({filesToUpload: Array.from(files)})
    }

    handleFormSubmit = (event) => {
        event.preventDefault();

        let formData = new FormData();
        this.state.filesToUpload.forEach(file => {
            formData.append("files", file);
            console.log("file", file);
        })

        axios({
            url: '/api/files',
            method: 'post',
            data: formData,
            headers: {
                "Authorization": `Bearer ${this.props.token?.access_token}`,
                "Content-Type": "multipart/form-data"
            }
        }).then(res => {
            if (res.status === 201) {
                window.location.reload();
            } else if (res.status === 401) {
                this.refreshToken();
                window.location.reload();
            }
        }).then(err => console.log(err));
    }

    refreshToken = () => {
        axios.post('/api/token/refresh', {
            'refresh_token': this.props.token?.refresh_token,
        }, {
            headers: {
                "Authorization": `Bearer ${this.props.token?.access_token}`
            },
        }).then(res => {
            const token = res.data;
            this.props.setToken(token);
        }).catch(err => {
            console.log("error refreshing token: ", err);
        });
    }
}

export default Dashboard;
