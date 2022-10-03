#!/bin/bash
sudo apt update
sudo apt -y full-upgrade
sudo apt -y install sudo curl wget lsb-release gnupg2 apt-transport-https
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
sudo sh -c 'echo "deb https://packagecloud.io/timescale/timescaledb/debian/ $(lsb_release -c -s) main" > /etc/apt/sources.list.d/timescaledb.list'
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
wget --quiet -O - https://packagecloud.io/timescale/timescaledb/gpgkey | sudo apt-key add -
sudo apt update
sudo apt -y install git postgresql-14 timescaledb-2-2.8.0-postgresql-14 timescaledb-tools timescaledb-toolkit-postgresql-14 grafana mosquitto
sudo timescaledb-tune  --quiet --yes
sudo bash <(curl -sL https://raw.githubusercontent.com/node-red/linux-installers/master/deb/update-nodejs-and-nodered) --confirm-root --confirm-install --skip-pi
sudo mkdir -p /root/.node-red
sudo wget https://raw.githubusercontent.com/noctarius/branded-workshop/main/settings.js -O /root/.node-red/settings.js
sudo /bin/systemctl daemon-reload
sudo systemctl enable nodered.service
sudo systemctl enable postgresql@
sudo systemctl enable grafana-server
