import { withStyles } from '@material-ui/core/styles';
import Paper from '@material-ui/core/Paper';
import gql from 'graphql-tag';
import React from 'react';
import Author from '../Author';
import { Avatar } from '../Author';
import Date from '../Date';

const styles = theme => ({
  author: {
    fontWeight: 'bold',
  },
  container: {
    display: 'flex',
  },
  avatar: {
    marginTop: 2,
  },
  bubble: {
    flex: 1,
    marginLeft: theme.spacing.unit,
  },
  header: {
    ...theme.typography.body2,
    color: '#444',
    padding: '0.5rem 1rem',
    borderBottom: '1px solid #ddd',
    display: 'flex',
  },
  title: {
    flex: 1,
  },
  tag: {
    ...theme.typography.button,
    color: '#888',
    border: '#ddd solid 1px',
    padding: '0 0.5rem',
    fontSize: '0.75rem',
    borderRadius: 2,
    marginLeft: '0.5rem',
  },
  body: {
    ...theme.typography.body1,
    padding: '1rem',
    whiteSpace: 'pre-wrap',
  },
});

const Message = ({ op, classes }) => (
  <article className={classes.container}>
    <Avatar author={op.author} className={classes.avatar} />
    <Paper elevation={1} className={classes.bubble}>
      <header className={classes.header}>
        <div className={classes.title}>
          <Author className={classes.author} author={op.author} />
          <span> commented </span>
          <Date date={op.createdAt} />
        </div>
        {op.edited && <div className={classes.tag}>Edited</div>}
      </header>
      <section className={classes.body}>{op.message}</section>
    </Paper>
  </article>
);

Message.createFragment = gql`
  fragment Create on TimelineItem {
    ... on CreateTimelineItem {
      createdAt
      author {
        ...Author
      }
      edited
      message
    }
  }
  ${Author.fragment}
`;

Message.commentFragment = gql`
  fragment AddComment on TimelineItem {
    ... on AddCommentTimelineItem {
      createdAt
      author {
        ...Author
      }
      edited
      message
    }
  }
  ${Author.fragment}
`;

export default withStyles(styles)(Message);
