import { Component } from 'react';
import PropTypes from 'prop-types';
import {
  TIMELINE_FILTER_ALL, TIMELINE_FILTER_RECEIVED, TIMELINE_FILTER_SENT, TIMELINE_FILTER_DRAFT,
} from '../../../../store/modules/message';

class TimelineFilter extends Component {
  static propTypes = {
    render: PropTypes.func.isRequired,
    filter: PropTypes.func.isRequired,
    currentFilter: PropTypes.string.isRequired,
    i18n: PropTypes.shape({}).isRequired,
  };
  static defaultProps = {
  };
  state = {};

  componentWillMount() {
    const { i18n } = this.props;

    this.translations = {
      [TIMELINE_FILTER_ALL]: i18n.t`timeline-filter.options.all`,
      [TIMELINE_FILTER_RECEIVED]: i18n.t`timeline-filter.options.received`,
      [TIMELINE_FILTER_SENT]: i18n.t`timeline-filter.options.sent`,
      [TIMELINE_FILTER_DRAFT]: i18n.t`timeline-filter.options.draft`,
    };
  }

  componentWillReceiveProps() {

  }

  handleSelect = type => this.props.filter(type);

  render() {
    const { render, currentFilter } = this.props;

    const options = [
      TIMELINE_FILTER_ALL,
      TIMELINE_FILTER_RECEIVED,
      TIMELINE_FILTER_SENT,
      TIMELINE_FILTER_DRAFT,
    ].map(filter => (
      {
        label: this.translations[filter],
        value: filter,
        select: () => this.handleSelect(filter),
      }
    ));

    return render(options, currentFilter);
  }
}

export default TimelineFilter;
