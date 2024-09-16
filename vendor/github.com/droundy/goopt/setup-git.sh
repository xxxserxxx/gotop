#!/bin/sh

# This script is to be run in the root of the repository, and will set
# up git hooks to do reasonable things.

set -ev

test -d git
test -d .git/hooks

for i in `ls git | grep -v '~'`; do
    echo Setting up $i hook...
    ln -sf ../../git/$i .git/hooks/$i
done
