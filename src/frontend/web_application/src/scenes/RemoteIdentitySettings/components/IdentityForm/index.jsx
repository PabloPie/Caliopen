import React, { Component } from 'react';
import PropTypes from 'prop-types';
import { Trans } from '@lingui/macro'; // eslint-disable-line import/no-extraneous-dependencies
import { RadioFieldGroup, Button, Section, FormGrid, FormRow, FormColumn, Link, Callout } from '../../../../components';
import RemoteIdentityEmail from '../RemoteIdentityEmail';

class IdentityForm extends Component {
  static propTypes = {
    remoteIdentity: PropTypes.shape({}),
    onRemoteIdentityChange: PropTypes.func.isRequired,
    onRemoteIdentityDelete: PropTypes.func.isRequired,
  };
  static defaultProps = {
    remoteIdentity: undefined,
  };

  state = {
    protocol: 'imap',
    newRemoteIdentity: undefined,
  };

  handleCreateChange = (event) => {
    const {
      target: {
        name, protocol, checked, value: inputValue,
      },
    } = event;
    const value = protocol === 'checkbox' ? checked : inputValue;

    this.setState({
      [name]: value,
    });
  }

  handleCreate = () => {
    this.setState(prevState => ({
      newRemoteIdentity: {
        protocol: prevState.protocol,
      },
    }));
  }

  handleChange = async (...params) => {
    const { onRemoteIdentityChange } = this.props;
    await onRemoteIdentityChange(...params);

    if (this.state.newRemoteIdentity) {
      this.setState({
        newRemoteIdentity: undefined,
      });
    }
  }

  handleCancel = () => {
    this.setState({
      newRemoteIdentity: undefined,
    });
  }

  renderType(remoteIdentity) {
    const {
      onRemoteIdentityDelete,
    } = this.props;

    switch (remoteIdentity.protocol) {
      default:
      case 'imap':
        return (
          <RemoteIdentityEmail
            key={remoteIdentity.identity_id || 'new'}
            remoteIdentity={remoteIdentity}
            onChange={this.handleChange}
            onDelete={onRemoteIdentityDelete}
            onCancel={this.handleCancel}
          />
        );
    }
  }

  renderCreate() {
    const options = [
      { value: 'imap', label: (<Trans id="remote_identity.protocol.mail">Mail</Trans>) },
    ];

    return (
      <Section
        title={(<Trans id="remote_identity.add_account">Add an external account</Trans>)}
      >
        <Callout color="info">
          <Trans id="remote_identity.how_to">
            <p>
              External accounts are fetched every 15 minutes.<br />
              Currently there is no indicator to inform that the account is correctly configured
              until first try is done.
            </p>
          </Trans>
        </Callout>
        <Callout color="warning">
          <Trans id="remote_identity.gmail_warning">
            <p>
              If you aim to add a Gmail account, please ensure that IMAP protocol is activated in
              your Gmail settings at <Link target="_blank" rel="noopener noreferrer" href="https://mail.google.com/mail/u/0/#settings/fwdandpop">“Forward and POP/IMAP”</Link>,
              and that <Link target="_blank" rel="noopener noreferrer" href="https://myaccount.google.com/lesssecureapps">Less secure application access</Link> is activated for your Google account.
            </p>
          </Trans>
        </Callout>
        <FormGrid>
          <FormRow>
            <FormColumn bottomSpace>
              <ul>
                <li>
                  <RadioFieldGroup
                    onChange={this.handleCreateChange}
                    value={this.state.protocol}
                    options={options}
                    name="protocol"
                  />
                </li>
              </ul>
            </FormColumn>
          </FormRow>
          <FormRow>
            <FormColumn bottomSpace>
              <Button onClick={this.handleCreate} shape="plain"><Trans id="remote_identity.action.continue">Continue</Trans></Button>
            </FormColumn>
          </FormRow>
        </FormGrid>
      </Section>
    );
  }

  render() {
    const remoteIdentity = this.props.remoteIdentity || this.state.newRemoteIdentity;
    if (remoteIdentity) {
      const context = remoteIdentity.status === 'active' ? 'safe' : 'disabled';

      return (
        <Section
          title={remoteIdentity.display_name}
          borderContext={context}
        >
          {this.renderType(remoteIdentity)}
        </Section>
      );
    }

    return this.renderCreate();
  }
}

export default IdentityForm;
