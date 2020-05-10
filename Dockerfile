# golang alpine 1.14.2
FROM golang@sha256:b0678825431fd5e27a211e0d7581d5f24cede6b4d25ac1411416fa8044fa6c51 as builder 

ENV USER_UID=10001 \
    USER_NAME=api \
    HOME=/opt/api

# ensure home exists and is accessible by group 0 (we don't know what the runtime UID will be)
RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd \
    && mkdir -p "${HOME}" \
    && chown "${USER_UID}:0" "${HOME}" \
    && chmod ug+rwx "${HOME}"

WORKDIR /go/src/go-prometheus-exporter
COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s -extldflags "-static"' -o /go/bin/api .

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /opt/api /opt/
COPY --from=builder /go/bin/api /bin/

USER ${USER_UID}
EXPOSE 4778
ENTRYPOINT ["/bin/api"]
