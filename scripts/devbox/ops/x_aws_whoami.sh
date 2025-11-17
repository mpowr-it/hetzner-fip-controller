#!/usr/bin/env bash
# --
# @version: 1.0.0
# @purpose: Shell script to identify yourself within gcloud api authentication scope
# --

if echo "" | aws projects list &> /dev/null
then
  C_DBX_INIT_AWS_ACCOUNT_CURRENT=$(aws config list account --format "value(core.account)")
  echo "Great! You are still logged in as $C_DBX_INIT_AWS_ACCOUNT_CURRENT"
else
  echo "Sorry, you're logged-out, please re-authenticate using 'devbox run login'"
fi
