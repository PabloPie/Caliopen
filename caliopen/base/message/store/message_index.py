# -*- coding: utf-8 -*-
"""Caliopen message index classes."""
from __future__ import absolute_import, print_function, unicode_literals

import elasticsearch_dsl as dsl
from caliopen.base.store.model import BaseIndexDocument


class IndexedMessage(BaseIndexDocument):

    """Contact indexed message model."""

    doc_type = 'messages'

    message_id = dsl.String()
    thread_id = dsl.String()
    type = dsl.String()
    from_ = dsl.String()
    date = dsl.Date()
    date_insert = dsl.Date()
    privacy_index = dsl.Integer()
    importance_level = dsl.Integer()
    subject = dsl.String()
    external_message_id = dsl.String()
    external_parent_id = dsl.String()
    external_thread_id = dsl.String()
    tags = dsl.String()
    flags = dsl.String()
    offset = dsl.Integer()

    # XXX better nested definition
    headers = dsl.Nested()
    contacts = dsl.Nested()
