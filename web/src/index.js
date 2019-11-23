import React from 'react';
import ReactDOM from 'react-dom';
import "antd/dist/antd.css";

import './index.sass';
import App from './App';
import * as serviceWorker from './serviceWorker';
import "./request-interceptors";

ReactDOM.render(<App />, document.getElementById('root'));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
