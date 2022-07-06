# IMPORTANT: Don't specify the FROM field here. The FROM field (with additional configuration) is injected during runtime.
# We will have a proper Devbook based image in the future.
{{ .BaseDockerfile }}

RUN apk add --no-cache php php-curl

WORKDIR code
RUN touch index.php

# TODO: Instll passed deps

# Set env vars for devbook-daemon
RUN echo RUN_CMD=php >> /.dbkenv
# Format: RUN_ARGS=arg1 arg2 arg3
RUN echo RUN_ARGS=index.php >> /.dbkenv
RUN echo WORKDIR=/code >> /.dbkenv
# Relative to the WORKDIR env.
RUN echo ENTRYPOINT=index.php >> /.dbkenv

WORKDIR /