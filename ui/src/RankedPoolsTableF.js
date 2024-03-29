import React from 'react';
import { Form, Table, Segment, Label, Popup } from 'semantic-ui-react'
//import './utilities';

function numberWithCommas(x) {
    return x.toString().replace(/\B(?<!\.\d*)(?=(\d{3})+(?!\d))/g, ",");
}

export default function RankedPoolsTableF({results}) {
    const rows = results.map(((result, index) => {
        let color;
        if (index === 0 || index === 1 || index === 2) {
            color='blue';
        } else {
            color='grey';
        }
        return (
            <Table.Row key={ index }>
                <td><Label class="ui horizontal label" color={color}>{ index + 1 }</Label></td>
                <td>{ result.backend_pair }</td>
                <td>{ result.pool_source }</td>
                <td class="right aligned">{ "$" + numberWithCommas(result.backend_poolsize.toFixed(0)) }</td>
                <td class="right aligned">{ "$" + numberWithCommas(result.backend_volume.toFixed(0)) }</td>
                <td class="right aligned">{ (result.pool_ratio_0 * 100).toFixed(1) + "%" }</td>
                <td class="right aligned">{ (result.backend_yield * 100).toFixed(2) + "%" }</td>
                <td class="right aligned">{ (result.volatility * 100).toFixed(1) + "%" }</td>
                <td class="right aligned">{ (result.ROIestimate * 100).toFixed(2) + "%" }</td>
                <td class="right aligned">{ (result.ROIvoladjest).toFixed(2) }</td>
                <td class="right aligned">{ (result.ROIhist * 100).toFixed(2) + "%" }</td>
            </Table.Row>
        );
    }));
    return (
        <div className="ui container">
            <Segment>
                <div className="recommended-float">
                    <div className="recommended-header">
                        <div>Recommended Liquidity Pools</div>
                    </div>
                    <div className="recommended-reset">
                        <Form className="tableButton">
                            <div class="ui animated fade button">
                                <div class="visible content">Refresh</div>
                                <div class="hidden content">
                                    <i class="refresh icon"></i>
                                </div>
                            </div>
                        </Form>
                    </div>
                </div>
                <div class="ui basic table">
                    <Table.Header>
                        <Table.Row>
                            <Table.HeaderCell><h3 className="headerTitle">Ranking</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Tokens</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Pool</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Pool Value Locked</h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Pool Trading Volume. <Popup header="Pool Volume" content="Average daily trading volume over the last 30 days." position="top center" trigger={<i class="info circle icon"></i>} /></h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Ratio Token0 <Popup header="% Yield" content="Interest rate of pool - ONLY applicable to LENDING pools." position="top center" trigger={<i class="info circle icon"></i>} /></h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">% Interest <Popup header="% Yield" content="Interest rate of pool - ONLY applicable to LENDING pools." position="top center" trigger={<i class="info circle icon"></i>} /></h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">% Historical Volatility <Popup header="Historical Volatility" content="Standard deviation of returns of assets in a pool over the last 30 days." position="top center" trigger={<i class="info circle icon"></i>} /></h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">% ROI Est. <Popup header="Raw Return On Investment Estimate" content="Interest rate (if applicable) + impermanent loss (to arbitrageurs) + returns from share of pool commissions + native token rewards. In USD terms. Prices of tokens assumed to stay constant compared to current prices." position="top center" trigger={<i class="info circle icon"></i>} /></h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">Sharpe Ratio <Popup header="Sharpe Ratio" content="Historical returns divided by volatility of returns (a standardised metric for evaluating risk-adjusted pool returns). Higher Sharpe = better risk-adjusted returns. Historical returns estimated over the last 30 days. In USD terms. Includes components used in calculation of ROI estimate + historical changes in token values in USD terms." position="top center" trigger={<i class="info circle icon"></i>} /></h3></Table.HeaderCell>
                            <Table.HeaderCell><h3 className="headerTitle">% ROI Hist <Popup header="Return On Investment" content="Return On Investment over the last 30 days." position="top center" trigger={<i class="info circle icon"></i>} /></h3></Table.HeaderCell>
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

/*
    <Table.Cell>{ result.max_amount }</Table.Cell>
    <Table.Cell>{ result.raw_yield }</Table.Cell>
    <Table.Cell>{ result.normalized_yield }</Table.Cell>
*/

// currencoutputtable needed for testing, DONT delete
/*
var ranked_pools_table = [
    {
        backend_pair: "ETH/DAI",
        backend_poolsize: 12039.129309,
        backend_volume: 21897.48722342,
        backend_yield: 10.98714235,
        pool_source: "Uniswap",
        volatility: 0.1976342134,
        ROIestimate: 0.1821039010,
        ROIvoladjest: 0.192187540,
        ROIhist: 0.060972182653,
    },
    {
        backend_pair: "WBTC/ETH",
        backend_poolsize: 123.198984124,
        backend_volume: 983.129876453,
        backend_yield: 5.0896212123,
        pool_source: "Uniswap",
        volatility: 0.88225764123,
        ROIestimate: 0.028765123,
        ROIvoladjest: 0.129864953,
        ROIhist: 0.0257678648391,
    },
    {
        backend_pair: "WBTC/DAI",
        backend_poolsize: 1998712.3253223,
        backend_volume: 973923,
        backend_yield: 460,
        pool_source: "Balancer",
        volatility: 123,
        ROIestimate: 10,
        ROIvoladjest: 100,
        ROIhist: 100,
    },
    {
        backend_pair: "ETH/DAI",
        backend_poolsize: 420,
        backend_volume: 420,
        backend_yield: 460,
        pool_source: "Uniswap",
        volatility: 123,
        ROIestimate: 10,
        ROIvoladjest: 100,
        ROIhist: 100,
    }
]*/

// <td class="right aligned">{ result.backend_yield === 0 ? result.backend_yield : (result.backend_yield * 100).toFixed(2) +  + "%" }</td>