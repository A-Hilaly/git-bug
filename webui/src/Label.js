import React from 'react';
import { withStyles } from '@material-ui/core/styles';
import {
  getContrastRatio,
  darken,
} from '@material-ui/core/styles/colorManipulator';
import { common } from '@material-ui/core/colors';

// Minimum contrast between the background and the text color
const contrastThreshold = 2.5;

// Guess the text color based on the background color
const getTextColor = background =>
  getContrastRatio(background, common.white) >= contrastThreshold
    ? common.white // White on dark backgrounds
    : common.black; // And black on light ones

const _rgb = (color) => 'rgb(' + color.R + ',' + color.G + ',' + color.B + ')';

// Create a style object from the label RGB colors
const createStyle = color => ({
  backgroundColor: _rgb(color),
  color: getTextColor(_rgb(color)),
  borderBottomColor: darken(_rgb(color), 0.2),
});

const styles = theme => ({
  label: {
    ...theme.typography.body2,
    padding: '0 6px',
    fontSize: '0.9em',
    margin: '0 1px',
    borderRadius: '3px',
    display: 'inline-block',
    borderBottom: 'solid 1.5px',
    verticalAlign: 'bottom',
  },
});

const Label = ({ label, classes }) => (
  <span className={classes.label} style={createStyle(label.color)}>
    {label.name}
  </span>
);

export default withStyles(styles)(Label);
