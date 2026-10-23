#!/usr/bin/env sh

# This script is intended to be used as a Docker HEALTHCHECK for the GitLab container.
# It prepares GitLab prior to running acceptance tests.
#
# This is a known workaround for docker-compose lacking lifecycle hooks.
# See: https://github.com/docker/compose/issues/1809#issuecomment-657815188

set -e

# Check for a successful HTTP status code from GitLab.
curl --silent --show-error --fail --output /dev/null 127.0.0.1:80

# Because this script runs on a regular health check interval,
# this file functions as a marker that tells us if initialization already finished.
done=/var/gitlab-acctest-initialized

test -f $done || {
  echo 'Initializing GitLab for acceptance tests'

  echo 'Creating access token'
  (
    cat <<EOF
client_go = PersonalAccessToken.create(
  user_id: 1,
  scopes: [:api, :read_user],
  name: :terraform,
  expires_at: Time.now + 30.days);
client_go.set_token('$GITLAB_TOKEN');
client_go.save!;
EOF
  ) | gitlab-rails console


  echo 'Enabling `retain_resource_access_token_user_after_revoke` feature flag'
  (
    cat <<EOF
Feature.enable(:retain_resource_access_token_user_after_revoke);
EOF
  ) | gitlab-rails console

  touch $done
}

echo 'GitLab is ready for acceptance tests'

