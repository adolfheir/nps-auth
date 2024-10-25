cp /ihouqi/nps-auth/systemd/nps-auth.service /etc/systemd/system/nps-auth.service

sudo systemctl daemon-reload

sudo systemctl start nps-auth
sudo systemctl stop nps-auth

sudo systemctl status nps-auth

sudo systemctl restart nps-auth


ps |grep nps-quth |  grep -v grep