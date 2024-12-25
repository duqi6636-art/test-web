echo '开始替换程序' 

cd /home/cherryproxy/program/

if [ -e "cherryproxy-web-api" ]; then
    bakName='cherryproxy-web-api-'`date -d "2 second" +"%m%d-%H%M"`
    mv -f 'cherryproxy-web-api' /home/cherryproxy/program/backups/$bakName
fi

mv -f /home/web/webmainnew /home/cherryproxy/program/cherryproxy-web-api

cd /home/cherryproxy/program/
chmod 777 'cherryproxy-web-api'

pidlist=`ps -ef | grep cherryproxy-web-api | grep -v 'grep' |awk '{print $2}'`

if [ "$pidlist" = "" ]
# 如果不存在
then
  echo "no ngrokd pid alive!" # 启动命令
  /usr/local/bin/grs 2 ./cherryproxy-web-api >  /dev/null 2>&1 &
  echo "启动。。。"
else
  #重启命令
  kill -1 $pidlist  && /usr/local/bin/grs 2 ./cherryproxy-web-api > /dev/null 2>&1 &
  echo "重启。。。"
fi

cd /home/
rm -rf  web