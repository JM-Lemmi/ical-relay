#! /bin/bash
DIR="$(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
cd "$DIR" || exit 1

cd ../../cmd/ical-relay/templates/static/js-dep/

set -e

wget https://cdn.jsdelivr.net/npm/jquery@3.6.3/dist/jquery.slim.min.js
wget https://cdn.jsdelivr.net/npm/select2@4.0.13/dist/js/select2.min.js
wget https://cdn.jsdelivr.net/npm/bootstrap@5.2.3/dist/js/bootstrap.min.js
wget https://cdn.jsdelivr.net/npm/bootstrap@5.2.3/dist/css/bootstrap.min.css
wget -P dayjs@1.11.7 https://cdn.jsdelivr.net/npm/dayjs@1.11.7/dayjs.min.js
wget -P dayjs@1.11.7/locale https://cdn.jsdelivr.net/npm/dayjs@1.11.7/locale/de.js
wget -P dayjs@1.11.7/plugin https://cdn.jsdelivr.net/npm/dayjs@1.11.7/plugin/advancedFormat.js
wget -P dayjs@1.11.7/plugin https://cdn.jsdelivr.net/npm/dayjs@1.11.7/plugin/weekOfYear.js
wget https://cdn.jsdelivr.net/npm/select2@4.0.13/dist/css/select2.min.css
wget https://cdn.jsdelivr.net/npm/@ttskch/select2-bootstrap4-theme@1.5.2/dist/select2-bootstrap4.min.css
