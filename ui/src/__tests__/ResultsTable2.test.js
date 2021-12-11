import React from 'react';
import {act} from 'react-dom/test-utils'
import {render, unmountComponentAtNode} from 'react-dom';
//import RankedPoolsTableF from '../RankedPoolsTableF.js';

beforeEach( ()=> { //before each test create a div element
    const elem = document.createElement('RankedPoolsTableF');
    elem.setAttribute('id', 'RankedPoolsTableF');
    document.body.appendChild(elem);
});

afterEach( ()=> { //remove div element so next test has clean <body>
    const elem = document.getElementById('RankedPoolsTableF');
    unmountComponentAtNode(elem);
    elem.remove();
})

test( 'Ranked Pools test)', () => {
    const elem = document.getElementById('RankedPoolsTableF');
    console.log('HTML ->', document.body.innerHTML);
    act( () => {
        render(<RankedPoolsTableF/>, elem);
    });
});
