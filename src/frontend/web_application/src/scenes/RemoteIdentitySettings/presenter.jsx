import React, { Component, Fragment } from 'react';
import PropTypes from 'prop-types';
import IdentityForm from './components/IdentityForm';
import { PageTitle } from '../../components/';

import './style.scss';

class RemoteIdentitySettings extends Component {
  static propTypes = {
    requestRemoteIdentities: PropTypes.func.isRequired,
    onIdentityChange: PropTypes.func.isRequired,
    onIdentityDelete: PropTypes.func.isRequired,
    // eslint-disable-next-line react/no-unused-prop-types_
    identities: PropTypes.arrayOf(PropTypes.shape({})),
  };
  static defaultProps = {
    identities: undefined,
  };

  componentDidMount() {
    this.props.requestRemoteIdentities();
  }

  handleChange = async (...params) => {
    const { onIdentityChange } = this.props;

    return onIdentityChange(...params);
  }

  handleDelete = async (...params) => {
    const { onIdentityDelete } = this.props;

    return onIdentityDelete(...params);
  }

  renderIdentity(identity) {
    return (
      <div className="s-settings-identities__identity" key={identity.identity_id}>
        <IdentityForm
          remoteIdentity={identity}
          onRemoteIdentityChange={this.handleChange}
          onRemoteIdentityDelete={this.handleDelete}
        />
      </div>
    );
  }

  render() {
    const { onIdentityChange, onIdentityDelete, identities } = this.props;

    return (
      <div className="s-settings-identities">
        <PageTitle />
        <div className="s-settings-identities__create">
          <IdentityForm
            onRemoteIdentityChange={onIdentityChange}
            onRemoteIdentityDelete={onIdentityDelete}
          />
        </div>
        <Fragment>
          {
            identities
              .map(identity => this.renderIdentity(identity))
          }
        </Fragment>
      </div>
    );
  }
}

export default RemoteIdentitySettings;
