% podman-volume(1)

## NAME
podman\-volume - Simple management tool for volumes

## SYNOPSIS
**podman volume** *subcommand*

## DESCRIPTION
podman volume is a set of subcommands that manage volumes.

## SUBCOMMANDS

| Command | Man Page                                               | Description                                                                    |
| ------- | ------------------------------------------------------ | ------------------------------------------------------------------------------ |
| create  | [podman-volume-create(1)](podman-volume-create.1.md)   | Create a new volume.                                                           |
| exists  | [podman-volume-exists(1)](podman-volume-exists.1.md)   | Check if the given volume exists.                                              |
| inspect | [podman-volume-inspect(1)](podman-volume-inspect.1.md) | Get detailed information on one or more volumes.                               |
| ls      | [podman-volume-ls(1)](podman-volume-ls.1.md)           | List all the available volumes.                                                |
| prune   | [podman-volume-prune(1)](podman-volume-prune.1.md)     | Remove all unused volumes.                                                     |
| rm      | [podman-volume-rm(1)](podman-volume-rm.1.md)           | Remove one or more volumes.                                                    |

## SEE ALSO
podman(1)

## HISTORY
November 2018, Originally compiled by Urvashi Mohnani <umohnani@redhat.com>
