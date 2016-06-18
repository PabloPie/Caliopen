# -*- coding: utf-8 -*-
"""Caliopen user message delivery logic."""
from __future__ import absolute_import, print_function, unicode_literals
import logging

from caliopen.base.exception import NotFound
from caliopen.base.user.core import ContactLookup

from caliopen.base.message.core import Message, RawMessage

from caliopen.base.message.core import (Thread,
                                        ThreadMessageLookup,
                                        ThreadRecipientLookup,
                                        ThreadExternalLookup)

from caliopen.base.message.parameters import Recipient
# XXX use a message formatter registry not directly mail format
from caliopen.base.message.format import MailMessage


log = logging.getLogger(__name__)


class UserMessageDelivery(object):

    """User message delivery processing."""

    _lookups = {
        'parent': ThreadMessageLookup,
        'list': ThreadRecipientLookup,
        'recipient': ThreadRecipientLookup,
        'external_thread': ThreadExternalLookup,
        'from': ThreadRecipientLookup,
    }

    def _get_recipients(self, user, msg):
        """Format recipients and try to resolve as contact."""
        recipients = []
        for type, recips in msg.recipients.iteritems():
            for addr, real_addr in recips:
                if addr != user.user_id:
                    recipient = Recipient()
                    recipient.address = real_addr
                    recipient.type = type
                    try:
                        log.debug('Try to resolve contact %s' % addr)
                        contact = ContactLookup.get(user, addr)
                        recipient.contact_id = str(contact.contact_id)
                    except NotFound:
                        pass

                    # XXX what to do if validation fail
                    recipient.validate()
                    recipients.append(recipient)
        return recipients

    def _get_tags(self, user, mail):
        """Evaluate user rules to get all tags for a mail."""
        tags = []
        for rule in user.rules:
            res, stop = rule.eval(mail)
            if res:
                tags.extend(res)
            if stop:
                break
        return tags

    def lookup(self, user, sequence):
        """Process lookup sequence to find thread to associate."""
        log.debug('Lookup sequence %r' % sequence)
        for prop in sequence:
            try:
                kls = self._lookups[prop[0]]
                log.debug('Will lookup %s with value %s' %
                          (prop[0], prop[1]))
                return kls.get(user, prop[1])
            except NotFound:
                log.debug('Lookup type %s with value %s failed' %
                          (prop[0], prop[1]))
        return None

    def __init_lookups(self, user, sequence, message):
        for prop in sequence:
            kls = self._lookups[prop[0]]
            params = {
                kls._pkey_name: prop[1],
                'thread_id': message.thread_id
            }
            if 'message_id' in kls._model_class._columns.keys():
                params.update({'message_id': message.message_id})
            lookup = kls.create(user, **params)
            log.debug('Create lookup %r' % lookup)
            if prop[0] == 'list':
                return

    def process(self, user, message_id):
        """Process a message for an user."""
        msg = RawMessage.get(user, message_id)
        log.debug('Retrieved raw message {} for message_id {}'.
                  format(msg.raw_id, message_id))

        # XXX should use raw message type to use correct message formatter
        mail = MailMessage(msg.data)

        message = mail.to_parameter()
        message.recipients = self._get_recipients(user, mail)
        addresses = [x.address for x in message.recipients]
        log.debug('Resolved recipients {}'.format(addresses))

        # compute tags
        message.tags = self._get_tags(user, mail)
        log.debug('Resolved tags {}'.format(message.tags))

        # lookup by external references
        lookup_sequence = mail.lookup_sequence()
        lookup = self.lookup(user, lookup_sequence)

        # XXX missing thread management
        msg = Message.create(user, message, thread_id=None, lookup=lookup)
        # XXX Init lookup
        if not lookup:
            self.__init_lookups(user, lookup_sequence, msg)
        else:
            if msg.external_message_id:
                params = {
                    'external_message_id': msg.external_message_id,
                    'thread_id': msg.thread_id,
                    'message_id': msg.message_id,
                }
                new_lookup = ThreadMessageLookup.create(user, **params)
                log.debug('Created message lookup %r' % new_lookup)
        return msg
