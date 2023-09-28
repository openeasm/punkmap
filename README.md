# Punk Map

端口扫描和服务识别软件。特点：快、准、全部开源，持续更新。

## 支持协议 / Supported Protocols

## 2023-10月  
- [x] MySQL

### 2023-09月
- [x] FTP
- [x] HTTP(S)
- [x] SSH
- [x] Redis
- [x] Socks5
- [x] RDP
- [x] MongoDB

## Roadmap

- [ ] 对每个协议增加测试环境
- [ ] OEDB数据运营，增加端口识别率
- [ ] 一个Wiki并正式对外发布一个版本
- [ ] 增加benchmark的数据
- [ ] 增加和Nmap / Zgrab2的对比数据

## Example

Look, It's awesome!

```bash
➜  cmd git:(master) ✗ go run punk_map.go --print-metrics-interval=-1
dl01.imfht.com:22
{"service":"ssh","banner":"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.3\r\n","conn_ip":"5.255.108.100","ip":"dl01.imfht.com","open":true,"port":"22","protocol":"tcp","time":1695868891}
110.42.6.141:6379
{"service":"redis","banner":"-NOAUTH Authentication required.\r\n","ip":"110.42.6.141","open":true,"port":"6379","protocol":"tcp","time":1695868906}
92.205.7.100:21
{"service":"ftp","banner":"220---------- Welcome to Pure-FTPd [privsep] [TLS] ----------\r\n220-You are user number 1 of 500 allowed.\r\n220-Local time is now 19:42. Server port: 21.\r\n220-This is a private system - No anonymous login\r\n220-IPv6 connections are also welcome on this server.\r\n220 You will be disconnected after 15 minutes of inactivity.\r\n","ip":"92.205.7.100","open":true,"port":"21","protocol":"tcp","time":1695868923}
8.129.9.201:9876
{"service":"rocketmq","banner":"\u0000\u0000\u0000c\u0000\u0000\u0000_{\"code\":0,\"flag\":1,\"language\":\"JAVA\",\"opaque\":0,\"serializeTypeCurrentRPC\":\"JSON\",\"version\":252}","ip":"8.129.9.201","open":true,"port":"9876","protocol":"tcp","time":1695868992}
```

