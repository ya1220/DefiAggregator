import React from 'react';
import {act} from 'react-dom/test-utils'
import {render, unmountComponentAtNode} from 'react-dom';
import OptimisedPortfolioC from '../OptimisedPortfolioC';


beforeEach( ()=> { //before each test create a div element
    const elem = document.createElement('OptimisedPortfolioC');
    elem.setAttribute('id', 'OptimisedPortfolioC');
    document.body.appendChild(elem);
});

afterEach( ()=> { //remove div element so next test has clean <body>
    const elem = document.getElementById('OptimisedPortfolioC');
    unmountComponentAtNode(elem);
    elem.remove();
})

test( 'OptimisedPortfolioC test, links front-end to back-end', () => {
    const elem = document.getElementById('OptimisedPortfolioC');
    act( () => {
        render(<OptimisedPortfolioC/>, elem);
    });
});