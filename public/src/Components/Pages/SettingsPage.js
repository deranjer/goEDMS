import React, { useState, useCallback } from "react";
import { makeStyles } from '@material-ui/core/styles';
import { Grid, Paper } from "@material-ui/core";

const useStyles = makeStyles(theme => ({
    root: {
      flexGrow: 1,
    },
    heading: {
      padding: theme.spacing(2),
      textAlign: 'center',
      color: theme.palette.text.secondary,
    },
    paper: {
        padding: theme.spacing(2),
        textAlign: 'left',
        color: theme.palette.text.secondary,
    },
  }));


const SettingsPage = props => {
    const classes = useStyles();

    return (
        <Grid container spacing={3}>
            <Grid item xs={12}>
                <Paper className={classes.heading}>goSimple EDMS Settings</Paper>
            </Grid>
            <Grid item xs={6}>
                <Paper className={classes.paper}>ServerPort:</Paper>
            </Grid>
            <Grid item xs={6}>
                <Paper className={classes.paper}>8000</Paper>
            </Grid>
        </Grid>
    )

}

export default SettingsPage;