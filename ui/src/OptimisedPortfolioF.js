import React from 'react';
import { Table, Segment, Label } from 'semantic-ui-react'

function numberWithCommas(x) {
    return x.toString().replace(/\B(?<!\.\d*)(?=(\d{3})+(?!\d))/g, ",");
}

export default function OptimisedPortfolioF({results}) {
    const rows = results.map(((result, index) => {
        let color='grey';
        return (
            <Table.Row key={ index }>
                <td><Label class="ui horizontal label" color={color}>{ index + 1 }</Label></td>
                <td>{ result.pool }</td>
                <td class="right aligned">{ result.token0}</td>
                <td class="right aligned">{ result.token1}</td>

                <td class="right aligned">{ result.amount_token0.toFixed(2) }</td>  
                <td class="right aligned">{ result.amount_token1.toFixed(3) }</td>

                <td class="right aligned">{ "$" + numberWithCommas(result.total_value_usd.toFixed(0)) }</td>

                <td class="right aligned">{ (result.percentageofportfolio * 100).toFixed(1) + "%" }</td>
                <td class="right aligned">{ (result.roi_estimate * 100).toFixed(1) + "%" }</td>
                <td class="right aligned">{ (result.volatility * 100).toFixed(1) + "%" }</td>
                <td class="right aligned">{ result.risk_setting }</td>
            </Table.Row>
        );
    }));
    return (
        <div className="ui container">
            <Segment class="ui inverted segment">
                <div className="recommended-header">Recommended Portfolio</div>
                <div class="ui basic table">
                    <Table.Header>
                        <Table.Row>
                            <Table.HeaderCell><h3 className="headerTitle">Ranking</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Pool</h3></Table.HeaderCell>

                            <Table.HeaderCell><h3 className="headerTitle">Token0</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Token1</h3></Table.HeaderCell>

                            <Table.HeaderCell><h3 className="headerTitle">Amt Token0</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Amt Token1</h3></Table.HeaderCell>

                            <Table.HeaderCell><h3 className="headerTitle">Value US$</h3></Table.HeaderCell>

                            <Table.HeaderCell><h3 className="headerTitle">% Portfolio US$</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">ROI Est %</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Volatility %</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Risk Setting</h3></Table.HeaderCell>
                        </Table.Row>
                    </Table.Header>
                    <Table.Body>
                        { rows }
                    </Table.Body>
                </div>
            </Segment>
        </div>
    );
}



// results needed for testing, DONT delete
/*
var results = [
    {
        tokenorpair: "ETH",
        pool: 'Uniswap',
        amount: 69,
        percentageofportfolio: 420,
        roi_estimate: 420,
        risk_setting: 420
    },
    {
        tokenorpair: "BTC",
        pool: "Uniswap",
        amount: 69,
        percentageofportfolio: 420,
        roi_estimate: 420,
        risk_setting: 420
    },
    {
        tokenorpair: "BTC",
        pool: 'Uniswap',
        amount: 420,
        percentageofportfolio: 420,
        roi_estimate: 50,
        risk_setting: 1,
    },
    {
        tokenorpair: "wETH",
        pool: 'Uniswap',
        amount: 123,
        percentageofportfolio: 10,
        roi_estimate: 100,
        risk_setting: 100,
    }
]*/