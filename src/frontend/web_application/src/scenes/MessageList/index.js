import { createSelector } from 'reselect';
import { bindActionCreators, compose } from 'redux';
import { connect } from 'react-redux';
import { withTranslator } from '@gandi/react-translate';
import { requestMessages, updateMessage } from '../../store/modules/message';
import Presenter from './presenter';

const messagesSelector = state => state.message.messages;
const messageByIdSelector = state => state.message.messagesById;
const discussionIdSelector = (state, ownProps) => ownProps.params.discussionId;

const mapStateToProps = createSelector(
  [messagesSelector, messageByIdSelector, discussionIdSelector],
  (messages, messagesById, discussionId) => ({
    discussionId,
    messages: messages
    .map(messageId => messagesById[messageId])
    .filter(message => message.discussion_id === discussionId),
  })
);

const mapDispatchToProps = dispatch => bindActionCreators({
  requestMessages,
  updateMessage,
}, dispatch);

export default compose(
  connect(mapStateToProps, mapDispatchToProps),
  withTranslator()
)(Presenter);