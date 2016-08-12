# coding=utf8

"""
Fabfile (http://www.fabfile.org/) to deploy banshee to remote host.

Requirements: fabric

Example usage:

    $ python deploy.py -u hit9 -H remote-host:22 --remote-path "/srv/banshee"
      --remote-user root:root --refresh

This script will do the following jobs:

    1. Install static depdencies.
    2. Build static files.
    3. Build banshee binary.
    4. Rsync the static files and binary to remote host.
    5. Restart banshee service via supervisorctl.

Note:

    1. If the remote host is linux, this script should be also called
    on linux to build the right binary, cause:
        https://github.com/mattn/go-sqlite3/issues/106
    2. The banshee service should be maintained in supervisor. You need
    to create a new service named banshee in supervisor.
"""

import os
import argparse
import contextlib

from fabric.api import (
    abort,
    env,
    execute,
    local,
    sudo,
    task,
    warn,
)
from fabric.contrib.files import exists
from fabric.contrib.project import rsync_project

###
# Global
###

LOCAL_DIR = "deploy-local-tmp"
LOCAL_STATIC_DIR = os.path.join(LOCAL_DIR, "static", "dist")
BINARY_NAME = "banshee"
STATIC_DIR = os.path.join("static", "dist")
SERVICE_NAME = "banshee"

###
# Local
###


def record_commit():
    """Record current commit.
    """
    local("git rev-parse HEAD > commit")


def install_static_deps():
    """Install local static dependencies.
    """
    local("cd static && npm install -q")
    local("cd static/public && npm install -q")


def build_static_files():
    """Build static files via gulp.
    """
    local("cd static && rm -rf dist/")
    local("cd static && gulp build")


def build_binary():
    """Build banshee binary via makefile.
    """
    local("make")


def make_local_dir():
    """Make local temporary directory.

        deploy-local-tmp/
            |- commit
            |- binary
            |- static/
                |- dist/
                    |- css/
                    |- js/
                    ...
    """
    local("mkdir -p {}".format(LOCAL_STATIC_DIR))
    local("cp {0} {1}".format(BINARY_NAME, LOCAL_DIR))
    local("cp -r {0}/* {1}".format(STATIC_DIR, LOCAL_STATIC_DIR))
    local("mv commit {}".format(LOCAL_DIR))
    extra_files = env.extra_files.split(";")
    for f in extra_files:
        if f != "":
            local("cp {0} {1}".format(f,LOCAL_DIR))

def remove_local_dir():
    """Remove local temporary directory.
    """
    local("rm -rf {}".format(LOCAL_DIR))


@contextlib.contextmanager
def local_tmp_build():
    try:
        record_commit()
        install_static_deps()
        build_static_files()
        if not env.only_static:
            build_binary()
        make_local_dir()
        yield
    finally:
        remove_local_dir()


###
# Remote
###


def upload():
    """Upload local directory to remote directory.
    """
    if not exists(env.remote_path):
        sudo("mkdir -p {}".format(env.remote_path))
    sudo("chmod 755 {}".format(env.remote_path))
    sudo("chown -R {0} {1}".format(env.user, env.remote_path))
    if env.only_static:
        rsync_project(env.remote_path, LOCAL_DIR + '/', exclude=BINARY_NAME)
    else:
        rsync_project(env.remote_path, LOCAL_DIR + '/')
    sudo("chown -R {0} {1}".format(env.remote_user, env.remote_path))


def refresh():
    """Refresh service via supervisor.
    """
    sudo("supervisorctl restart {}".format(SERVICE_NAME))


@task
def deploy():
    """Deploy banshee.
    """
    upload()
    if env.refresh and not env.only_static:
        refresh()


def main(host=None, user=None):
    parser = argparse.ArgumentParser()
    parser.add_argument('-u', '--user', help="user to connect")
    parser.add_argument('-H', '--hosts', help="hosts to deploy", required=True)
    parser.add_argument('--refresh', help="whether to refresh service",
                        action='store_true', default=False)
    parser.add_argument('--remote-path', help="remote service path",
                        required=True)
    parser.add_argument("--remote-user", help="remote service user",
                        default="root:root")
    parser.add_argument("--only-static", help="deploy only static files",
                        action='store_true', default=False)
    parser.add_argument("--extra-files",help="deploy extra files",default="")
    args = parser.parse_args()

    hosts = map(lambda s: s.strip(),  args.hosts.split(','))

    if not hosts:
        abort("No hosts to deploy.")

    if not args.user:
        warn("Using default user: {}".format(env.user))
    else:
        env.user = args.user

    if not args.remote_user:
        env.remote_user = "root:root"
        warn("Using default remote user: root:root")
    else:
        env.remote_user = args.remote_user

    env.remote_path = args.remote_path
    env.only_static = args.only_static
    env.refresh = args.refresh
    env.use_ssh_config = True
    env.parallel = True
    env.extra_files = args.extra_files

    with local_tmp_build():
        execute(deploy, hosts=hosts)


if __name__ == '__main__':
    main()
