FROM cgr.dev/chainguard/glibc-dynamic:latest@sha256:cabf47ee4e6e339b32a82cb84b6779e128bb9e1f2441b0d8883ffbf1f8b54dd2
ARG TARGETARCH
USER nonroot
COPY --chmod=0755 dist/binaries/oslc-request-server-linux-$TARGETARCH /usr/bin/oslc-request-server
ENTRYPOINT ["/usr/bin/oslc-request-server"]