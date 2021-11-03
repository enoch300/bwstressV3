#!/bin/bash
echo "Start install..."
systemctl stop bwstress
sleep 1
rm -rf /ipaas/bwstress
wget -q -O bwstress.tar.gz http://salt-source.oss-cn-hangzhou.aliyuncs.com/ipaas/bwstress/bwstress.tar.gz
tar xf bwstress.tar.gz -C /ipaas/
chmod +x /ipaas/bwstress/bin/*
chown -R root.root /ipaas/bwstress
cp -rf /ipaas/bwstress/conf/bwstress.service /lib/systemd/system/
systemctl daemon-reload
systemctl start bwstress
systemctl enable bwstress
echo "Install done"