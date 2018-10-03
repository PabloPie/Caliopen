# -*- coding: utf-8 -*-
"""Caliopen user message qualification logic."""
from __future__ import absolute_import, print_function, unicode_literals
import json
import logging
from datetime import datetime
import pytz

from caliopen_main.message.parameters import (NewInboundMessage,
                                              Participant,
                                              Attachment)
from caliopen_main.message.parsers.twitter import TwitterDM

log = logging.getLogger(__name__)


class UserDMQualifier(object):
    """
    Process a Twitter direct message to unmarshal it in our stack
    """

    def __init__(self, user):
        self.user = user

    def get_participant(self, participant):
        """TODO: find related contact"""
        p = Participant()
        p.address = participant.address
        p.type = participant.type
        p.label = participant.label
        p.protocol = "twitter"
        return p

    def process_inbound(self, raw):
        """Process inbound message.

        @param raw: a RawMessage object
            which should be a json conforming to
            https://developer.twitter.com/en/docs/direct-messages/\
            sending-and-receiving/guides/message-create-object
        @rtype: NewMessage
        """

        message = TwitterDM(raw.raw_data)
        new_message = NewInboundMessage()
        new_message.raw_msg_id = raw.raw_msg_id
        new_message.body_plain = message.body_plain
        new_message.date = message.date
        new_message.type = message.type
        new_message.is_unread = True
        new_message.is_draft = False
        new_message.is_answered = False
        new_message.is_received = True
        new_message.importance_level = 0  # XXX tofix on parser

        participants = []
        for p in message.participants:
            participant = self.get_participant(p)
            new_message.participants.append(participant)
            participants.append(participant)
        if not participants:
            raise Exception("no participant found in raw message {}".format(
                raw.raw_msg_id))

        try:
            new_message.validate()
        except Exception as exc:
            log.error(
                "validation failed with error : « {} » \
                for new_message {}[dump : {}]".format(
                    exc, new_message, vars(new_message)))
            raise exc

        return new_message
