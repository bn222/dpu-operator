git submodule init
git submodule update

cd cluster-deployment-automation
systemctl restart libvirtd

python3.11 -m venv /tmp/ocp-venv
source /tmp/ocp-venv/bin/activate
sh ./dependencies.sh

