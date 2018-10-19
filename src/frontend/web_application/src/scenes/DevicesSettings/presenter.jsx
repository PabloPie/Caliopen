import React, { PureComponent } from 'react';
import PropTypes from 'prop-types';
import { Trans } from '@lingui/macro'; // eslint-disable-line import/no-extraneous-dependencies
import { PageTitle, Section } from '../../components/';
import DeviceSettings from './components/DeviceSettings';
import './style.scss';

class DevicesSettings extends PureComponent {
  static propTypes = {
    devices: PropTypes.arrayOf(PropTypes.shape({})),
    isCurrentDeviceVerified: PropTypes.bool,
  };

  static defaultProps = {
    devices: [],
    isCurrentDeviceVerified: undefined,
  };

  render() {
    const { devices, isCurrentDeviceVerified } = this.props;

    return (
      <div className="s-devices-settings">
        <PageTitle />
        {isCurrentDeviceVerified === false && (
          <div className="s-devices-settings__info">
            <Section>
              <div>
                <Trans id="devices.feedback.unverified_device">
                  It&apos;s the first time you attempt to connect to your Caliopen account on this
                  device.
                </Trans>
              </div>
              <div>
                <Trans id="devices.feedback.unverified_device_more">
                  To respect privacy and security rules, your discussions history will not fully
                  appear according to Privacy settings<br /> Please verify this device and
                  eventually set restrictions from your trusted device.
                </Trans>
              </div>
            </Section>
          </div>
        )}
        {devices.map(device => (
          <div key={device.device_id} className="s-devices-settings__device">
            <DeviceSettings device={device} />
          </div>
        ))}
      </div>
    );
  }
}

export default DevicesSettings;
