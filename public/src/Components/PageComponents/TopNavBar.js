import React, { useState, useCallback } from "react";
import { AppBar, Toolbar, IconButton, Typography, Menu, MenuItem, Button, ListItem, Drawer, List } from "@material-ui/core";
import MenuIcon from '@material-ui/icons/Menu';


const TopNavBar = props => {
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
      <React.Fragment>
      <Drawer open={open} onClose={toggleDrawer}>
      <List>
          <ListItem button key="Settings">Settings</ListItem>
          <ListItem button key="Schedules">Schedules</ListItem>
          <ListItem button key="Logs">Logs</ListItem>
         </List>

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
      </React.Fragment>
    );
  };
  
export default TopNavBar;