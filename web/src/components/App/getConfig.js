import { useState } from 'react';

export default function getConfig() {
    const fetchConfig = () => {
        return fetch('/config/auth', {
            method: 'GET',
        })
            .then(response => response.json())
            .catch(err => console.log(err));
    };

    const [config] = useState(fetchConfig());

    return {
        config
    }
}
