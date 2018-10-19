#!/bin/bash
set -ev

CURRENT_BRANCH=`git rev-parse --abbrev-ref HEAD`
TARGET_BRANCH=${TRAVIS_BRANCH:="master"}
PROJECT_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

if [[ "${CURRENT_BRANCH}" == "master" ]]; then
    # IF something happen on master, test everything
    BACKEND_CHANGE="yes"
    FRONTEND_CHANGE="yes"
else
    BACKEND_CHANGE=`(cd $PROJECT_DIRECTORY && git diff-tree --no-commit-id --name-only -r HEAD..$TARGET_BRANCH -- src/backend)`
    FRONTEND_CHANGE=`(cd $PROJECT_DIRECTORY && git diff-tree --no-commit-id --name-only -r HEAD..$TARGET_BRANCH -- src/frontend)`
fi

cd ${PROJECT_DIRECTORY}/devtools

function do_backend_tests {
    # Test build of backend Docker containers
    docker build -f ${PROJECT_DIRECTORY}/src/backend/Dockerfile.caliopen-go -t public-registry.caliopen.org/caliopen_go ../src/backend --no-cache
    docker build -f ${PROJECT_DIRECTORY}/src/backend/Dockerfile.caliopen-python -t public-registry.caliopen.org/caliopen_py ../src/backend --no-cache
    docker-compose build apiv2 lmtpd identity-poller imap-worker apiv1 cli mq-worker
}

function do_frontend_tests {
    # Test build of frontend Docker containers
    docker-compose build frontend
}


if [[ "x${BACKEND_CHANGE}" != "x" ]]; then
    echo "##### Doing backend tests"
    do_backend_tests
fi

if [[ "x${FRONTEND_CHANGE}" != "x" ]]; then
    echo "##### Doing frontend tests"
    do_frontend_tests
fi
