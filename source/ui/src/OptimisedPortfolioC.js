import React from 'react';
import OptimisedPortfolioF from './OptimisedPortfolioF';
import Pusher from 'pusher-js';
//import 'semantic-ui-css/semantic.js';
const socket = new Pusher('7885860875bb513c3e34', {
    cluster: 'eu',
    encrypted: true
});

export default class OptimisedPortfolioC extends React.Component {
    state = {
        optimised_portfolio: []
    };
    componentDidMount() {
        const channel = socket.subscribe('optimised_portfolio');
        channel.bind('optimised_portfolio', (optimised_portfolio_data) => {
            this.setState(optimised_portfolio_data);
        });


        // change this url:

        fetch('http://localhost:8080/optimised_portfolio')
            .then((response) => response.json())
            .then((response) => this.setState(response));
    }
    render() {
        return <OptimisedPortfolioF results={this.state.optimised_portfolio} />;
    }
}