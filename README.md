# Kuta - devcontainer entrypoint

**STATUS: experimental**

A common pattern with docker and docker-compose is to volume mount the project
code into one or more containers. The goal is to allow faster feedback loop
than rebuilding a new container on every code change. And yet control the
developer environment.

A common issue with that approach is that the user UID and GID are not the
same between the host machine and inside the docker runtime. This causes
either access issues from the container side, or, if the container is running
as root, it leaves files owned by root on the host filesystem that needs `sudo`
to be able to delete or edit.

To work around this issue, this project introduces a small suid binary that
quickly changes the user inside the container to match whatever --user that is
passed to docker or docker-compose.

## Usage

```
ADD https://TODO /entrypoint
RUN chown 0:0 /entrypoint && chmod +xs /entrypoint
ENTRYPOINT ["/entrypoint"]
```

## Assumptions

The entrypoint script assumes that:
* bash is installed in the container.
* the USER environment variable is set to whatever main user in the container.

