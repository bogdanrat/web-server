const axios = require("axios");

class Uploader {
    upload(file, onUploadProgress, access_token) {
        let formData = new FormData();
        formData.append("file", file);

        return axios.post('http://localhost:8080/files', formData, {
            headers: {
                "Content-Type": "multipart/form-data",
                "Authorization": `Bearer ${access_token}`
            },
            onUploadProgress,
        });
    }
}