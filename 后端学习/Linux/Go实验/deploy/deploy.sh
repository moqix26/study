#!/usr/bin/env bash
set -Eeuo pipefail

APP_NAME="${APP_NAME:-shortlink}"
ARTIFACT="${1:-}"
VERSION="${2:-}"
ROOT="${ROOT:-/opt/${APP_NAME}}"
RELEASES="${ROOT}/releases"
CURRENT="${ROOT}/current"
SERVICE="${SERVICE:-${APP_NAME}.service}"
HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:8080/healthz}"
LOCK_FILE="${LOCK_FILE:-${ROOT}/.deploy.lock}"

log() {
	printf '[%s] %s\n' "$(date '+%F %T')" "$*"
}

fail() {
	log "ERROR: $*"
	exit 1
}

[[ -n "$ARTIFACT" ]] || fail "usage: $0 /path/to/shortlink-api VERSION"
[[ -f "$ARTIFACT" ]] || fail "artifact does not exist: $ARTIFACT"
[[ -n "$VERSION" ]] || fail "VERSION must not be empty"
[[ "$VERSION" =~ ^[0-9A-Za-z._-]+$ ]] || fail "VERSION contains unsafe characters"

mkdir -p "$RELEASES"
ROOT_REAL="$(realpath "$ROOT")"
RELEASES_REAL="$(realpath "$RELEASES")"
[[ "$RELEASES_REAL" == "$ROOT_REAL/releases" ]] || fail "unexpected releases path: $RELEASES_REAL"

exec 9>"$LOCK_FILE"
flock -n 9 || fail "another deployment is running"

RELEASE_DIR="${RELEASES}/${VERSION}"
[[ ! -e "$RELEASE_DIR" ]] || fail "release already exists: $RELEASE_DIR"

PREVIOUS=""
if [[ -L "$CURRENT" ]]; then
	PREVIOUS="$(readlink -f "$CURRENT")"
fi

rollback() {
	if [[ -n "$PREVIOUS" && -d "$PREVIOUS" ]]; then
		log "rolling back to $PREVIOUS"
		ln -sfn "$PREVIOUS" "${CURRENT}.rollback"
		mv -Tf "${CURRENT}.rollback" "$CURRENT"
		systemctl restart "$SERVICE" || true
	else
		log "no previous release is available for automatic rollback"
	fi
}

log "installing release $VERSION"
install -d -m 0755 "$RELEASE_DIR"
install -m 0755 "$ARTIFACT" "${RELEASE_DIR}/shortlink-api"
sha256sum "${RELEASE_DIR}/shortlink-api" >"${RELEASE_DIR}/SHA256SUMS"

ln -sfn "$RELEASE_DIR" "${CURRENT}.new"
mv -Tf "${CURRENT}.new" "$CURRENT"

if ! systemctl restart "$SERVICE"; then
	rollback
	fail "systemd could not restart $SERVICE"
fi

for attempt in $(seq 1 20); do
	if curl --fail --silent --show-error --max-time 2 "$HEALTH_URL" >/dev/null; then
		log "deployment succeeded: version=$VERSION attempt=$attempt"
		exit 0
	fi
	sleep 1
done

rollback
fail "health check failed after 20 attempts"
