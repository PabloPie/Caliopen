#!/bin/bash
set -e

CURRENT_DIR=$(pwd)
SRC_CHANGED="false"
BRANCHES=""

# Images based on alpine do not contain git
if [ -f "/etc/alpine-release" ];
then
	apk add --no-cache git >/dev/null 2>&1
fi

if [ "$DRONE_BUILD_EVENT" = "pull_request" ];
then
	git fetch origin --no-tags $DRONE_BRANCH:$DRONE_BRANCH
	BRANCHES="$DRONE_BRANCH...HEAD"
elif [ $(git show --no-patch --format="%P" $DRONE_COMMIT | awk '{print NF}') = 2 ];
then
	# Two parents means merge commit
	BRANCHES="HEAD HEAD^"
else
	# Any way to check changes from a direct push with multiple possible commits?
	echo "Not a PR or a Merge, assuming change to every service"
	SRC_CHANGED="true"
	return
fi

####

# If DEPS is defined it means we are checking for changes in dependencies
# otherwise we are just checking if it's front or back
if [ -n "$DEPS" ];
then
	cd $BASE_DIR
	# Python programs don't have their own directory as a dependency
	if [ "$LANG" != "js" ]; then
		DEPS="$DEPS $PROG"
	fi
else
	DEPS=$BASE_DIR
fi

echo "Checking changes to $DEPS between $BRANCHES"

if ! git --no-pager diff --quiet --exit-code $BRANCHES -- $DEPS;
then
	echo "Changes found"
	SRC_CHANGED="true"
fi

cd $CURRENT_DIR
