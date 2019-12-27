import React, { useCallback, useState } from "react";
import { Route, Switch, BrowserRouter as Router, Link } from "react-router-dom";
import { AppBar, Toolbar, IconButton, Typography, Menu, MenuItem, Button, ListItem, Drawer, List } from "@material-ui/core";
import MenuIcon from '@material-ui/icons/Menu';
import { makeStyles } from '@material-ui/core/styles';

import HomePage from "./Components/Pages/HomePage";
import SettingsPage from "./Components/Pages/SettingsPage";
import SchedulePage from "./Components/Pages/SchedulePage";
import LogsPage from "./Components/Pages/LogsPage";

const useStyles = makeStyles({
    list: {
      width: 250,
    },
  });


const App = props => {
    const classes = useStyles();
    const [open, setOpen] = React.useState(false);

    const toggleDrawer = event => {
      if (event.type === 'keydown' && (event.key === 'Tab' || event.key === 'Shift')) {
        return;
      }
      if (open === true){
        setOpen(false)
      }
      if (open === false){
        setOpen(true)
      }
    }
    return (
        <Router>
            <div id="root">
                <React.Fragment>
                <Drawer open={open} onClose={toggleDrawer}>
                    <div
                    className={classes.list}
                    role="presentation"
                    onClick={toggleDrawer}
                    onKeyDown={toggleDrawer}
                    >
                        <List>
                            <ListItem button component={Link} to={"/"} exact="true" key="Home" onClick={toggleDrawer}>Home</ListItem>
                            <ListItem button component={Link} to={"/settings"} exact="true" key="Settings" onClick={toggleDrawer}>Settings</ListItem>
                            <ListItem button component={Link} to={"/schedule"} exact="true" key="Schedule" onClick={toggleDrawer}>Schedule</ListItem>
                            <ListItem button component={Link} to={"/logs"} exact="true" key="Logs" onClick={toggleDrawer}>Logs</ListItem>
                            </List>
                    </div>
                </Drawer>
                <AppBar position="static">
                    <Toolbar>
                    <IconButton edge="start" color="inherit" aria-label="menu" onClick={toggleDrawer}>
                    <MenuIcon />
                    </IconButton>
                    <Typography variant="h6">
                        GoSimple EDMS
                    </Typography>
                    </Toolbar>
                </AppBar>
                <Switch>
                    <Route path="/" exact component={HomePage} />
                    <Route path="/settings" exact component={SettingsPage} />
                    <Route path="/schedule" exact component={SchedulePage} />
                    <Route path="/logs" exact component={LogsPage} />
                </Switch>
                </React.Fragment>
            </div>
      </Router>
    )
}
export default App;