import React from 'react';
import RawPortfolioF from './RawPortfolioF.js';
import Pusher from 'pusher-js';
//import 'semantic-ui-css/semantic.js';
const socket = new Pusher('7885860875bb513c3e34', {
    cluster: 'eu',
    encrypted: true
});

export default class RawPortfolioC extends React.Component {
    state = {
        raw_portfolio: []
    };
    componentDidMount() {
        const channel = socket.subscribe('raw_portfolio');
        channel.bind('raw_portfolio', (raw_portfolio_data) => {
            this.setState(raw_portfolio_data);
        });


        // change this url:

        fetch('http://localhost:8080/raw_portfolio')
            .then((response) => response.json())
            .then((response) => this.setState(response));
    }
    render() {
        return <RawPortfolioF results={this.state.raw_portfolio} />;
    }
}
		