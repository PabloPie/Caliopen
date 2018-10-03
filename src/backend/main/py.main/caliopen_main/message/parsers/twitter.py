# -*- coding: utf-8 -*-

import logging
import base64
import json
from itertools import groupby
from mailbox import Message
from email.header import decode_header
from datetime import datetime
import pytz
import hashlib

from email.utils import parsedate_tz, mktime_tz, getaddresses

import zope.interface

from caliopen_main.common.helpers.normalize import clean_email_address
from caliopen_main.common.helpers.strings import to_utf8
from caliopen_main.common.interfaces import (IAttachmentParser, IMessageParser,
                                             IParticipantParser)

log = logging.getLogger(__name__)


class TwitterDM(object):
    """
    Twitter direct message structure
    """

    zope.interface.implements(IMessageParser)

    recipient_headers = ['From', 'To']
    message_type = 'DM twitter'
    warnings = []
    body_html = ""
    body_plain = ""

    def __init__(self, raw_data):
        self.raw = raw_data
        dm = json.loads(self.raw)
        log.info(dm)  #  TODO: remove
        self.recipient_name = dm["message_create"]["target"][
            "recipient_screen_name"]
        self.sender_name = dm["message_create"]["sender_screen_name"]
        self.body_plain = dm["message_create"]["message_data"]["text"]
        self.date = datetime.fromtimestamp(float(dm["created_timestamp"][0:10]),
                                           tz=pytz.utc)
        # self.size TODO: what to put here ?
        #            size of text message or size of json returned by twitter ?
        self.type = self.message_type
        self.is_unread = True  # TODO: handle DM sent by user
        #                                     if broker keeps them when fetching
        self.is_draft = False
        self.is_answered = False
        self.is_received = True  # TODO: handle DM sent by user
        #                                     if broker keeps them when fetching
        self.importance_level = 0

    @property
    def participants(self):
        "one sender only for now"
        return [TwitterParticipant("To", self.recipient_name),
                TwitterParticipant("From", self.sender_name)]


class TwitterParticipant(object):
    """
    Twitter sender and recipient parser
    """

    zope.interface.implements(IParticipantParser)

    def __init__(self, type, screen_name):
        """Parse an email address and create a participant."""
        self.type = type
        self.address = screen_name
        self.label = screen_name
