#!/bin/bash
set -e

usage() {
    cat << EOF
    Usage: GIFDIR=<dir> MEMEDIR=<dir> $0

    Runs memeoid. Behaviour is controlled by the following environment 
    variables:

    * GIFDIR is the directory where the originals are contained
    * MEMEDIR is the output directory for the memes
    * PORT is the port you want to expose the service on. By default it's port 3000

    You should also ensure the UID variable is set to the same value you used
    when building the image.
EOF
    exit 2
}
test -z $GIFDIR && usage
test -z $MEMEDIR && usage
if [ "$1" == "-h" ]; then
    usage
fi

G=$(realpath $GIFDIR)
M=$(realpath $MEMEDIR)
PORT=${PORT:-3000}
chmod 0755 $G
chown -R $UID $M
MSG="Launching memeoid on :$PORT, using '$G' as a gif source"
if which cowsay > /dev/null; then
    cowsay "$MSG"
else
    echo "Please install `cowsay` for a better user experience."
    echo "$MSG"
fi
if [ -z $TEMPLATEDIR ]; then
    docker run --rm -p $PORT:3000 -v $G:/gifs:ro -v $M:/memes:rw memeoid:latest
else
    T=$(realpath $TEMPLATEDIR)
    docker run --rm -p $PORT:3000 -v $T:/src/templates:ro -v $G:/gifs:ro -v $M:/memes:rw memeoid:latest
fi