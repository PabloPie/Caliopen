import React from 'react';
// import PropTypes from 'prop-types';
// import { createSelector } from 'reselect';

const popups = {};

export const withAuthorizePopup = () => (C) => {
  const getName = ({ identity }) => `authorization_${identity.protocol}`;

  // a popup must be initialized in a synchronous way or it will be blocked by the browser
  const initPopup = ({ identity }) => {
    const name = getName({ identity });
    if (!popups[name] || popups[name].closed) {
      popups[name] = window.open('', name, 'resizable,scrollbars,status');
    }
  };
  const authorizePopup = ({ identity }) => {
    const name = getName({ identity });

    if (popups[name] && !popups[name].closed) {
      popups[name].location.href = identity.infos.authorization_popup_url;
    }
  };

  return props => (<C authorizePopup={authorizePopup} initPopup={initPopup} {...props} />);
};
