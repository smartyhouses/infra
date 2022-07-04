# IMPORTANT: Don't specify the FROM field here. The FROM field (with additional configuration) is injected during runtime.
# We will have a proper Devbook based image in the future.
{{ .BaseDockerfile }}

RUN apk add curl

WORKDIR code
RUN npm init -y

{{ if .Deps }}
  RUN npm i {{ range .Deps }}{{ . }} {{ end }}

  # {
  #   "dep1": true
  #   ,"dep2": true
  # }
  RUN echo { >> /.dbkdeps.json
  {{ range $idx, $el := .Deps }}
    RUN echo '{{if $idx}},{{end}}"{{ $el }}": true' >> /.dbkdeps.json
  {{ end }}
  RUN echo } >> /.dbkdeps.json
{{ end }}

# Set env vars for devbook-daemon
RUN echo RUN_CMD=node >> /.dbkenv
# Format: RUN_ARGS=arg1 arg2 arg3
RUN echo RUN_ARGS=index.js >> /.dbkenv
RUN echo WORKDIR=/code >> /.dbkenv
# Relative to the WORKDIR env.
RUN echo ENTRYPOINT=index.js >> /.dbkenv

# Deps installation
RUN echo DEPS_CMD=npm >> /.dbkenv
RUN echo DEPS_INSTALL_ARGS=install >> /.dbkenv
RUN echo DEPS_UNINSTALL_ARGS=uninstall >> /.dbkenv

WORKDIR /