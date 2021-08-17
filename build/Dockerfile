FROM ubuntu:18.04 

RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /usr/sbin/nologin HwHiAiUser

COPY ./npu-exporter  /usr/local/bin/
RUN chown root:root /usr/local/bin/npu-exporter  &&\
    chmod 500 /usr/local/bin/npu-exporter &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> ~/.bashrc
ENV LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/add-ons:/usr/local/Ascend/driver/lib64:/usr/local/dcmi

CMD /usr/local/bin/npu-exporter

