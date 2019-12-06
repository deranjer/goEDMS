import React, { useState, useCallback } from "react";
import { Link } from "react-router-dom";
import { Button, Toolbar, AppBar, Grid, Avatar, Badge, Typography } from "@material-ui/core";
import { Menu as DropMenu, MenuItem, Hidden } from "@material-ui/core";
import Chat from "@material-ui/icons/Chat";
import Menu from "@material-ui/icons/Menu";
// import SearchBar from '../searchBar';


const TopNavBar = props => {
  return (
    <AppBar position="sticky" style={{ marginBottom: "1%", padding: "0" }}>
      <Toolbar>
        <Grid container>
          <Grid style={{ marginLeft: "auto" }}>
            <Button style={{ borderRadius: 35 }}>
              <Avatar>H</Avatar>
            </Button>
          </Grid>
        </Grid>
      </Toolbar>
    </AppBar>
  );
};

export default TopNavBar;