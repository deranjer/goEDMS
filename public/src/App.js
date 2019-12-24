import React, { useCallback, useState } from "react";

import TopNavBar from "./Components/PageComponents/TopNavBar";
import { Route, Switch, BrowserRouter as Router, Link } from "react-router-dom";
import FolderTree from "./Components/PageComponents/FolderTree";

const App = props => {
    return (
        <Router>
            <div id="root">
                <React.Fragment>
                <TopNavBar />
                <FolderTree />
                </React.Fragment>
            </div>
      </Router>
    )
}
export default App;