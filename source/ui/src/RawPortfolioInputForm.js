import React from 'react';
import { Form, Segment, Button, Popup } from 'semantic-ui-react'

// does not do any validation
export default class RawPortfolioInputForm extends React.Component {
    state = {
        token: '',
        amount: '',
    };
    onChangeName = this._onChangeName.bind(this);
    onChangeTime = this._onChangeTime.bind(this);
    onSubmit = this._onSubmit.bind(this);
    render() {
        return (
            <div className="ui container">
                <Segment> 
                    <h3>Enter Your Portfolio <Popup content="Enter your portfolio - our engine will suggest where to deploy it. Enter negative numbers to remove amounts" position="top center" trigger={<i class="info circle icon portfolio-popup"></i>} /></h3>
                    <Form onSubmit={this.onSubmit}>
                        <Form.Field>
                            <label><div className="inputLabel">Token</div></label>
                            <select value={this.state.token} onChange={this.onChangeName}>
                                <option value="" selected disabled hidden>Select Token</option>
                                <option value="DAI">DAI</option>
                                <option value="USDC">USDC</option>
                                <option value="USDT">USDT</option>
                                <option value="WETH">ETH</option>
                                <option value="WBTC">BTC</option>
                                <option value="DOGE">DOGE</option>
                            </select>
                        </Form.Field>
                        <Form.Field>
                            <label><div className="inputLabel">Amount</div></label>
                            <input type="number" step="any" placeholder='Enter Amount' value={this.state.amount} onChange={this.onChangeTime} />
                        </Form.Field>
                        <Form.Field className="portfolioSubmit">
                            <Button type='submit'>Submit</Button>
                        </Form.Field>
                    </Form>
                </Segment>
            </div>
        );
    }
    _onChangeName(e) {
        this.setState({
            token: e.target.value
        });
    }
    _onChangeTime(e) {
        this.setState({
            amount: e.target.value
        });
    }

    _onSubmit() {
        const payload = {
            token: this.state.token,
            amount: parseFloat(this.state.amount),
        };
        fetch('http://localhost:8080/raw_portfolio_input', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        });
        this.setState({
            token: '',
            amount: '',
        });
    }
}

/*
                        <Form.Field>
                            <label>Pool Size</label>
                            <input placeholder='Pool Size' value={this.state.pool_sz} onChange={this.onChangePool_sz} />
                        </Form.Field>
*/