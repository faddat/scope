# NB only to be sourced

set -e

# these ought to match what is in Vagrantfile
# exported to override weave config.sh
export SSH_DIR="$PWD"
export HOSTS

: ${WEAVE_REPO:=github.com/weaveworks/weave}
: ${WEAVE_ROOT:="$(go list -e -f {{.Dir}} $WEAVE_REPO)"}

RUNNER="$WEAVE_ROOT/testing/runner/runner"
[ -x "$RUNNER" ] || (echo "Could not find weave test runner at $RUNNER." >&2 ; exit 1)

. "$WEAVE_ROOT/test/config.sh"

WEAVE="./weave"
SCOPE="../scope"

scope_on() {
	local host=$1
	shift 1
	[ -z "$DEBUG" ] || greyly echo "Scope on $host: $@" >&2
	DOCKER_HOST=tcp://$host:$DOCKER_PORT $SCOPE "$@"
}

weave_on() {
	local host=$1
	shift 1
	[ -z "$DEBUG" ] || greyly echo "Weave on $host: $@" >&2
	DOCKER_HOST=tcp://$host:$DOCKER_PORT $WEAVE "$@"
}

# this checks we have a named container
has_container() {
	local host=$1
	local name=$2
	local count=${3:-1}
	assert "curl -s http://$host:4040/api/topology/containers?system=show | jq -r '[.nodes[] | select(.label_major == \"$name\")] | length'" $count
}

has_container_id() {
	local host=$1
	local id=$2
	local count=${3:-1}
	assert "curl -s http://$host:4040/api/topology/containers?system=show | jq -r '[.nodes[] | select(.id == \"$id\")] | length'" $count
}

scope_end_suite() {
	end_suite
	for host in $HOSTS; do
		docker_on $host rm -f $(docker_on $host ps -a -q) 2>/dev/null 1>&2 || true
	done
}

container_id() {
	local host="$1"
	local name="$2"
	echo $(curl -s http://$host:4040/api/topology/containers?system=show | jq -r ".nodes[] | select(.label_major == \"$name\") | .id")
}

# this checks we have an edge from container 1 to container 2
has_connection() {
	local host="$1"
	local from_id="$2"
	local to_id="$3"
	local timeout="${4:-60}"

	for i in $(seq $timeout); do
		local containers="$(curl -s http://$host:4040/api/topology/containers?system=show)"
		local edge=$(echo "$containers" |  jq -r ".nodes[\"$from_id\"].adjacency | contains([\"$to_id\"])" 2>/dev/null)
		if [ "$edge" = "true" ]; then
			echo "Found edge $from -> $to after $i secs"
			assert "curl -s http://$host:4040/api/topology/containers?system=show |  jq -r '.nodes[\"$from_id\"].adjacency | contains([\"$to_id\"])'" true
			return
		fi
		sleep 1
	done

	echo "Failed to find edge $from -> $to after $timeout secs"
	assert "curl -s http://$host:4040/api/topology/containers?system=show |  jq -r '.nodes[\"$from_id\"].adjacency | contains([\"$to_id\"])'" true
}

wait_for_containers() {
	local host="$1"
	local timeout="$2"
	shift 2

	for i in $(seq $timeout); do
		local containers="$(curl -s http://$host:4040/api/topology/containers?system=show)"
		local found=0
		for name in "$@"; do
			local count=$(echo "$containers" | jq -r "[.nodes[] | select(.label_major == \"$name\")] | length")
			if [ -n "$count" ] && [ "$count" -ge 1 ]; then
				found=$(( found + 1 ))
			fi
		done

		if [ "$found" -eq $# ]; then
			echo "Found $found containers after $i secs"
			return
		fi

		sleep 1
	done

	echo "Failed to find containers $@ after $i secs"
}
