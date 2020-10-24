# Memeoid

This is a simple meme generator with support for animated gifs.

I could not find another FLOSS meme generator which supported them, so I created a new one.

Still in super-early stages of development, probably full of very bad design choices as I cargo-culted my way around GIF manipulation.

Still, with the provided dockerfile you have a functioning meme generator you can use.

## Installation

Docker is the only supported platform. Build your image by running
```bash
$ docker build . -t memeoid:latest
```
Please note: by running the build, you'll implicitly accept the EULA for the Microsoft core fonts, as the docker image uses Microsoft Impact by default.

By default the image is built to run memeoid as user 1000, for ease of use by me during development. You should export the UID variable if you want to change that.

## Running

First, you need a directory containing gifs you want to use as base for the memes. The directory should be accessible to the UID we chose at build time. 

Optionally, create a directory where the output memes will be saved.

Now you can run memeoid

```bash
# UID is the UID you used when building the image.
# GIFDIR is the directory where the originals are contained
# MEMEDIR is the output directory for the memes
# PORT is the port you want to expose the service on
chmod 0755 $GIFDIR
chown -R $UID
docker run --rm -p $PORT:3000 -v $GIFDIR:/gifs:ro \
    -v $MEMEDIR:/memes:rw memeoid:latest
```

## Modifying templates without a rebuild
You can just mount your templates directory in your docker run:
```bash
docker run --rm -p $PORT:3000 -v $GIFDIR:/gifs:ro \
    -v $TEMPLATEDIR:/src/templates:ro \
    -v $MEMEDIR:/memes:rw memeoid:latest
```
Just remember that directory needs to be readable by the user running the application.