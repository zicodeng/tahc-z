#!/usr/bin/env bash
docker run -it \
--network host \
--rm \
$INFO_344_MYSQL_IMAGE sh -c "mysql -h127.0.0.1 -uroot -p$MYSQL_ROOT_PASSWORD"
