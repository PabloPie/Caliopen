import { createSelector } from 'reselect';
import { bindActionCreators, compose } from 'redux';
import { connect } from 'react-redux';
import { reduxForm, formValues } from 'redux-form';
import { withTranslator } from '@gandi/react-translate';
import { push } from 'react-router-redux';
import { matchPath } from 'react-router-dom';
import { requestContact, updateContact, createContact } from '../../store/modules/contact';
import { removeTab } from '../../store/modules/tab';
import fetchLocation from '../../services/api-location';
import { getNewContact } from '../../services/contact';
import Presenter from './presenter';

const contactIdSelector = (state, ownProps) => ownProps.match.params.contactId;
const contactSelector = state => state.contact;
const settingsSelector = state => state.settings.settings;
const currentTabSelector = createSelector(
  [state => state.tab.tabs, state => state.router.location && state.router.location.pathname],
  (tabs, pathname) => tabs.find(tab => matchPath(pathname, { path: tab.pathname, exact: true }))
);

const mapStateToProps = createSelector(
  [contactIdSelector, contactSelector, settingsSelector, currentTabSelector],
  (contactId, contactState, { contact_display_format }, currentTab) => ({
    contactId,
    contact: contactState.contactsById[contactId],
    form: `contact-${contactId || 'new'}`,
    // TODO: the following key fix this bug: https://github.com/erikras/redux-form/issues/2886#issuecomment-299426767
    key: `contact-${contactId || 'new'}`,
    initialValues: contactState.contactsById[contactId] || getNewContact(),
    isFetching: contactState.isFetching,
    contact_display_format,
    currentTab,
  })
);

const mapDispatchToProps = dispatch => ({
  ...bindActionCreators({
    requestContact,
    updateContact,
    createContact,
  }, dispatch),
  onSubmit: async (values, disp, props) => {
    // XXX: refactor me in middleware ?
    if (props.contactId) {
      await dispatch(updateContact({ contact: values, original: props.contact }));

      return dispatch(requestContact({ contactId: props.contact.contact_id }));
    }

    const resultAction = await dispatch(createContact({ contact: values }));
    const { location } = resultAction.payload.data;
    const { data: contact } = await fetchLocation(location);

    return Promise.all([
      dispatch(push(`/contacts/${contact.contact_id}`)),
      dispatch(removeTab(props.currentTab)),
    ]);
  },
});

export default compose(
  connect(mapStateToProps, mapDispatchToProps),
  reduxForm({
    destroyOnUnmount: false,
    enableReinitialize: true,
  }),
  formValues({ birthday: 'info.birthday' }),
  withTranslator()
)(Presenter);
