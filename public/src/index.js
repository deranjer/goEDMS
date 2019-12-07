import React from 'react';
import App from './App';
import ReactDOM from 'react-dom';
//import 'typeface-roboto';
import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

toast.configure({
    autoClose: 8000,
    draggable: false,
    position: toast.POSITION.BOTTOM_RIGHT
})

ReactDOM.render(<App />, document.getElementById('root'));