export const IDENTITY_TYPE_REMOTE = 'remote';
export const REMOTE_IDENTITY_STATUS_ACTIVE = 'active';
export const REMOTE_IDENTITY_STATUS_INACTIVE = 'inactive';
export const PROTOCOL_IMAP = 'imap';
export const PROTOCOL_GMAIL = 'gmail';
export const PROTOCOL_TWITTER = 'twitter';

/* eslint-disable camelcase */
export class Identity {
  constructor(props) {
    Object.assign(this, props);
  }

  display_name
  credentials: {}
  identifier
  infos: {}
  protocol
  status: REMOTE_IDENTITY_STATUS_INACTIVE
  type: IDENTITY_TYPE_REMOTE
}
