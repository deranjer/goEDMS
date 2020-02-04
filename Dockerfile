# Stage 1
FROM alpine:latest as build
LABEL Author="deranjer"
LABEL name="goEDMS"
RUN mkdir -p /opt/goEDMS/public/built && \
  mkdir /opt/goEDMS/config && \
  adduser -S goEDMS && addgroup -S goEDMS
WORKDIR /opt/goEDMS
COPY LICENSE README.md /opt/goEDMS/config/
COPY public/built/* /opt/goEDMS/public/built/
COPY dist/goEDMS_linux_amd64/goEDMS /opt/goEDMS/goEDMS
RUN chmod +x /opt/goEDMS/goEDMS && \
  chown -R goEDMS:goEDMS /opt/goEDMS/ && \
  apk update && apk add imagemagick tesseract-ocr

# Stage 2
FROM scratch
COPY --from=build / /
EXPOSE 8000
WORKDIR /opt/goEDMS
ENTRYPOINT [ "/opt/goEDMS/goEDMS" ]

#docker build -t deranjer/goedms:latest .
