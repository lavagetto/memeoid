FROM golang:buster AS build
COPY . /src
RUN cd /src && go mod vendor && GOOS=linux GOARCH=amd64 go build .  -a -installsuffix cgo -ldflags="-w -s" -o /src/memeoid

# We're accepting the MS corefonts EULA implicitly.
# TODO: add a way to dynamically accept before build? Or at least a warning.
FROM debian:buster AS fonts
RUN echo 'deb http://deb.debian.org/debian buster contrib non-free' > /etc/apt/sources.list.d/extra.list \
    && apt-get update \
    && echo "ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true" | debconf-set-selections \
    &&  apt-get install -y ttf-mscorefonts-installer

FROM debian:buster
ENV USER=application
ENV UID=1000

# See https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Add the directory to copy the font to
RUN mkdir -p /usr/share/fonts/truetype/msttcorefonts && chmod 0755 /usr/share/fonts/truetype/msttcorefonts && mkdir -p /src/templates

COPY --from=build /src/memeoid /bin/
COPY --from=fonts /usr/share/fonts/truetype/msttcorefonts/Impact.ttf /usr/share/fonts/truetype/msttcorefonts
COPY templates /src/templates
# Add the user we will run as, and the /gif and /memes directories we'll be serving content from.
RUN mkdir -p /memes && mkdir -p /gifs \
    && chown ${USER} /memes && chown ${USER} /gifs \
    && chmod -R a+r /src/templates/*

# drop privileges
USER ${USER}

CMD [  "/bin/memeoid", "serve", "-i", "/gifs", "-m", "/memes", "--templates", "/src/templates", "-p", "3000", "-f", "Impact"]