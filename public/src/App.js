import React, { useCallback, useState, useEffect } from "react";
import useDebounce from './Components/Hooks/use-debounce';
import { Route, Switch, BrowserRouter as Router, Link } from "react-router-dom";
import { AppBar, Toolbar, IconButton, Typography, Menu, MenuItem, Button, ListItem, Drawer, List } from "@material-ui/core";
import MenuIcon from '@material-ui/icons/Menu';
import { makeStyles, fade } from '@material-ui/core/styles';
import InputBase from '@material-ui/core/InputBase';
import SearchIcon from '@material-ui/icons/Search';

import HomePage from "./Components/Pages/HomePage";
import SettingsPage from "./Components/Pages/SettingsPage";
import SchedulePage from "./Components/Pages/SchedulePage";
import LogsPage from "./Components/Pages/LogsPage";

const useStyles = makeStyles(theme => ({
    list: {
      width: 250,
    },
    root: {
      flexGrow: 1,
    },
    menuButton: {
      marginRight: theme.spacing(2),
    },
    title: {
      flexGrow: 1,
      display: 'none',
      [theme.breakpoints.up('sm')]: {
        display: 'block',
      },
    },
    search: {
      position: 'relative',
      marginLeft: '10px',
      width: 'auto',
      float: 'right',
      justifyContent: 'right',
      alignItems: 'right',
       borderRadius: theme.shape.borderRadius,
      backgroundColor: fade(theme.palette.common.white, 0.15),
      '&:hover': {
        backgroundColor: fade(theme.palette.common.white, 0.25),
      },
       [theme.breakpoints.up('sm')]: {
        marginLeft: theme.spacing(1),
        width: 'auto',
      },  
    },
    searchIcon: {
      width: theme.spacing(7),
      height: '100%',
      position: 'absolute',
      pointerEvents: 'none',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    },
    inputRoot: {
      color: 'inherit',
    },
    inputInput: {
      float: 'right',
      padding: theme.spacing(1, 1, 1, 7),
      transition: theme.transitions.create('width'),
      width: 'auto',
      [theme.breakpoints.up('sm')]: {
        width: 120,
        '&:focus': {
          width: 200,
        },
      }, 
    },
  }));


const App = props => {
    const classes = useStyles();
    const [open, setOpen] = React.useState(false);
    const [searchTerm, setSearchTerm] = React.useState(''); //when search term is complete this is updated
    const [isSearching, setIsSearching] = React.useState(false);
    const [tempSearch, setTempSearch] = React.useState(''); //temporarily holds the search term

    const debouncedSearchTerm = useDebounce(tempSearch, 1000);

    useEffect(() => {
      if (debouncedSearchTerm) {
        setIsSearching(true);
        setSearchTerm(tempSearch)
        setIsSearching(false)
      }
    }, [debouncedSearchTerm])

    const handleKeyDown = (value) => {
      if (e.keyCode === 13) {
        setSearchTerm(tempSearch)
      }
    }

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
                <div className={classes.root}>
                  <AppBar position="static">
                      <Toolbar>
                      <IconButton edge="start" color="inherit" aria-label="menu" onClick={toggleDrawer}>
                      <MenuIcon />
                      </IconButton>
                      <Typography variant="h6" className={classes.title} noWrap>
                          GoSimple EDMS
                      </Typography>
                        <div className={classes.search}>
                          <div className={classes.searchIcon}>
                            <SearchIcon />
                          </div>
                        <InputBase
                          placeholder="Search…"
                          classes={{
                            root: classes.inputRoot,
                            input: classes.inputInput,
                          }}
                          inputProps={{ 'aria-label': 'search' }}
                          //value={searchTerm}
                          onChange={(e) => setTempSearch(e.target.value)}
                          onKeyDown={(e) => handleKeyDown(e.target.value)}
                        />
                        </div>
                      </Toolbar>
                  </AppBar>
                  
                </div>
                <Switch>
                    <Route path="/" exact render={(props) => <HomePage searchTerm={searchTerm} />} />
                    <Route path="/settings" exact component={SettingsPage} />
                    <Route path="/schedule" exact component={SchedulePage} />
                    <Route path="/logs" exact component={LogsPage} />
                </Switch>
                </React.Fragment>
                {isSearching && <div>Searching ...</div>}
            </div>
      </Router>
    )
}
export default App;