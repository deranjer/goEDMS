FROM debian:latest
LABEL Author="deranjer"
LABEL name="goEDMS"
EXPOSE 8000
RUN mkdir /opt/goEDMS
RUN mkdir -p /opt/goEDMS/public/built
RUN mkdir /opt/goEDMS/config
RUN useradd goEDMS
WORKDIR /opt/goEDMS
COPY LICENSE /opt/goEDMS/config/LICENSE
COPY README.md /opt/goEDMS/config/README.md
COPY public/built/* /opt/goEDMS/public/built/
COPY dist/goEDMS_linux_amd64/goEDMS /opt/goEDMS/goEDMS
RUN chmod +x /opt/goEDMS/goEDMS
RUN chown -R goEDMS:goEDMS /opt/goEDMS/
RUN apt-get update && apt-get upgrade -y
RUN apt-get -y install imagemagick tesseract-ocr
ENTRYPOINT [ "/opt/goEDMS/goEDMS" ]

#docker build -t deranjer/goedms:latest .