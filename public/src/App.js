import React, { useCallback, useState } from "react";

import TopNavBar from "./Components/PageComponents/TopNavBar";
import { Route, Switch, BrowserRouter as Router, Link } from "react-router-dom";


const App = props => {
    return (
        <Router>
            <div id="root">
                <TopNavBar/>
            </div>
      </Router>
    )
}
export default App;