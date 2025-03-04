#!/usr/bin/env bash

set -e

cd cluster-deployment-automation
rm -rf /tmp/cda-venv
python3.11 -m venv /tmp/cda-venv
source /tmp/cda-venv/bin/activate
sh ./dependencies.sh
deactivate
cd -

cd ocp-traffic-flow-tests
rm -rf /tmp/tft-venv
python3.11 -m venv /tmp/tft-venv
source /tmp/tft-venv/bin/activate
pip install -r requirements.txt
deactivate
cd -
