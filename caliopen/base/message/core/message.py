# -*- coding: utf-8 -*-
"""Caliopen core user's messages."""
from __future__ import absolute_import, print_function, unicode_literals
import logging
import uuid

from datetime import datetime


from caliopen.base.core import BaseUserCore
from caliopen.base.parameters import ReturnCoreObject

from caliopen.base.message.store import Message as ModelMessage, RawMessage
from caliopen.base.message.parameters import Message as ParamMessage


log = logging.getLogger(__name__)


class Message(BaseUserCore):

    """Message core object."""

    _model_class = ModelMessage
    _pkey_name = 'message_id'

    @classmethod
    def create(cls, user, message, thread_id=None, lookup=None):
        """Create a new message for a given user."""
        message.validate()
        parent_id = message.external_parent_id
        message_id = uuid.uuid4()
        answer_to = lookup.message_id if lookup else None

        # TODO : index parts information
        extras = {'headers': message.headers,
                  'text': message.text,
                  'answer_to': answer_to,
                  'contacts': [contact.to_primitive()
                               for contact in message.recipients],
                  'date': message.date,
                  'size': message.size,
                  'from_': message.from_}
        attrs = {'message_id': message_id,
                 'thread_id': thread_id,
                 'type': message.type,
                 'from_': message.from_,
                 'date': message.date,
                 'date_insert': datetime.utcnow(),
                 'privacy_index': message.privacy_index,
                 'importance_level': message.importance_level,
                 'subject': message.subject,
                 'external_message_id': message.external_message_id,
                 'external_parent_id': parent_id,
                 'tags': message.tags,
                 'flags': ['Recent'],
                 'lookup': lookup,
                 '_indexed_extra': extras}
        return super(Message, cls).create(user, **attrs)

    @classmethod
    def by_thread_id(cls, user, thread_id, order=None, limit=None, offset=0):
        """Get messages for a given thread from index."""
        raise NotImplementedError()

    @property
    def text(self):
        """Return text from message."""
        # XXX do not use RawMessage lookup
        raw = RawMessage.get(self.user, str(self.external_message_id))
        msg = raw.parse()
        return msg.text


class ReturnMessage(ReturnCoreObject):

    """Return parameter from a core message."""

    _core_class = Message
    _return_class = ParamMessage
