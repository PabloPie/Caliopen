# -*- coding: utf-8 -*-
"""
Caliopen mail message format management.

mail parsing is included in python, so this is not
getting external dependencies.

For formats with needs of external packages, they
must be defined outside of this one.
"""

import logging
import base64
from itertools import groupby
from mailbox import Message
from email.header import decode_header
import datetime
import pytz
import hashlib

from email.utils import parsedate_tz, mktime_tz, getaddresses

import zope.interface

from caliopen_main.common.helpers.normalize import clean_email_address
from caliopen_main.common.helpers.strings import to_utf8
from caliopen_main.common.interfaces import (IAttachmentParser, IMessageParser,
                                             IParticipantParser)

log = logging.getLogger(__name__)


class MailAttachment(object):
    """Mail part structure."""

    zope.interface.implements(IAttachmentParser)

    def __init__(self, part):
        """Extract attachment attributes from a mail part."""
        self.content_type = part.get_content_type()
        self.filename = part.get_filename()
        content_disposition = part.get("Content-Disposition")
        if content_disposition:
            dispositions = content_disposition.strip().split(";")
            self.is_inline = bool(dispositions[0].lower() == "inline")
        else:
            self.is_inline = True
        data = part.get_payload()
        self.size = len(data) if data else 0
        self.can_index = False
        if 'text' in part.get_content_type():
            self.can_index = True
            charsets = part.get_charsets()
            if len(charsets) > 1:
                raise Exception('Too many charset %r for %s' %
                                (charsets, part.get_payload()))
            self.charset = charsets[0]
            if 'Content-Transfer-Encoding' in part.keys():
                if part.get('Content-Transfer-Encoding') == 'base64':
                    data = base64.b64decode(data)
            if self.charset:
                data = data.decode(self.charset, 'replace'). \
                    encode('utf-8')
        boundary = part.get("Mime-Boundary", failobj="")
        if boundary is not "":
            self.mime_boundary = boundary
        else:
            self.mime_boundary = ""
        self.data = data

    @classmethod
    def is_attachment(cls, part):
        """Check if a part conform to Caliopen's attachment definition.

        A part is an "attachment" if it verifies ANY of this conditions :
        - it has a Content-Disposition header with param "attachment"
        - the main part of the Content-Type header
                                        is within "attachment_types" list below

        see https://www.iana.org/assignments/media-types/media-types.xhtml

        :param part: an email/message's part as return by the walk() func.
        :return: true or false
        """
        content_disposition = part.get("Content-Disposition")
        if content_disposition:
            dispositions = content_disposition.strip().split(";")
            if bool(dispositions[0].lower() == "attachment") or \
                    bool(dispositions[0].lower() == "inline"):
                return True

        attachment_types = (
            "application", "image", "video", "audio", "message", "font")
        if part.get_content_maintype() in attachment_types:
            return True
        return False


class MailParticipant(object):
    """Mail participant parser."""

    zope.interface.implements(IParticipantParser)

    def __init__(self, type, addr):
        """Parse an email address and create a participant."""
        self.type = type
        parts = clean_email_address(addr)
        self.address = parts[0]
        self.label = parts[1]


class MailMessage(object):
    """
    Mail message structure.

    Got a mail in raw rfc2822 format, parse it to
    resolve all recipients emails, parts and group headers
    """

    zope.interface.implements(IMessageParser)

    recipient_headers = ['From', 'To', 'Cc', 'Bcc']
    message_type = 'email'
    warnings = []
    body_html = ""
    body_plain = ""

    def __init__(self, raw_data):
        """Parse an RFC2822,5322 mail message."""
        self.raw = raw_data
        try:
            self.mail = Message(raw_data)
        except Exception as exc:
            log.error('Parse message failed %s' % exc)
            raise exc
        if self.mail.defects:
            # XXX what to do ?
            log.warn('Defects on parsed mail %r' % self.mail.defects)
            self.warning = self.mail.defects
        self.get_bodies()

    def get_bodies(self):
        """Extract body alternatives, if any."""
        body_html = ""
        body_plain = ""

        if self.mail.get("Content-Type", None):
            if self.mail.is_multipart():
                for top_level_part in self.mail.get_payload():
                    if top_level_part.get_content_maintype() == "multipart":
                        for alternative in top_level_part.get_payload():
                            charset = alternative.get_param("charset")
                            if isinstance(charset, tuple):
                                charset = unicode(charset[2],
                                                  charset[0] or "us-ascii")
                            if alternative.get_content_type() == "text/plain":
                                body_plain = alternative.get_payload(
                                    decode=True)
                                self.body_plain = to_utf8(body_plain, charset)
                            elif alternative.get_content_type() == "text/html":
                                body_html = alternative. \
                                    get_payload(decode=True)
                                self.body_html = to_utf8(body_html, charset)
                        break
                    else:
                        charset = top_level_part.get_param("charset")
                        if isinstance(charset, tuple):
                            charset = unicode(charset[2],
                                              charset[0] or "us-ascii")
                        if top_level_part.get_content_type() == "text/plain":
                            body_plain = top_level_part. \
                                get_payload(decode=True)
                            self.body_plain = to_utf8(body_plain, charset)
                        elif top_level_part.get_content_type() == "text/html":
                            body_html = top_level_part.get_payload(decode=True)
                            self.body_html = to_utf8(body_html, charset)
            else:
                charset = self.mail.get_param("charset")
                if isinstance(charset, tuple):
                    charset = unicode(charset[2], charset[0] or "us-ascii")
                if self.mail.get_content_type() == "text/html":
                    body_html = self.mail.get_payload(decode=True)
                    self.body_html = to_utf8(body_html, charset)
                else:
                    body_plain = self.mail.get_payload(decode=True)
                    self.body_plain = to_utf8(body_plain, charset)
        else:
            self.body_plain = self.mail.get_payload(decode=True)

    @property
    def subject(self):
        """Mail subject."""
        s = decode_header(self.mail.get('Subject'))
        charset = s[0][1]
        if charset is not None:
            return s[0][0].decode(charset, "replace"). \
                encode("utf-8", "replace")
        else:
            return s[0][0]

    @property
    def size(self):
        """Get mail size in bytes."""
        return len(self.mail.as_string())

    @property
    def external_references(self):
        """Return mail references to be used as external references.

         making use of RFC5322 headers :
            message-id
            in-reply-to
            references
        headers' strings are pruned to extract email addresses only.
        """
        ext_id = self.mail.get('Message-Id')
        parent_id = self.mail.get('In-Reply-To')
        ref = self.mail.get_all("References")
        ref_addr = getaddresses(ref) if ref else None
        ref_ids = [address[1] for address in ref_addr] if ref_addr else []
        mid = clean_email_address(ext_id)[1] if ext_id else None
        pid = clean_email_address(parent_id)[1] if parent_id else None
        return {
            'message_id': mid,
            'parent_id': pid,
            'ancestors_ids': ref_ids}

    @property
    def date(self):
        """Get UTC date from a mail message."""
        mail_date = self.mail.get('Date')
        if mail_date:
            tmp_date = parsedate_tz(mail_date)
            return datetime.datetime.fromtimestamp(mktime_tz(tmp_date))
        log.debug('No date on mail using now (UTC)')
        return datetime.datetime.now(tz=pytz.utc)

    @property
    def participants(self):
        """Mail participants."""
        participants = []
        for header in self.recipient_headers:
            addrs = []
            participant_type = header.capitalize()
            if self.mail.get(header):
                if ',' in self.mail.get(header):
                    parts = self.mail.get(header).split(',')
                    filtered = [x for x in parts if '@' in x]
                    addrs.extend(filtered)
                else:
                    addrs.append(self.mail.get(header))
            for addr in addrs:
                participant = MailParticipant(participant_type, addr)
                participants.append(participant)
        return participants

    @property
    def hash_participants(self):
        """Create an hash from participants addresses for global lookup."""
        addresses = [x.address for x in self.participants]
        addresses = list(set(addresses))
        addresses.sort()
        return hashlib.sha256(''.join(addresses)).hexdigest()

    @property
    def attachments(self):
        """Extract parts which we consider as attachments."""
        if not self.mail.is_multipart():
            return []
        attchs = []
        for p in walk_with_boundary(self.mail, ""):
            if not p.is_multipart():
                if MailAttachment.is_attachment(p):
                    attchs.append(MailAttachment(p))
        return attchs

    @property
    def extra_parameters(self):
        """Mail message extra parameters."""
        lists = self.mail.get_all("List-ID")
        lists_addr = getaddresses(lists) if lists else None
        lists_ids = [address[1] for address in lists_addr] \
            if lists_addr else []
        return {'lists': lists_ids}

    def lookup_discussion_sequence(self, *args, **kwargs):
        """Return list of lookup type, value from a mail message."""
        seq = []

        # list lookup first
        for list_id in self.extra_parameters.get('lists', []):
            seq.append(('list', list_id))

        seq.append(('global', self.hash_participants))

        # try to link message to external thread's root message-id
        if len(self.external_references["ancestors_ids"]) > 0:
            seq.append(("thread",
                        self.external_references["ancestors_ids"][0]))
        elif self.external_references["parent_id"]:
            seq.append(("thread", self.external_references["parent_id"]))
        elif self.external_references["message_id"]:
            seq.append(("thread", self.external_references["message_id"]))

        return seq

    # Others parameters specific for mail message

    @property
    def headers(self):
        """Extract all headers into list.

        Duplicate on headers exists, group them by name
        with a related list of values
        """
        def keyfunc(item):
            return item[0]

        # Group multiple value for same headers into a dict of list
        headers = {}
        data = sorted(self.mail.items(), key=keyfunc)
        for k, g in groupby(data, key=keyfunc):
            headers[k] = [x[1] for x in g]
        return headers


def walk_with_boundary(message, boundary):
    """Recurse in boundaries."""
    message.add_header("Mime-Boundary", boundary)
    yield message
    if message.is_multipart():
        subboundary = message.get_boundary("")
        for subpart in message.get_payload():
            for subsubpart in walk_with_boundary(subpart, subboundary):
                yield subsubpart
