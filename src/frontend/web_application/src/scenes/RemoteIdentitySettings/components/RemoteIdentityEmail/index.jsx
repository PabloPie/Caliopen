import React, { Component } from 'react';
import PropTypes from 'prop-types';
import { Trans, withI18n } from 'lingui-react';
import { Confirm, Button, TextFieldGroup, SelectFieldGroup, CheckboxFieldGroup, FormGrid, FormRow, FormColumn, FieldErrors } from '../../../../components';
import { REMOTE_IDENTITY_STATUS_ACTIVE, REMOTE_IDENTITY_STATUS_INACTIVE, PROTOCOL_IMAP, Identity } from '../../../../modules/remoteIdentity';
import Phase0 from '../Phase0';
import Status from '../Status';
import './style.scss';

const MAX_PHASE = 3;
const MAIL_PROTOCOLS = [PROTOCOL_IMAP];

function generateStateFromProps(props, prevState) {
  const {
    remoteIdentity: {
      identity_id: identityId, infos, credentials, status, protocol,
      identifier,
    },
  } = props;
  const [inserverHostname = '', inserverPort = ''] = infos && infos.inserver ? infos.inserver.split(':') : [];
  const active = status ?
    status === REMOTE_IDENTITY_STATUS_ACTIVE :
    prevState.remoteIdentity.active;

  return {
    phase: identityId ? 0 : 1,
    remoteIdentity: {
      ...prevState.remoteIdentity,
      ...(identifier ? { identifier } : {}),
      inserverHostname,
      inserverPort,
      ...(
        credentials ?
          {
            inusername: credentials.inusername,
            inpassword: credentials.inpassword,
          } : {}
      ),
      active,
      ...(protocol ? { protocol } : {}),
    },
  };
}

function getIdentityFromState(state, props) {
  const {
    remoteIdentity: {
      identifier, inserverHostname, inserverPort, inusername, inpassword, active, protocol,
    },
  } = state;
  const { remoteIdentity } = props;
  const credentials = (inusername || inpassword) ? {
    ...(remoteIdentity.credentials ? remoteIdentity.credentials : {}),
    inusername,
    inpassword,
  } : undefined;

  return new Identity({
    ...remoteIdentity,
    credentials,
    display_name: identifier,
    identifier,
    infos: {
      ...(remoteIdentity.infos ? remoteIdentity.infos : {}),
      inserver: `${inserverHostname}:${inserverPort}`,
    },
    protocol,
    status: active ? REMOTE_IDENTITY_STATUS_ACTIVE : REMOTE_IDENTITY_STATUS_INACTIVE,
  });
}

@withI18n()
class RemoteIdentityEmail extends Component {
  static propTypes = {
    remoteIdentity: PropTypes.shape({}).isRequired,
    onChange: PropTypes.func.isRequired,
    onDelete: PropTypes.func.isRequired,
    onCancel: PropTypes.func.isRequired,
    i18n: PropTypes.shape({}).isRequired,
  };

  static defaultProps = {
  }

  state = {
    phase: 0,
    formErrors: {},
    advancedForm: false,
    remoteIdentity: {
      identifier: '',
      inserverHostname: '',
      inserverPort: '',
      inusername: '',
      inpassword: '',
      active: true,
      protocol: PROTOCOL_IMAP,
    },
  };

  componentWillMount() {
    this.setState(prevState => generateStateFromProps(this.props, prevState));
  }

  handleEdit = () => {
    this.setState({ phase: 1 });
  }

  handleClickPrevious = () => {
    this.setState(prevState => ({
      phase: prevState.phase - 1,
    }));
  }

  handleClickNext = () => {
    if (this.validate({ phase: this.state.phase })) {
      this.setState((prevState) => {
        const { remoteIdentity } = prevState;
        if (!remoteIdentity.entity_id && !remoteIdentity.inusername.length) {
          remoteIdentity.inusername = remoteIdentity.identifier;
        }

        return {
          phase: prevState.phase + 1,
          remoteIdentity,
        };
      });
    }
  }

  handleSave = async () => {
    const { remoteIdentity } = this.props;

    try {
      await this.props.onChange({
        identity: getIdentityFromState(this.state, this.props),
      });
      if (remoteIdentity.identity_id) {
        this.setState({ phase: 0 });
      }
    } catch (errs) {
      if (errs.some(err => err.code === 6)) {
        this.setState({
          phase: 1,
          formErrors: { identifier: [(<Trans id="remote_identity.form.identifier.error.uniqueness">This address is already configured</Trans>)] },
        });
      } else {
        this.setState({
          formErrors: { global: errs.map(({ message }) => message) },
        });
      }
    }
  }

  handleDelete = () => {
    this.props.onDelete({ identity: getIdentityFromState(this.state, this.props) });
  }

  handleCancel = () => {
    const { onCancel } = this.props;
    this.setState(prevState => generateStateFromProps(this.props, prevState), () => {
      onCancel();
    });
  }

  handleParamsChange = (event) => {
    const { name, value } = event.target;

    this.setState(prevState => ({
      remoteIdentity: {
        ...prevState.remoteIdentity,
        [name]: value,
      },
    }));
  }

  handleActivate = (active) => {
    this.setState(prevState => ({
      remoteIdentity: {
        ...prevState.remoteIdentity,
        active,
      },
    }), () => {
      this.props.onChange({ identity: getIdentityFromState(this.state, this.props) });
    });
  }

  validate = ({ phase }) => {
    let isValid = true;

    const phaseValidation = (properties) => {
      properties.forEach(({ formProperty, error }) => {
        const value = this.state.remoteIdentity[formProperty];

        if (!value || value.length === 0) {
          this.setState(prevState => ({
            formErrors: {
              ...prevState.formErrors,
              [formProperty]: [error],
            },
          }));
          isValid = false;
        } else {
          this.setState(prevState => ({
            formErrors: {
              ...prevState.formErrors,
              [formProperty]: [],
            },
          }));
        }
      });
    };

    if (phase >= 1) {
      phaseValidation([
        { formProperty: 'identifier', error: <Trans id="remote_identity.form.identifier.error">a valid email is required</Trans> },
      ]);
    }

    if (phase >= 3) {
      phaseValidation([
        { formProperty: 'inusername', error: <Trans id="remote_identity.form.username.error">login is required</Trans> },
        { formProperty: 'inpassword', error: <Trans id="remote_identity.form.password.error">password is required</Trans> },
      ]);
    }

    if (phase >= 2) {
      phaseValidation([
        { formProperty: 'inserverHostname', error: <Trans id="remote_identity.form.serverHostname.error">mail server is required</Trans> },
        { formProperty: 'inserverPort', error: <Trans id="remote_identity.form.serverPort.error">port is required</Trans> },
        { formProperty: 'protocol', error: <Trans id="remote_identity.form.protocol.error">protocol is required</Trans> },
      ]);
    }

    return isValid;
  }

  renderFormPhase0() {
    const { remoteIdentity } = this.props;

    return (
      <Phase0
        onToggleActivate={this.handleActivate}
        remoteIdentity={remoteIdentity}
        active={this.state.remoteIdentity.active}
        errors={this.state.formErrors}
      />
    );
  }

  renderFormPhase1() {
    const { remoteIdentity, i18n } = this.props;

    return (
      <FormGrid>
        <FormRow>
          <FormColumn bottomSpace>
            <TextFieldGroup
              type="email"
              label={<Trans id="remote_identity.form.identifier.label">Email:</Trans>}
              placeholder={i18n._('remote_identity.form.identifier.placeholder', { defaults: 'john@doe.tld' })}
              value={this.state.remoteIdentity.identifier}
              errors={this.state.formErrors.identifier}
              onChange={this.handleParamsChange}
              name="identifier"
              // specificity of backend: the identifier and protocol are unique and immutable
              // cf. https://github.com/CaliOpen/Caliopen/blob/d6bbe43cc1098844f53eaec283e08c19c5f871bc/doc/specifications/identities/index.md#model
              disabled={remoteIdentity.identity_id && true}
              autoComplete="email"
              required
            />
          </FormColumn>
          <FormColumn bottomSpace>
            {remoteIdentity.identity_id && (
              <CheckboxFieldGroup
                value={this.state.remoteIdentity.active}
                errors={this.state.formErrors.status}
                onChange={this.handleParamsChange}
                name="active"
                label={<Status status={this.state.remoteIdentity.active ? 'active' : 'inactive'} />}
                displaySwitch
                showTextLabel
              />
            )}
          </FormColumn>
        </FormRow>
      </FormGrid>
    );
  }

  renderFormPhase2() {
    return (
      <FormGrid>
        <FormRow>
          <FormColumn bottomSpace>
            <SelectFieldGroup
              label={<Trans id="remote_identity.form.protocol.label">Protocol:</Trans>}
              value={this.state.remoteIdentity.protocol}
              options={MAIL_PROTOCOLS.map(key => ({ value: key, label: key }))}
              errors={this.state.formErrors.protocol}
              onChange={this.handleParamsChange}
              name="protocol"
              required
              disabled
            />
          </FormColumn>
          <FormColumn bottomSpace>
            <TextFieldGroup
              label={<Trans id="remote_identity.form.incomming_mail_server.label">Incoming mail server:</Trans>}
              value={this.state.remoteIdentity.inserverHostname}
              errors={this.state.formErrors.inserverHostname}
              onChange={this.handleParamsChange}
              name="inserverHostname"
              autoComplete="on"
              required
            />
          </FormColumn>
          <FormColumn bottomSpace>
            <TextFieldGroup
              label={<Trans id="remote_identity.form.port.label">Port:</Trans>}
              value={this.state.remoteIdentity.inserverPort}
              errors={this.state.formErrors.inserverPort}
              onChange={this.handleParamsChange}
              name="inserverPort"
              type="number"
              autoComplete="on"
              required
            />
          </FormColumn>
        </FormRow>
      </FormGrid>
    );
  }

  renderFormPhase3() {
    return (
      <FormGrid>
        <FormRow>
          <FormColumn bottomSpace>
            <TextFieldGroup
              label={<Trans id="remote_identity.form.username.label">Login:</Trans>}
              value={this.state.remoteIdentity.inusername}
              errors={this.state.formErrors.inusername}
              onChange={this.handleParamsChange}
              name="inusername"
              autoComplete="username"
              required
            />
          </FormColumn>
          <FormColumn bottomSpace>
            <TextFieldGroup
              label={<Trans id="remote_identity.form.password.label">Password:</Trans>}
              type="password"
              value={this.state.remoteIdentity.inpassword}
              errors={this.state.formErrors.inpassword}
              onChange={this.handleParamsChange}
              name="inpassword"
              autoComplete="new-password"
              required
            />
          </FormColumn>
        </FormRow>
      </FormGrid>
    );
  }

  renderActionsCreate() {
    return (
      <div>
        {this.state.phase >= 1 && (
          <Button onClick={this.handleCancel} shape="plain" className="m-remote-identity-email__action">
            <Trans id="remote_identity.action.cancel">Cancel</Trans>
          </Button>
        )}
        {this.state.phase > 1 && (
          <Button onClick={this.handleClickPrevious} shape="hollow" className="m-remote-identity-email__action">
            <Trans id="remote_identity.action.back">Previous</Trans>
          </Button>
        )}
        {this.state.phase < MAX_PHASE && (
          <Button shape="plain" onClick={this.handleClickNext} className="m-remote-identity-email__action">
            <Trans id="remote_identity.action.next">Next</Trans>
          </Button>
        )}
        {this.state.phase === MAX_PHASE && (
          <Button onClick={this.handleSave} shape="plain" color="secondary" className="m-remote-identity-email__action m-remote-identity-email__action--push-right">
            <Trans id="remote_identity.action.finish">Connect</Trans>
          </Button>
        )}
      </div>
    );
  }

  renderActionsEdit() {
    return (
      <div>
        <Confirm
          title={<Trans id="remote_identity.confirm-delete.title">Delete the external account</Trans>}
          content={<Trans id="remote_identity.confirm-delete.content">The external account will deactivated then deleted.</Trans>}
          onConfirm={this.handleDelete}
          render={confirm => (
            <Button onClick={confirm} shape="plain" color="alert" className="m-remote-identity-email__action">
              <Trans id="remote_identity.action.delete">Delete</Trans>
            </Button>
          )}
        />
        {this.state.phase === 0 && (
          <Button onClick={this.handleEdit} shape="hollow" className="m-remote-identity-email__action">
            <Trans id="remote_identity.action.edit">Edit</Trans>
          </Button>
        )}
        {this.state.phase >= 1 && (
          <Button onClick={this.handleCancel} shape="hollow" className="m-remote-identity-email__action">
            <Trans id="remote_identity.action.cancel">Cancel</Trans>
          </Button>
        )}
        {this.state.phase > 1 && (
          <Button onClick={this.handleClickPrevious} shape="hollow" className="m-remote-identity-email__action">
            <Trans id="remote_identity.action.back">Previous</Trans>
          </Button>
        )}
        {this.state.phase >= 1 && this.state.phase < MAX_PHASE && (
          <Button shape="plain" onClick={this.handleClickNext} className="m-remote-identity-email__action">
            <Trans id="remote_identity.action.next">Next</Trans>
          </Button>
        )}
        {this.state.phase >= 1 && (
          <Button onClick={this.handleSave} shape="plain" className="m-remote-identity-email__action m-remote-identity-email__action--push-right">
            <Trans id="remote_identity.action.save">Save</Trans>
          </Button>
        )}
      </div>
    );
  }

  renderActions() {
    const { remoteIdentity } = this.props;

    if (remoteIdentity && remoteIdentity.identity_id) {
      return this.renderActionsEdit();
    }

    return this.renderActionsCreate();
  }

  renderPhase() {
    switch (this.state.phase) {
      default:
      case 0:
        return this.renderFormPhase0();
      case 1:
        return this.renderFormPhase1();
      case 2:
        return this.renderFormPhase2();
      case 3:
        return this.renderFormPhase3();
    }
  }

  renderGlobalError() {
    if (!this.state.formErrors.global) {
      return null;
    }

    return (
      <FieldErrors errors={this.state.formErrors.global} />
    );
  }

  render() {
    return (
      <div className="m-remote-identity-email">
        {this.renderGlobalError()}
        {this.renderPhase()}
        {this.renderActions()}
      </div>
    );
  }
}

export default RemoteIdentityEmail;
