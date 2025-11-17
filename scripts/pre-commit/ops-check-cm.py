#!/usr/bin/env python3

import sys
import re

# Allowed prefix/suffix stack
ALLOWED_PREFIXES = ["feat", "feature", "bugfix", "hotfix", "fix", "refactor", "chore", "doc", "test", "check"]
ALLOWED_SUFFIXES = ["draft", "nobuild", "no-build", "notest", "no-test", "norelease", "no-release"]

# Maximal commit-message len
MAX_MESSAGE_LENGTH = 192


def validate_commit_message(message):
    """
    regex based validation function for our commit-messages (<prefix>:<message>+<suffix:optional>)
    e.g. fix: handle bootstrap init-config failed in current io-handler +no-test
    """
    prefixes_pattern = '|'.join(ALLOWED_PREFIXES)
    suffixes_pattern = '|'.join(ALLOWED_SUFFIXES)

    pattern = rf'^(?:{prefixes_pattern}):\s.{{1,{MAX_MESSAGE_LENGTH}}}(?:\+(?:{suffixes_pattern}))?$'

    match = re.match(pattern, message)

    if not match:
        print("ERROR: Commit message does not match the required format.")
        return False

    if len(match.group(0).strip()) > MAX_MESSAGE_LENGTH:
        print(f"ERROR: Commit message exceeds {MAX_MESSAGE_LENGTH} characters.")
        return False

    return True


if __name__ == "__main__":
    commit_message_file = sys.argv[1]

    with open(commit_message_file, 'r') as file:
        commit_message = file.read().strip()

    if not validate_commit_message(commit_message):
        sys.exit(1)  # something doesn't match the rules, commit is rejected
