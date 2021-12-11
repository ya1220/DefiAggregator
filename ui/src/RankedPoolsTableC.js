import React from 'react';
import RankedPoolsTableF from './RankedPoolsTableF';
import Pusher from 'pusher-js';

const socket = new Pusher('7885860875bb513c3e34', {
    cluster: 'eu',
    encrypted: true
});

export default class ConnectedRankedPoolsTableF extends React.Component {
    state = {
        ranked_pools_table: []
    };
    componentDidMount() {
        const channel = socket.subscribe('ranked_pools_table');
        channel.bind('ranked_pools_table', (ranked_pools_table_data) => {
            this.setState(ranked_pools_table_data);
        });

        // change this url:

        fetch('http://localhost:8080/ranked_pools_table')
            .then((response) => response.json())
            .then((response) => this.setState(response));
    }
    render() {
        return <RankedPoolsTableF results={this.state.ranked_pools_table} />;
    }
}