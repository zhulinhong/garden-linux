#!/bin/bash

[ -n "$DEBUG" ] && set -o xtrace
set -o nounset
set -o errexit
shopt -s nullglob

cd $(dirname $0)/../

source etc/config

# write uid map if user namespacing is enabled
# if [ "$root_uid" -ne 0 ]
# then
# cat > /proc/$PID/uid_map <<EOF
# 0 $root_uid 1
# $user_uid $user_uid 1
# EOF

# cat > /proc/$PID/gid_map <<EOF
# 0 $root_uid 1
# $user_uid $user_uid 1
# EOF
# fi

# Add new group for every subsystem

# cpuset must be set up first, so that cpuset.cpus and cpuset.mems is assigned
# otherwise adding the process to the subsystem's tasks will fail with ENOSPC
for system_path in ${GARDEN_CGROUP_PATH}/{cpuset,cpu,cpuacct,devices,memory}
do
  instance_path=$system_path/instance-$id

  mkdir -p $instance_path

  if [ $(basename $system_path) == "cpuset" ]
  then
    cat $system_path/cpuset.cpus > $instance_path/cpuset.cpus
    cat $system_path/cpuset.mems > $instance_path/cpuset.mems
  fi

  if [ $(basename $system_path) == "devices" ]
  then
    # Deny everything, allow explicitly
    echo a > $instance_path/devices.deny

    # Allow mknod for everything.
    echo "c *:* m" > $instance_path/devices.allow
    echo "b *:* m" > $instance_path/devices.allow

    # /dev/null
    echo "c 1:3 rwm" > $instance_path/devices.allow
    # /dev/zero
    echo "c 1:5 rwm" > $instance_path/devices.allow
    # /dev/full
    echo "c 1:7 rwm" > $instance_path/devices.allow
    # /dev/random
    echo "c 1:8 rwm" > $instance_path/devices.allow
    # /dev/urandom
    echo "c 1:9 rwm" > $instance_path/devices.allow
    # /dev/tty0
    echo "c 4:0 rwm" > $instance_path/devices.allow
    # /dev/tty1
    echo "c 4:1 rwm" > $instance_path/devices.allow
    # /dev/tty
    echo "c 5:0 rwm" > $instance_path/devices.allow
    # /dev/console
    echo "c 5:1 rwm" > $instance_path/devices.allow
    # /dev/ptmx
    echo "c 5:2 rwm" > $instance_path/devices.allow
    # /dev/pts/*
    echo "c 136:* rwm" > $instance_path/devices.allow
    # tuntap (?)
    echo "c 10:200 rwm" > $instance_path/devices.allow
    # /dev/fuse
    echo "c 10:229 rwm" > $instance_path/devices.allow
  fi

  echo $PID > $instance_path/tasks
done

echo $PID > ./run/wshd.pid

exit 0
