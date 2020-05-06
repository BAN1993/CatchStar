pwd=`pwd`

killall DataBaseServer
killall GameServer
killall GatewayServer

rm -rf log/*

$pwd/DataBaseServer 1>/dev/null 2>/dev/null &
$pwd/GameServer 1>/dev/null 2>/dev/null &
$pwd/GatewayServer 1>/dev/null 2>/dev/null &

