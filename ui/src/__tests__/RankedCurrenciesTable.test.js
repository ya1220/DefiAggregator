import React from 'react';
import {act} from 'react-dom/test-utils'
import {render, unmountComponentAtNode} from 'react-dom';
import RankedPoolsTableC from '../RankedPoolsTableC.js';


beforeEach( ()=> { //before each test create a div element
    const elem = document.createElement('RankedPoolsTableC');
    elem.setAttribute('id', 'RankedPoolsTableC');
    document.body.appendChild(elem);
});

afterEach( ()=> { //remove div element so next test has clean <body>
    const elem = document.getElementById('RankedPoolsTableC');
    unmountComponentAtNode(elem);
    elem.remove();
})

test( 'RankedPoolsTableC test, fetches data from backend for RankedPoolsTableF', () => {
    const elem = document.getElementById('RankedPoolsTableC');
    act( () => {
        render(<RankedPoolsTableC/>, elem);
    });
});