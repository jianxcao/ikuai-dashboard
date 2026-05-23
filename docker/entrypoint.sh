#!/bin/sh
set -eu

PUID="${PUID:-1000}"
PGID="${PGID:-1000}"
UMASK="${UMASK:-022}"
IKUAI_CONFIG="${IKUAI_CONFIG:-/config/config.yaml}"

case "${PUID}" in
	''|*[!0-9]*)
		echo "PUID must be a numeric user id, got: ${PUID}" >&2
		exit 1
		;;
esac

case "${PGID}" in
	''|*[!0-9]*)
		echo "PGID must be a numeric group id, got: ${PGID}" >&2
		exit 1
		;;
esac

case "${UMASK}" in
	[0-7][0-7][0-7]|[0-7][0-7][0-7][0-7])
		;;
	*)
		echo "UMASK must be an octal value like 022 or 0022, got: ${UMASK}" >&2
		exit 1
		;;
esac

mkdir -p /config "$(dirname "${IKUAI_CONFIG}")"

if [ ! -f "${IKUAI_CONFIG}" ]; then
	cp /defaults/config.example.yaml "${IKUAI_CONFIG}"
fi

groupmod -o -g "${PGID}" ikuai
usermod -o -u "${PUID}" ikuai

chown -R ikuai:ikuai /app /config
umask "${UMASK}"

if [ "$#" -eq 0 ]; then
	set -- /app/ikuai-dashboard
fi

if [ "$1" = "/app/ikuai-dashboard" ] || [ "$1" = "ikuai-dashboard" ]; then
	set -- "$@" -config "${IKUAI_CONFIG}"
fi

exec dumb-init gosu ikuai:ikuai "$@"
