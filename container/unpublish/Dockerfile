FROM alpine:latest

RUN apk update;
RUN apk add --no-cache nodejs npm jq

RUN npm install -g wrangler@latest

WORKDIR /project

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]