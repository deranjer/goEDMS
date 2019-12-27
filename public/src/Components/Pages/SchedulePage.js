import React, { useState, useCallback } from "react";
import { makeStyles } from '@material-ui/core/styles';
import { Grid, Paper, Button } from "@material-ui/core";

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


const SchedulePage = props => {
    const classes = useStyles();

    return (
        <Grid container spacing={3}>
            <Grid item xs={12}>
                <Paper className={classes.heading}>goSimple EDMS Schedules</Paper>
            </Grid>
            <Grid item xs={4}>
                <Paper className={classes.paper}>Ingress:</Paper>
            </Grid>
            <Grid item xs={4}>
                <Paper className={classes.paper}>8000</Paper>
            </Grid>
            <Grid item xs={4}>
                <Paper className={classes.paper}><Button>Run Now</Button></Paper>
            </Grid>
        </Grid>
    )

}

export default SchedulePage;