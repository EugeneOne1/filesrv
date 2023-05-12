#!/bin/sh

verbose="${VERBOSE:-0}"
readonly verbose

if [ "$verbose" -gt '0' ]
then
	set -x
fi

set -e -f -u

# bump_minor is an awk program that reads a minor release version, increments
# the minor part of it, and prints the next version.
#
# shellcheck disable=SC2016
bump_minor='/^v[0-9]+\.[0-9]+\.0$/ {
	print($1 "." $2 + 1 ".0");

	next;
}

{
	printf("invalid release version: \"%s\"\n", $0);

	exit 1;
}'
readonly bump_minor

# get_last_minor_zero returns the last new minor release.
get_last_minor_zero() {
	# List all tags.  Then, select those that fit the pattern of a new minor
	# release: a semver version with the patch part set to zero.
	#
	# Then, sort them first by the first field ("1"), starting with the
	# second character to skip the "v" prefix (".2"), and only spanning the
	# first field (",1").  The sort is numeric and reverse ("nr").
	#
	# Then, sort them by the second field ("2"), and only spanning the
	# second field (",2").  The sort is also numeric and reverse ("nr").
	#
	# Finally, get the top (that is, most recent) version.
	git tag\
		| grep -e 'v[0-9]\+\.[0-9]\+\.0$'\
		| sort -k 1.2,1nr -k 2,2nr -t '.'\
		| head -n 1
}

# last_tag is the most recent git tag.
last_tag="$( git describe --abbrev=0 )"
readonly last_tag

version="$last_tag"

# Finally, make sure that we don't output invalid versions.
if ! echo "$version" | grep -E -e '^v[0-9]+\.[0-9]+\.[0-9]+(-[ab]\.[0-9]+)?(\+[[:xdigit:]]+)?$' -q
then
	echo "generated an invalid version '$version'" 1>&2

	exit 1
fi

echo "$version"
