taskkill /im DataBaseServer.exe /F /T
taskkill /im GameServer.exe /F /T
taskkill /im Gateway.exe /F /T

start DataBaseServer.exe
start GameServer.exe
start Gateway.exe