#!/bin/bash
set +e
PREFIX=${PREFIX:-date %s}
DOCKER_ARGS=""

DOCKER_IMAGE=${DOCKER_IMAGE:-ehazlett/docker:17.06.2-ce}
DOCKER_ARGS=${DOCKER_ARGS:-}
DOCKER_VOLUME_PREFIX="vol"
NETWORK_NAME="stellar"
NODES="stellar-node-00 stellar-node-01"
STELLAR_IMAGE=${STELLAR_IMAGE:-ehazlett/stellar:dev}

function start_engine() {
    NODE=$1
    VOL_NAME=${DOCKER_VOLUME_PREFIX}-${NODE}
    docker volume create -d local ${VOL_NAME}
    docker run \
        --privileged \
        --net ${NETWORK_NAME} \
        --name ${NODE} \
        --hostname ${NODE} \
        --tmpfs /run \
        -v /lib/modules:/lib/modules:ro \
        -v ${VOL_NAME}:/var/lib/docker \
        -v /run/containerd/containerd.sock:/run/containerd/containerd.sock \
        -d \
        ${DOCKER_IMAGE} -H unix:// -s overlay2 ${DOCKER_ARGS}
    while true; do
        RES=$(docker exec -ti ${NODE} docker -v)
        if [ $? -eq 0 ]; then
            break
        fi
        sleep .5
    done

    # propagate mounts
    docker exec -ti ${NODE} mount --make-rshared "/"
}

function launch_nodes() {
    docker network create ${NETWORK_NAME}

    # check for image
    image_name=$(echo ${STELLAR_IMAGE} | awk -F: '{ print $1; }')

    exists=$(docker images | grep ${image_name})
    if [ -z "$exists" ]; then
        echo "You must build the Stellar image first (make binaries image)"
        exit 1
    fi

    INITIAL=""
    for NODE in ${NODES}; do
        start_engine ${NODE}

        docker save ${STELLAR_IMAGE} | docker exec -i $NODE docker load

        ip=$(docker exec -i ${NODE} ip a s eth0 | grep inet | awk '{ print $2; }' | awk -F/ '{ print $1; }')
        stellar_cmd="docker run -ti --name stellar -d --privileged --net=host -v /run/containerd/containerd.sock:/run/containerd/containerd.sock ${STELLAR_IMAGE}"
        if [ -z "$INITIAL" ]; then
            NODE=$NODE docker_cmd $stellar_cmd -D --bind-addr $ip --advertise-addr $ip
            INITIAL=$ip
        else
            NODE=$NODE docker_cmd $stellar_cmd -D --bind-addr $ip --advertise-addr $ip --peer $INITIAL:7946
        fi
    done
}

function docker_cmd() {
    docker exec -ti $NODE "$@"
}

function remove_nodes() {
    for NODE in ${NODES}; do
        docker kill ${NODE}
        docker network disconnect -f ${NETWORK_NAME} ${NODE}
        docker rm -fv ${NODE}
        docker volume rm -f ${DOCKER_VOLUME_PREFIX}-${NODE}
    done
}

function up() {
    launch_nodes
}

function down() {
    remove_nodes
}

case "$1" in
    "up")
        up
        ;;
    "down")
        down
        ;;
    "-h")
        echo "Usage: $0 <up|down>"
        exit 1
        ;;
esac
