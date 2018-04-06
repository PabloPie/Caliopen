#!/usr/bin/env python

"""
Command Line Interface (CLI) for caliopen project


"""
import sys
import argparse
import logging

from caliopen_storage.config import Configuration
from caliopen_storage.helpers.connection import connect_storage
from caliopen_cli.commands import (shell, import_email, set_notifications_ttls,
                                   setup_storage, create_user,
                                   import_vcard, dump_model, dump_indexes,
                                   inject_email, basic_compute, migrate_index,
                                   import_reserved_names, import_given_email)

logging.basicConfig(level=logging.INFO)


def main(args=sys.argv):
    parser = argparse.ArgumentParser()
    parser.add_argument('-f', dest='conffile', default='development.ini')
    subparsers = parser.add_subparsers(title="action")

    sp_import = subparsers.add_parser('import', help='import existing mailbox')
    sp_import.set_defaults(func=import_email)
    sp_import.add_argument('-f', dest='format', choices=['mbox', 'maildir'],
                           default='mbox')
    sp_import.add_argument('-p', dest='import_path')
    sp_import.add_argument('-e', dest='email')
    sp_import.add_argument('--contact-probability', dest='contact_probability',
                           default=1.0)

    sp_given = subparsers.add_parser('import_given',
                                     help='import existing mailbox')
    sp_given.set_defaults(func=import_given_email)
    sp_given.add_argument('-f', dest='format', choices=['mbox', 'maildir'],
                          default='mbox')
    sp_given.add_argument('-p', dest='import_path')
    sp_given.add_argument('-e', dest='email')
    sp_given.add_argument('--contact-probability', dest='contact_probability',
                          default=1.0)

    sp_import_vcard = subparsers.add_parser('import_vcard',
                                            help='import vcard')
    sp_import_vcard.set_defaults(func=import_vcard)
    sp_import_vcard.add_argument('-u', dest='username', help='username')
    sp_import_vcard.add_argument('-d', dest='directory', help='directory')
    sp_import_vcard.add_argument('-f', dest='file_vcard', help='file')

    sp_setup_storage = subparsers.add_parser('setup',
                                             help='initialize the storage engine')
    sp_setup_storage.set_defaults(func=setup_storage)

    sp_create_user = subparsers.add_parser('create_user',
                                           help='Create a new user')
    sp_create_user.set_defaults(func=create_user)
    sp_create_user.add_argument('-e', dest='email', help='user email')
    sp_create_user.add_argument('-p', dest='password', help='password')
    sp_create_user.add_argument('-g', dest='given_name',
                                help='user given name')
    sp_create_user.add_argument('-f', dest='family_name',
                                help='user family name')

    sp_shell = subparsers.add_parser('shell')
    sp_shell.set_defaults(func=shell)

    sp_set_ttls = subparsers.add_parser('set_notifs_ttls',
                                        help='fillup table notification_ttl'
                                             'with a set of default values')
    sp_set_ttls.set_defaults(func=set_notifications_ttls)

    sp_dump = subparsers.add_parser('dump')
    sp_dump.set_defaults(func=dump_model)
    sp_dump.add_argument('-m', dest='model', help='model to dump')
    sp_dump.add_argument('-o', dest='output_path', help='output path')

    sp_dump_index = subparsers.add_parser('dump_index')
    sp_dump_index.set_defaults(func=dump_indexes)
    sp_dump_index.add_argument('-o', dest='output_path', help='output path')

    sp_migrate_index = subparsers.add_parser('migrate_index')
    sp_migrate_index.set_defaults(func=migrate_index)
    sp_migrate_index.add_argument('-s', dest='input_script',
                                  help='python script to execute against index')

    sp_inject = subparsers.add_parser('inject')
    sp_inject.set_defaults(func=inject_email)
    sp_inject.add_argument('-f', dest='format', choices=['mbox', 'maildir'],
                           default='mbox')
    sp_inject.add_argument('-p', dest='import_path')
    sp_inject.add_argument('-e', dest='email')
    sp_inject.add_argument('--host', dest='host', help='host to send mail to',
                           default='localhost:25')

    sp_compute = subparsers.add_parser('compute', help='Launch basic compute')
    sp_compute.set_defaults(func=basic_compute)
    sp_compute.add_argument('-u', dest='username', help='username')
    sp_compute.add_argument('-j', dest='job', help='job name')

    sp_reserved = subparsers.add_parser('reserved_names',
                                        help='Import reserved names list')
    sp_reserved.set_defaults(func=import_reserved_names)
    sp_reserved.add_argument('-i', dest='input_file', help='csv file')

    kwargs = parser.parse_args(args[1:])
    kwargs = vars(kwargs)

    config_uri = kwargs.pop('conffile')
    func = kwargs.pop('func')

    Configuration.load(config_uri, 'global')
    connect_storage()

    func(**kwargs)


if __name__ == '__main__':
    main()
