import React from 'react';
import {act} from 'react-dom/test-utils'
import {render, unmountComponentAtNode} from 'react-dom';
import RawPortfolioF from '../RawPortfolioF.js';

beforeEach( ()=> { //before each test create a div element
    const elem = document.createElement('RawPortfolioF');
    elem.setAttribute('id', 'RawPortfolioF');
    document.body.appendChild(elem);
});

afterEach( ()=> { //remove div element so next test has clean <body>
    const elem = document.getElementById('RawPortfolioF');
    unmountComponentAtNode(elem);
    elem.remove();
})

test( 'RawPortfolioF test, renders table 1 (User specified Table)', () => {
    const elem = document.getElementById('RawPortfolioF');

    console.log('HTML ->', document.body.innerHTML);

    act( () => {
        render(<RawPortfolioF/>, elem);
    });

});