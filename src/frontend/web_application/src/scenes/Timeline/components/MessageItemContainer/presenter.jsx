import React, { Component } from 'react';
import PropTypes from 'prop-types';
import Swipeable from 'react-swipeable';
import classnames from 'classnames';
import { Trans } from 'lingui-react';
import Button from '../../../../components/Button';
import Modal from '../../../../components/Modal';
import ManageTags from '../ManageTags';

import './style.scss';

class MessageItemContainer extends Component {
  static propTypes = {
    message: PropTypes.shape({}).isRequired,
    children: PropTypes.node.isRequired,
    i18n: PropTypes.shape({}).isRequired,
    replyToMessage: PropTypes.func.isRequired,
    deleteMessage: PropTypes.func.isRequired,
  };

  state = {
    isActive: false,
    isSwiped: false,
    isTagModalOpen: false,
  };

  handleOpenTags = () => {
    this.setState(prevState => ({
      ...prevState,
      isTagModalOpen: true,
    }));
  }

  handleCloseTags = () => {
    this.setState(prevState => ({
      ...prevState,
      isTagModalOpen: false,
    }));
  }

  handleReply = () => {
    const { replyToMessage, message } = this.props;
    replyToMessage({ internalId: message.discussion_id, message });
  }

  handleDelete = () => {
    const { deleteMessage, message } = this.props;
    deleteMessage({ message });
  }

  handleSwipeLeft = () => {
    this.setState(prevState => ({
      ...prevState,
      isActive: true,
      isSwiped: true,
    }));
  }

  handleSwipeRight = () => {
    this.setState(prevState => ({
      ...prevState,
      isActive: prevState.isTagModalOpen === true ? prevState.isActive : false,
      isSwiped: prevState.isTagModalOpen === true ? prevState.isSwiped : false,
    }));
  }

  handleHover = () => {
    this.setState(prevState => ({
      ...prevState,
      isActive: prevState.isSwiped ? prevState.isActive : true,
    }));
  }

  handleBlur = () => {
    this.setState((prevState) => {
      const isActive = prevState.isSwiped ? false : prevState.isActive;

      return {
        ...prevState,
        isActive: prevState.isTagModalOpen === true ? isActive : false,
      };
    });
  }

  renderActions= () => {
    const { message } = this.props;
    const actionsClassName = classnames(
      'm-message-item-container__actions',
      { 'm-message-item-container__actions--swiped': this.state.isSwiped },
      { 'm-message-item-container__actions--hover': this.state.isActive },
    );

    return (
      <div className={actionsClassName}>
        <Button shape="plain" className="m-message-item-container__action" onClick={this.handleDelete}>
          <Trans id="message-item.action.delete">message-item.action.delete</Trans>
        </Button>
        {!message.is_draft && (<Button shape="plain" className="m-message-item-container__action" onClick={this.handleReply}>
          <Trans id="message-item.action.reply">message-item.action.reply</Trans>
        </Button>)}
        <Button shape="plain" className="m-message-item-container__action" onClick={this.handleOpenTags}>
          <Trans id="message-item.action.manage_tags">message-item.action.manage_tags</Trans>
        </Button>
        {this.renderTagsModal()}
      </div>
    );
  }


  renderTagsModal = () => {
    const { message, i18n } = this.props;
    const count = message.tags ? message.tags.length : 0;
    const title = [
      i18n.t`tags.header.title`,
      (<span key="1" className="m-tags-form__count"><Trans id="tags.header.count" values={{ count }}>tags.header.count</Trans></span>),
    ];

    return (
      <Modal
        isOpen={this.state.isTagModalOpen}
        contentLabel={i18n.t`tags.header.title`}
        title={title}
        onClose={this.handleCloseTags}
      >
        <ManageTags message={message} />
      </Modal>
    );
  }

  render() {
    const { children } = this.props;

    return (
      <Swipeable
        onSwipingLeft={this.handleSwipeLeft}
        onSwipingRight={this.handleSwipeRight}
        onMouseOver={this.state.isSwiped ? null : this.handleHover}
        onMouseOut={this.state.isSwiped ? null : this.handleBlur}
        className={classnames(
          'm-message-item-container',
          { 'm-message-item-container--active': this.state.isActive },
        )}
      >
        {children}
        {this.renderActions()}
      </Swipeable>
    );
  }
}

export default MessageItemContainer;
