######################################
# 编译脚本
# 编译目录下的多个服务并放入bin
# 每次编译都会自动清理
# 需要自己填充buildmap
######################################
# 编译列表
# ["服务名"]="路径"
# 路径为当前的相对路径
declare -A buildmap=(
	["DataBaseServer"]="DataBaseServer"
	["GameServer"]="GameServer"
	["GatewayServer"]="Gateway"
	["RobotClient"]="client/RobotClient"
)
######################################

buildFunc(){
	# 1.clean
	# 直接手动删除吧
	if [ -f "./bin/$1" ];then
		echo "[.] Clean ./bin/$1"
		rm -r ./bin/$1
	fi	

	# 2.build
	echo "[.] Begin Server:[$1]  Path:[$2]"
	go build -o ./bin/$1 ./$2
	if [ $? -eq 0 ]; then
		echo "[+] Build [$1] Success"
	else
		echo "[!] Build [$1] Fail"
	fi
}

for key in ${!buildmap[@]}
do
	buildFunc $key ${buildmap[$key]}
done

