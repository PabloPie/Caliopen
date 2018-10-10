import { createSelector } from 'reselect';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import { createIdentity, deleteIdentity, updateIdentity, identitiesSelector, identityStateSelector } from '../../modules/remoteIdentity';
import { requestRemoteIdentities } from '../../store/modules/remote-identity';
import Presenter from './presenter';

const mapStateToProps = createSelector(
  [identityStateSelector, identitiesSelector],
  (identityState, identities) => ({
    isFetching: identityState.isFetching,
    identities: [...identities]
      .sort((a, b) => a.display_name.localeCompare(b.display_name)),
  })
);

const onIdentityChange = ({ identity }) => (dispatch) => {
  if (!identity.identity_id) {
    return dispatch(createIdentity({ identity }));
  }

  return dispatch(updateIdentity({ identity }));
};

const mapDispatchToProps = dispatch => bindActionCreators({
  requestRemoteIdentities,
  onIdentityChange,
  onIdentityDelete: deleteIdentity,
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(Presenter);
