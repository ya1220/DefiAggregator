import React from 'react';
import {act} from 'react-dom/test-utils'
import {render, unmountComponentAtNode} from 'react-dom';
import RawPortfolioC from "../RawPortfolioC";


beforeEach( ()=> { //before each test create a div element
    const elem = document.createElement('RawPortfolioC');
    elem.setAttribute('id', 'RawPortfolioC');
    document.body.appendChild(elem);
});

afterEach( ()=> { //remove div element so next test has clean <body>
    const elem = document.getElementById('RawPortfolioC');
    unmountComponentAtNode(elem);
    elem.remove();
})

test( 'RawPortfolioC test, links front-end to back-end', () => {
    const elem = document.getElementById('RawPortfolioC');
    act( () => {
        render(<RawPortfolioC/>, elem);
    });
});