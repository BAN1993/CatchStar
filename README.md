# CatchStar
 cococ creator,typescript,golang

 - 基于cocos creator官方教程编写的联网对战实例,所有素材均来自于官方资源  
 - 客户端脚本由TypeScript编写  
 - 服务端由Golang编写  
	基于自己的"DrillServerGo"框架(Gateway+GameServer+DataBaseServer)  
 - 双人联机对战,采用帧同步方式  
	有对网络延迟做了简单优化(预测+修复位置)  
 - 运行:  
	客户端:  
		1.依次导入资源包LoginScence.zip和GameScence.zip到assets下(若提示重复直接覆盖即可)  
		2.选择Login场景即可开始调试  
		3.服务端地址配置在 NetConfig.ts  
	-服务端:  
		1.需要安装Go,mysql  
		2.执行根目录下的buildall.sh编译  
		3.根据需要修改bin目录下的三个(gw_config.ini,gs_config.ini,dbs_config.ini)配置文件  
			dbs_config.ini的mysql项配置数据库连接,redis暂时用不到  
		4.数据库要创建一个drillserver库,并创建一张Players表  
			CREATE TABLE `players` (  
				`numid` int(11) NOT NULL AUTO_INCREMENT,  
				`account` varchar(64) NOT NULL,  
				`nickname` varchar(64) NOT NULL,  
				`passwd` varchar(64) NOT NULL,  
				PRIMARY KEY (`numid`),  
				UNIQUE KEY `account` (`account`) USING BTREE  
			) ENGINE=InnoDB DEFAULT CHARSET=utf8;  
		5.插入两条数据方便测试  
			INSERT INTO `drillserver`.`players` (`numid`, `account`, `nickname`, `passwd`) VALUES ('100000', 'test01', 'test01', '123456');  
			INSERT INTO `drillserver`.`players` (`numid`, `account`, `nickname`, `passwd`) VALUES ('100001', 'test02', 'test02', '123456');  
		6.执行rall.sh(linux)或restart.bat(windows)开启服务  
