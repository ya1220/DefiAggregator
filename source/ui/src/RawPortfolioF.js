import React from 'react';
import { Form, Table, Segment, Label } from 'semantic-ui-react'

function numberWithCommas(x) {
    return x.toString().replace(/\B(?<!\.\d*)(?=(\d{3})+(?!\d))/g, ",");
}

//export default function RawPortfolioF() {
export default function RawPortfolioF({results}) {
    const rows = results.map(((result, index) => {
        let color='grey';
        return (
            <Table.Row key={ index }>
                <td><Label class="ui horizontal label" color={color}>{ index + 1 }</Label></td>
                <td>{ result.token }</td>
                <td class="right aligned">{ numberWithCommas(result.amount) }</td>
            </Table.Row>
        );
    }));
    return (
        <div className="ui container">
            <Segment class="ui inverted segment">
                <div className="recommended-float">
                    <div className="recommended-header">Current Portfolio</div>
                    <div className="recommended-reset">
                        <Form className="tableButton">
                            <div class="ui button">Reset</div>
                        </Form>
                    </div>
                </div>
                <div class="ui basic table">
                    <Table.Header>
                        <Table.Row>
                            <Table.HeaderCell><h3 className="headerTitle">Ranking</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Token/Pair</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Amount</h3></Table.HeaderCell>

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


// needed for testing, DONT delete
/*
var results = [
    {
        tokenorpair: "ETH",
        amount: 69,
        percentageofportfolio: 420,
        risk_setting: 420,
    },
    {
        tokenorpair: "BTC",
        amount: 69,
        percentageofportfolio: 420,
        risk_setting: 420,
    },
    {
        tokenorpair: "BTC",
        amount: 420,
        percentageofportfolio: 420,
        risk_setting: 1,
    },
    {
        tokenorpair: "wETH",
        amount: 123,
        percentageofportfolio: 10,
        risk_setting: 100,
    }
]
*/