FROM ubuntu:18.04
RUN useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX
RUN sed '1c root:x:0:0:root:/root:/usr/sbin/nologin'  /etc/passwd
COPY ./npu-exporter  /usr/local/bin/
COPY ./lib  /usr/local/lib
RUN chown root:root /usr/local/bin/npu-exporter  &&\
    chmod 550 /usr/local/bin/ &&\
    chmod 500 /usr/local/lib/* &&\
    chmod 500 /usr/local/bin/npu-exporter &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> ~/.bashrc
ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/add-ons:/usr/local/Ascend/driver/lib64:/usr/local/dcmi

CMD /usr/local/bin/npu-exporter

