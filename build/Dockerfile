FROM registry.access.redhat.com/ubi9/ubi

COPY arts /usr/local/bin/arts

RUN chmod +x /usr/local/bin/arts && \
    chgrp -R 0 /usr/local/bin/arts && \
    chmod -R g=u /usr/local/bin/arts

CMD arts 