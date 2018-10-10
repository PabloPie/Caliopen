// TODO: refactor with Gmail: RemoteIdentityWithOAuth
import React, { Component, Fragment } from 'react';
import PropTypes from 'prop-types';
import classnames from 'classnames';
import { Trans, withI18n } from 'lingui-react';
import { Confirm, Button, Callout, FormGrid, FormRow, FormColumn, TextFieldGroup, Spinner } from '../../../../components';
import { REMOTE_IDENTITY_STATUS_ACTIVE, REMOTE_IDENTITY_STATUS_INACTIVE, Identity, withAuthorizePopup } from '../../../../modules/remoteIdentity';
import Phase0 from '../Phase0';

@withI18n()
@withAuthorizePopup()
class RemoteIdentityTwitter extends Component {
  static propTypes = {
    i18n: PropTypes.shape({}).isRequired,
    className: PropTypes.string,
    remoteIdentity: PropTypes.shape({}).isRequired,
    onChange: PropTypes.func.isRequired,
    onDelete: PropTypes.func.isRequired,
    onCancel: PropTypes.func.isRequired,
    initPopup: PropTypes.func.isRequired,
    authorizePopup: PropTypes.func.isRequired,
  };
  static defaultProps = {
    className: undefined,
  };
  state = {
    hasActivity: false,
    remoteIdentity: {
      identifier: '',
      active: true,
    },
    formErrors: {},
  };

  componentWillMount() {
    this.setState(this.getStateFromProps());
  }

  componentDidMount() {

  }

  componentWillReceiveProps() {

  }

  getStateFromProps = () => {
    const {
      remoteIdentity: {
        identity_id: identityId,
        identifier,
        status,
      },
    } = this.props;
    const active = status ?
      status === REMOTE_IDENTITY_STATUS_ACTIVE :
      this.state.remoteIdentity.active;

    return {
      phase: identityId ? 0 : 1,
      remoteIdentity: {
        ...this.state.remoteIdentity,
        identifier,
        active,
      },
    };
  }

  getIdentityFromState = () => {
    const {
      remoteIdentity: {
        identifier,
        active,
      },
    } = this.state;
    const { remoteIdentity } = this.props;

    return new Identity({
      ...remoteIdentity,
      identifier,
      display_name: identifier,
      status: active ? REMOTE_IDENTITY_STATUS_ACTIVE : REMOTE_IDENTITY_STATUS_INACTIVE,
    });
  }

  handleDelete = () => {
    const { remoteIdentity, onDelete } = this.props;

    onDelete({ identity: remoteIdentity });
  }

  handleEdit = () => {
    this.setState({ phase: 1 });
  }

  handleCancel = () => {
    const { onCancel } = this.props;
    this.setState(this.getStateFromProps(), () => {
      onCancel();
    });
  }

  authorize = async () => {
    if (!this.validate()) {
      return;
    }

    const { authorizePopup, onChange, initPopup } = this.props;
    const identity = this.getIdentityFromState();

    initPopup({ identity });

    this.setState({
      hasActivity: true,
    });

    const identityUpToDate = await onChange({ identity });

    // do not await anything, this component has been unmounted if the identity just been created
    authorizePopup({ identity: identityUpToDate });
  }

  handleActivate = (active) => {
    this.setState(prevState => ({
      remoteIdentity: {
        ...prevState.remoteIdentity,
        active,
      },
    }));

    return this.authorize();
  }

  validate = () => {
    let isValid = true;

    if (
      !this.state.remoteIdentity.identifier || this.state.remoteIdentity.identifier.length === 0
    ) {
      isValid = false;
      this.setState({
        formErrors: {
          identifier: [(
            <Trans id="remote_identity.form.gmail.identifier.error">a twitter name is required</Trans>
          )],
        },
      });
    } else {
      this.setState({ formErrors: { identifier: [] } });
    }

    return isValid;
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

  renderActionsCreate() {
    return (
      <div>
        <Button onClick={this.handleCancel} shape="plain" className="m-remote-identity-email__action">
          <Trans id="remote_identity.action.cancel">Cancel</Trans>
        </Button>
        <Button
          onClick={this.authorize}
          shape="plain"
          className="m-remote-identity-email__action"
          icon={this.state.hasActivity ? (<Spinner isLoading display="inline" />) : 'twitter'}
          disabled={this.state.hasActivity}
        >
          <Trans id="remote_identity.action.authorize">Authorize</Trans>
        </Button>
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
            <Trans id="remote_identity.action.cancel">Edit</Trans>
          </Button>
        )}
        {this.state.phase >= 0 && (
          <Fragment>
            <Button onClick={this.handleCancel} shape="hollow" className="m-remote-identity-email__action">
              <Trans id="remote_identity.action.cancel">Cancel</Trans>
            </Button>
            <Button
              onClick={this.authorize}
              shape="plain"
              className="m-remote-identity-email__action"
              icon={this.state.hasActivity ? (<Spinner isloading display="inline" />) : 'twitter'}
              disabled={this.state.hasActivity}
            >
              <Trans id="remote_identity.action.authorize">Authorize</Trans>
            </Button>
          </Fragment>
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

  renderFormPhase1 = () => {
    const { i18n, remoteIdentity } = this.props;

    return (
      <Fragment>
        <Callout color="info">
          <Trans id="remote_identity.twitter_how_to">
            You will be asked to signin to twitter and authorize Caliopen to access to your direct
            messages. Afterward you will be able to see your last messages (until 30 days) and send
            direct messages to your contacts.
          </Trans>
        </Callout>
        <FormGrid>
          <FormRow>
            <FormColumn bottomSpace>
              <TextFieldGroup
                label={<Trans id="remote_identity.gmail.form.identifier.label">Email:</Trans>}
                placeholder={i18n._('remote_identity.gmail.form.identifier.placeholder', { defaults: 'johndoe@gmail.com' })}
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
          </FormRow>
        </FormGrid>
      </Fragment>
    );
  }

  renderPhase() {
    switch (this.state.phase) {
      default:
      case 0:
        return this.renderFormPhase0();
      case 1:
        return this.renderFormPhase1();
    }
  }

  render() {
    const { className } = this.props;

    return (
      <div className={classnames(className)}>
        {this.renderPhase()}
        {this.renderActions()}
      </div>
    );
  }
}

export default RemoteIdentityTwitter;
