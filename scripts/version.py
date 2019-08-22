#!/usr/bin/env python3
import os
import sys
import semver
import fileinput
from git import Repo

if sys.version_info[0] != 3:
    print("This script requires Python version 3")
    sys.exit(1)

# Helpers


def get_specter_path():
    script_path = os.path.dirname(os.path.realpath(sys.argv[0]))
    return script_path.split("scripts")[0]


def read_current_version():
    return open(get_specter_path() + 'VERSION', 'r').read()


def write_version(prev_version, new_version):
    version_file = open(get_specter_path() + 'VERSION', 'w')
    version_file.write(new_version)
    version_file.close()
    with fileinput.FileInput(get_specter_path() + 'web/public/package.json', inplace=True) as file:
        for line in file:
            print(line.replace('"version": "'+prev_version +
                               '"', '"version": "'+new_version+'"'), end='')
    with fileinput.FileInput(get_specter_path() + 'web/index.tmpl', inplace=True) as file:
        for line in file:
            print(line.replace('app.min.js?v='+prev_version,
                               'app.min.js?v='+new_version), end='')


# --User Input--
# input is used to read text (strings) from the user
try:
    version_type = input(
        'Select version type (major/minor/patch): ').strip().lower()
except:
    print("The version type is required and should be one of major/minor/patch")
    exit(1)

prev_version = read_current_version()

if version_type == 'major':
    new_version = semver.bump_major(prev_version)
    write_version(prev_version, new_version)
elif version_type == 'minor':
    new_version = semver.bump_minor(prev_version)
    write_version(prev_version, new_version)
elif version_type == 'patch':
    new_version = semver.bump_patch(prev_version)
    write_version(prev_version, new_version)
else:
    print("The version type is required and should be one of major/minor/patch")
    exit(1)

repo = Repo(get_specter_path())
assert not repo.bare
repo.git.commit('-am', 'Bump version to ' + new_version)
repo.create_tag('v' + new_version)

print('Version bumped to ' + new_version)
print('Now run: You should now merge this branch into master, then run "git push --tags"')
