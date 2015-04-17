#!/bin/bash

set -e

action=$1
container_path=$2
overlay_path=$2/overlay
rootfs_path=$2/rootfs
base_path=$3

function get_mountpoint() {
  df -P $1 | tail -1 | awk '{print $NF}'
}

function current_fs() {
  mountpoint=$(get_mountpoint $1)

  local mp
  local fs

  while read _ mp fs _; do
    if [ "$fs" = "rootfs" ]; then
      continue
    fi

    if [ "$mp" = "$mountpoint" ]; then
      echo $fs
    fi
  done < /proc/mounts
}

function should_use_overlayfs() {
  # load it so it's in /proc/filesystems
  modprobe -q overlayfs >/dev/null 2>&1 || true

  # cannot mount overlayfs in aufs
  if [ "$(current_fs $overlay_path)" == "aufs" ]; then
    return 1
  fi

  # cannot mount overlayfs in overlayfs; whiteout not supported
  if [ "$(current_fs $overlay_path)" == "overlayfs" ]; then
    return 1
  fi

  # check if it's a known filesystem
  grep -q overlayfs /proc/filesystems
}

function should_use_aufs() {
  # load it so it's in /proc/filesystems
  modprobe -q aufs >/dev/null 2>&1 || true

  # cannot mount aufs in aufs
  if [ "$(current_fs $overlay_path)" == "aufs" ]; then
    return 1
  fi

  # cannot mount aufs in overlayfs
  if [ "$(current_fs $overlay_path)" == "overlayfs" ]; then
    return 1
  fi

  # check if it's a known filesystem
  grep -q aufs /proc/filesystems
}

function setup_fs() {
  mkdir -p $overlay_path
  mkdir -p $rootfs_path

  if should_use_aufs; then
    mount -n -t aufs -o br:$overlay_path=rw:$base_path=ro+wh none $rootfs_path
  elif should_use_overlayfs; then
    mount -n -t overlayfs -o rw,upperdir=$overlay_path,lowerdir=$base_path none $rootfs_path
  else
    # aufs and overlayfs are the only supported mount types.
    # aufs and overlayfs can be used in nested containers by mounting
    # the overlay directories on e.g. tmpfs
    echo "the directories that contain the depot and rootfs must be mounted on a filesystem type that supports aufs or overlayfs" >&2
    exit 222
  fi
}

function try_umount() {
  local mountpoint=$1
  if [ ! -z "$mountpoint" ]; then
    umount -l "$mountpoint" || true
  fi
}

function teardown_fs() {
  export -f try_umount
  local file=/proc/mounts
  local mount_points=$(cat $file | grep $container_path | awk '{print $2}')

  for mount_point in $mount_points; do
    echo "$(echo $mount_point | wc -m) $mount_point"
  done | sort -n -r | awk '{ print $2 }' | xargs bash -c 'try_umount "$@"' _
}

if [ "$action" = "create" ]; then
  setup_fs
else
  teardown_fs
fi
