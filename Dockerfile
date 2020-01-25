FROM debian:latest
LABEL Author="deranjer"
LABEL name="goEDMS"
#documents is the document library, ingress is the folder that takes in new documents, temp is the directory where temporary files are stored
VOLUME [ "/opt/goEDMS/documents", "/opt/goEDMS/ingress", "/opt/goEDMS/temp" ]
EXPOSE 8000
RUN mkdir /opt/goEMDS
RUN mkdir -p /opt/goEDMS/public/built
RUN useradd goEDMS
WORKDIR /opt/goEDMS
COPY LICENSE /opt/goEDMS/LICENSE
COPY README.md /opt/goEDMS/README.md
COPY public/built/* /opt/goEDMS/public/built/
COPY dist-specific-files/docker/serverConfig.toml /opt/goEDMS/serverConfig.toml
COPY dist/goEDMS_linux_amd64/goEDMS /opt/goEDMS/goEDMS
RUN chmod +x /opt/goEDMS/goEDMS
RUN chown -R goEDMS:goEDMS /opt/goEDMS/
RUN apt-get update && apt-get upgrade && apt-get -y install imagemagick tesseract-ocr
RUN apt-get -y install net-tools
RUN ./goEDMS
