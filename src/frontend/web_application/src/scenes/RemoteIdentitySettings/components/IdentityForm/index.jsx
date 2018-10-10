import React, { Component, Fragment } from 'react';
import PropTypes from 'prop-types';
import { Trans } from 'lingui-react';
import { RadioFieldGroup, Button, Section, FormGrid, FormRow, FormColumn, Link, Callout, Icon } from '../../../../components';
import { PROTOCOL_IMAP, PROTOCOL_GMAIL, PROTOCOL_TWITTER } from '../../../../modules/remoteIdentity';
import RemoteIdentityEmail from '../RemoteIdentityEmail';
import RemoteIdentityGmail from '../RemoteIdentityGmail';
import RemoteIdentityTwitter from '../RemoteIdentityTwitter';

const iconTypeProtocols = {
  [PROTOCOL_IMAP]: 'email',
  [PROTOCOL_GMAIL]: 'google',
  [PROTOCOL_TWITTER]: 'twitter',
};

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
    const result = await onRemoteIdentityChange(...params);

    if (this.state.newRemoteIdentity) {
      this.setState({
        newRemoteIdentity: undefined,
      });
    }

    return result;
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

    const renderComponent = C => (
      <C
        key={remoteIdentity.identity_id || 'new'}
        remoteIdentity={remoteIdentity}
        onChange={this.handleChange}
        onDelete={onRemoteIdentityDelete}
        onCancel={this.handleCancel}
      />
    );

    switch (remoteIdentity.protocol) {
      default:
      case PROTOCOL_IMAP:
        return renderComponent(RemoteIdentityEmail);
      case PROTOCOL_GMAIL:
        return renderComponent(RemoteIdentityGmail);
      case PROTOCOL_TWITTER:
        return renderComponent(RemoteIdentityTwitter);
    }
  }

  renderCreate() {
    const options = [
      { value: PROTOCOL_IMAP, label: (<Fragment><Icon type={iconTypeProtocols[PROTOCOL_IMAP]} rightSpaced /><Trans id="remote_identity.protocol.mail">Mail</Trans></Fragment>) },
      { value: PROTOCOL_GMAIL, label: (<Fragment><Icon type={iconTypeProtocols[PROTOCOL_GMAIL]} rightSpaced /><Trans id="remote_identity.protocol.gmail">Gmail</Trans></Fragment>) },
      { value: PROTOCOL_TWITTER, label: (<Fragment><Icon type={iconTypeProtocols[PROTOCOL_TWITTER]} rightSpaced /><Trans id="remote_identity.protocol.twitter">Twitter</Trans></Fragment>) },
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

  renderTitle = remoteIdentity => (
    <Fragment>
      <Icon type={iconTypeProtocols[remoteIdentity.protocol]} rightSpaced />
      {remoteIdentity.display_name}
    </Fragment>
  );

  render() {
    const remoteIdentity = this.props.remoteIdentity || this.state.newRemoteIdentity;
    if (remoteIdentity) {
      const context = remoteIdentity.status === 'active' ? 'safe' : 'disabled';

      return (
        <Section
          title={this.renderTitle(remoteIdentity)}
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
