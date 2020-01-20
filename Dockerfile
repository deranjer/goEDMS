FROM debian:latest
LABEL Author="deranjer"
LABEL name="goEDMS"
VOLUME [ "/opt/goEDMS/documents", "/opt/goEDMS/ingress", "/opt/goEDMS/temp" ]
EXPOSE 8000
RUN mkdir /opt/goEMDS
RUN mkdir -p /opt/goEDMS/public/built
RUN useradd goEDMS
WORKDIR /opt/goEDMS
#COPY dist-specific-files/Linux-systemd/goEDMS.service /etc/systemd/system/goEDMS.service
#RUN systemctl enable /etc/systemd/system/goEDMS.service
COPY dist/goEDMS_linux_amd64/goEDMS /opt/goEDMS/goEDMS
RUN chmod +x /opt/goEDMS/goEDMS
COPY LICENSE /opt/goEDMS/LICENSE
COPY README.md /opt/goEDMS/README.md
COPY public/built/* /opt/goEDMS/public/built/
COPY dist-specific-files/docker/serverConfig.toml /opt/goEDMS/serverConfig.toml
RUN chown -R goEDMS:goEDMS /opt/goEDMS/
RUN apt-get update && apt-get upgrade && apt-get -y install imagemagick tesseract-ocr
RUN apt-get -y install net-tools
#RUN systemctl start goEDMS
RUN ./goEDMS
