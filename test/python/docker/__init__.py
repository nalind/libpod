import configparser
import json
import os
import pathlib
import shutil
import subprocess
import tempfile

from docker import DockerClient

from .compat import constant


class Podman(object):
    """
    Instances hold the configuration and setup for running podman commands
    """

    def __init__(self):
        """Initialize a Podman instance with global options"""
        binary = os.getenv("PODMAN", "bin/podman")
        self.cmd = [binary, "--storage-driver=vfs"]

        cgroupfs = os.getenv("CGROUP_MANAGER", "systemd")
        self.cmd.append(f"--cgroup-manager={cgroupfs}")

        # No support for tmpfs (/tmp) or extfs (/var/tmp)
        # self.cmd.append("--storage-driver=overlay")

        if os.getenv("DEBUG"):
            self.cmd.append("--log-level=debug")
            self.cmd.append("--syslog=true")

        self.anchor_directory = tempfile.mkdtemp(prefix="podman_docker_")

        self.image_cache = os.path.join(self.anchor_directory, "cache")
        os.makedirs(self.image_cache, exist_ok=True)

        self.cmd.append("--root=" + os.path.join(self.anchor_directory, "crio"))
        self.cmd.append("--runroot=" + os.path.join(self.anchor_directory, "crio-run"))

        os.environ["CONTAINERS_REGISTRIES_CONF"] = os.path.join(
            self.anchor_directory, "registry.conf"
        )
        p = configparser.ConfigParser()
        p.read_dict(
            {
                "registries.search": {"registries": "['quay.io', 'docker.io']"},
                "registries.insecure": {"registries": "[]"},
                "registries.block": {"registries": "[]"},
            }
        )
        with open(os.environ["CONTAINERS_REGISTRIES_CONF"], "w") as w:
            p.write(w)

        os.environ["CNI_CONFIG_PATH"] = os.path.join(
            self.anchor_directory, "cni", "net.d"
        )
        os.makedirs(os.environ["CNI_CONFIG_PATH"], exist_ok=True)
        self.cmd.append("--cni-config-dir=" + os.environ["CNI_CONFIG_PATH"])
        cni_cfg = os.path.join(
            os.environ["CNI_CONFIG_PATH"], "87-podman-bridge.conflist"
        )
        # json decoded and encoded to ensure legal json
        buf = json.loads(
            """
            {
              "cniVersion": "0.3.0",
              "name": "default",
              "plugins": [{
                  "type": "bridge",
                  "bridge": "cni0",
                  "isGateway": true,
                  "ipMasq": true,
                  "ipam": {
                    "type": "host-local",
                    "subnet": "10.88.0.0/16",
                    "routes": [{
                      "dst": "0.0.0.0/0"
                    }]
                  }
                },
                {
                  "type": "portmap",
                  "capabilities": {
                    "portMappings": true
                  }
                }
              ]
            }
            """
        )
        with open(cni_cfg, "w") as w:
            json.dump(buf, w)

    def open(self, command, *args, **kwargs):
        """Podman initialized instance to run a given command

        :param self: Podman instance
        :param command: podman sub-command to run
        :param args: arguments and options for command
        :param kwargs: See subprocess.Popen() for shell keyword
        :return: subprocess.Popen() instance configured to run podman instance
        """
        cmd = self.cmd.copy()
        cmd.append(command)
        cmd.extend(args)

        shell = kwargs.get("shell", False)

        return subprocess.Popen(
            cmd,
            shell=shell,
            stdin=subprocess.DEVNULL,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
        )

    def run(self, command, *args, **kwargs):
        """Podman initialized instance to run a given command

        :param self: Podman instance
        :param command: podman sub-command to run
        :param args: arguments and options for command
        :param kwargs: See subprocess.Popen() for shell and check keywords
        :return: subprocess.Popen() instance configured to run podman instance
        """
        cmd = self.cmd.copy()
        cmd.append(command)
        cmd.extend(args)

        check = kwargs.get("check", False)
        shell = kwargs.get("shell", False)

        return subprocess.run(
            cmd,
            shell=shell,
            check=check,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )

    def tear_down(self):
        shutil.rmtree(self.anchor_directory, ignore_errors=True)

    def restore_image_from_cache(self, client: DockerClient):
        path = os.path.join(self.image_cache, constant.ALPINE_TARBALL)
        if not os.path.exists(path):
            img = client.images.pull(constant.ALPINE)
            with open(path, mode="wb") as tarball:
                for frame in img.save(named=True):
                    tarball.write(frame)
        else:
            self.run("load", "-i", path, check=True)

    def flush_image_cache(self):
        for f in pathlib.Path(self.image_cache).glob("*.tar"):
            f.unlink(f)
