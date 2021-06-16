FROM ubuntu:18.04 

RUN useradd -d /home/hwMindX -u 9000 -m -s /bin/bash hwMindX && \
    useradd -d /home/HwHiAiUser -u 1000 -m -s /bin/bash HwHiAiUser && \
    usermod -a -G HwHiAiUser hwMindX

COPY ./npu-exporter  /usr/local/bin/
COPY ./lib  /usr/local/lib
RUN chown -R hwMindX:hwMindX /usr/local/bin/npu-exporter  &&\
    chmod +x /usr/local/bin/npu-exporter &&\
    echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> ~/.bashrc
ENV LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/add-ons:/usr/local/Ascend/driver/lib64:/usr/local/dcmi:/usr/local/dcmi
#暂时不使用hwMindX作为容器默认用户，待驱动支持非root监控后，可以启用
#USER hwMindX
CMD /usr/local/bin/npu-exporter -v=3

