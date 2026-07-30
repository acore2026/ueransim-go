# NAS Cooperation 测试报告

## 1. 测试概述

本报告记录了使用 ueransim-go 对 free5gc AMF（commit `b489a3d`）两种 NAS Cooperation 功能的端到端测试：

| 测试项 | NAS 消息类型 | PayloadContainerType | 结果 |
|--------|-------------|---------------------|------|
| NAS Transport Passthrough | UL=0x67 / DL=0x68（标准 NAS Transport） | 4 | PASS |
| AP Container Cooperation | UL=0xE1 / DL=0xE2（自定义 Cooperation） | 0x0101 | PASS |

两种模式均完成 UE 注册 -> 上行消息发送 -> AMF NAgent 转发 -> NAgent Mock 响应 -> 下行消息回传的完整闭环。

---

## 2. 测试环境

### 2.1 组件版本

| 组件 | 版本 |
|------|------|
| AMF 镜像 | `free5gc/amf:buildv16`（源码 commit `b489a3d`） |
| AMF 编译 | Go 1.25.5, CGO_ENABLED=0 |
| UERANSIM | v3.3.0（ueransim-go `NewNAS-ApContaine` 分支） |
| free5gc 其他 NF | v4.2.1（NRF/SMF/UDM/UDR/AUSF/PCF/NSSF/CHF/NEF/WebUI） |
| MongoDB | 4.4 |
| Docker | legacy builder |

### 2.2 网络拓扑

```
宿主机 (10.100.200.1, br-free5gc)
  |
  +-- nr-gnb (10.100.200.1)        <-- UERANSIM gNB, 宿主机进程
  |     |
  |     +-- SCTP/N2 --> amf (10.100.200.12:38412)  <-- Docker 容器
  |
  +-- nr-ue (10.100.200.1)         <-- UERANSIM UE, 宿主机进程
        |
        +-- RLS/UDP --> nr-gnb
        +-- NAS via NGAP --> AMF
```

### 2.3 AMF NAgent 配置

```yaml
nagent:
  enabled: true
  baseUri: http://127.0.0.1:8088
  mock:
    enabled: true
    listenAddress: 127.0.0.1:8088
    delayMs: 8000          # 模拟 8 秒处理延迟
    status: 200
  transportPassthrough:
    enabled: true
    payloadContainerType: 4
    maxPayloadBytes: 1400
```

### 2.4 UE Subscriber 信息

| 字段 | 值 |
|------|-----|
| SUPI | `imsi-001010000000001` |
| MCC/MNC | `001/01` |
| K (encPermanentKey) | `465B5CE8B199B49FAA5F0A2EE238A6BC` |
| OPC (encOpcKey) | `E8ED289DEBA952E4283B54E88E6183CA` |
| AMF | `8000` |
| S-NSSAI | SST=1, SD=0x000000 |
| DNN | `internet` |

---

## 3. 前置准备

### 3.1 构建 UERANSIM

```bash
cd ueransim-go
make build
# 产物：build/nr-gnb, build/nr-ue, build/nr-cli
```

依赖：`cmake`, `libsctp-dev`, `g++`

### 3.2 gNB 配置 (`config/free5gc-gnb.yaml`)

```yaml
mcc: '001'
mnc: '01'
linkIp: 10.100.200.1
ngapIp: 10.100.200.1
gtpIp: 10.100.200.1
amfConfigs:
  - address: 10.100.200.12    # AMF 容器 IP
    port: 38412
slices:
  - sst: 0x1
    sd: 0x000000
ignoreStreamIds: true
```

> 注意：gNB 需直连 AMF 容器 IP（10.100.200.12），Docker 的 SCTP 端口映射存在 COOKIE_WAIT 问题。gNB 的 `linkIp`/`ngapIp`/`gtpIp` 应设为宿主机在 Docker 网络（br-free5gc）上的地址 10.100.200.1。

### 3.3 UE 配置 (`config/free5gc-ue.yaml`)

```yaml
supi: 'imsi-001010000000001'
mcc: '001'
mnc: '01'
key: '465B5CE8B199B49FAA5F0A2EE238A6BC'
op: 'E8ED289DEBA952E4283B54E88E6183CA'
opType: 'OPC'
amf: '8000'

gnbSearchList:
  - 10.100.200.1

sessions:
  - type: 'IPv4'
    apn: 'internet'
    slice:
      sst: 0x01
      sd: 0x000000

configured-nssai:
  - sst: 0x01
    sd: 0x000000

# AP Container 测试开关
cooperationTest:
  enabled: true          # <-- AP Container 测试
  messageIdentity: 1
  containerType: 0x0101
  pti: 0
  payloadId: 0x1234
  flags: 0x02
  includeIe10: true
  ie10Value: 1
  intentId: 'ueransim-test'
  issuer: 'ueransim'
  intentPriority: 1
  intentType: 'location'
  intentDescription: 'locate target'
  object: 'ue'
  constraint: ''
  target: 'amf-test'

# NAS Transport 测试开关
nasTransportTest:
  enabled: false         # <-- NAS Transport 测试（与 cooperationTest 互斥启用）
  payloadContainerType: 4
  payload: '{"intentType":"location","target":"ue-001","description":"locate target UE"}'
```

### 3.4 启动流程

```bash
# 1. 启动 free5gc 核心网（docker-compose）
docker-compose up -d

# 2. 停止 UPF 释放 GTP 2152 端口（gNB 需要绑定该端口）
docker stop upf

# 3. 启动 gNB
./build/nr-gnb -c config/free5gc-gnb.yaml &

# 4. 启动 UE
./build/nr-ue -c config/free5gc-ue.yaml &
```

---

## 4. 测试一：NAS Transport Passthrough

### 4.1 测试配置

UE 配置中启用：
```yaml
nasTransportTest:
  enabled: true
  payloadContainerType: 4
  payload: '{"intentType":"location","target":"ue-001","description":"locate target UE"}'
```

AMF 配置中启用：
```yaml
nagent:
  transportPassthrough:
    enabled: true
    payloadContainerType: 4
    maxPayloadBytes: 1400
```

### 4.2 流程说明

```
UE                     gNB                    AMF                  NAgent Mock
 |                       |                     |                       |
 |--Registration------->|---NGAP------------->|                       |
 |<--Auth/Security----->|<---NGAP------------|                       |
 |<--Reg Accept-------->|<---NGAP------------|                       |
 |--Reg Complete------->|---NGAP------------->|                       |
 |                       |                     |                       |
 |--UL NAS Transport--->|---UplinkNASTrans-->|                       |
 |  (type=0x67,          |                     |--HTTP POST---------->|
 |   PayloadContainer    |                     |  body=payload(76B)    |
 |   Type=4,             |                     |                       |
 |   payload=JSON)       |                     |     [8s mock delay]   |
 |                       |                     |<--HTTP 200------------|
 |                       |                     |  body=response(76B)  |
 |<--DL NAS Transport---|<--DlinkNASTrans----|                       |
 |  (type=0x68,          |                     |                       |
 |   PayloadContainer    |                     |                       |
 |   Type=4,             |                     |                       |
 |   payload=JSON)       |                     |                       |
```

### 4.3 UE 日志

```
[07:45:59.046] [nas] [info] UE switches to state [MM-REGISTER-INITIATED]
[07:45:59.046] [rrc] [info] RRC connection established
[07:45:59.055] [nas] [debug] Authentication Request received
[07:45:59.059] [nas] [debug] Security Mode Command received
[07:45:59.092] [nas] [debug] Registration accept received
[07:45:59.092] [nas] [info] UE switches to state [MM-REGISTERED/NORMAL-SERVICE]
[07:45:59.092] [nas] [debug] Sending Registration Complete
[07:45:59.092] [nas] [info] Sending UL NAS Transport test NAS plain=[7E006704004C7B22696E74656E7454797065223A226C6F636174696F6E222C22746172676574223A2275652D303031222C226465736372697074696F6E223A226C6F6361746520746172676574205545227D]
[07:45:59.092] [nas] [info] Initial Registration is successful
...（8 秒等待 NAgent Mock 响应）...
[07:46:07.295] [nas] [info] DL NAS Transport type=4 received payloadHex=[7B22696E74656E7454797065223A226C6F636174696F6E222C22746172676574223A2275652D303031222C226465736372697074696F6E223A226C6F6361746520746172676574205545227D]
[07:46:07.295] [nas] [info] DL NAS Transport type=4 payloadText=[{"intentType":"location","target":"ue-001","description":"locate target UE"}]
```

### 4.4 AMF 日志

```
[07:45:59.293] [INFO][AMF][Ngap] Send Downlink Nas Transport
[07:45:59.293] [INFO][AMF][Gmm] Handle UL NAS Transport
[07:45:59.293] [INFO][AMF][Gmm] [NAgent NAS Transport] Sending passthrough payload: supi=imsi-001010000000001 accessType=3GPP_ACCESS payloadLength=76
[07:46:07.295] [INFO][AMF][Gmm] [NAgent NAS Transport] Received success response: responseLength=76
[07:46:07.295] [INFO][AMF][Gmm] Send DL NAS Transport
[07:46:07.295] [INFO][AMF][Ngap] Send Downlink Nas Transport
```

### 4.5 时间线

| 时间 | 事件 |
|------|------|
| 07:45:59.046 | UE 开始注册 |
| 07:45:59.092 | UE 注册成功，发送 UL NAS Transport (type=0x67) |
| 07:45:59.293 | AMF 收到 UL NAS Transport，转发到 NAgent (payloadLength=76) |
| 07:46:07.295 | NAgent Mock 响应返回 (responseLength=76)，AMF 发送 DL NAS Transport |
| 07:46:07.295 | UE 收到 DL NAS Transport (type=4)，payload 完整匹配 |

**端到端延迟**：8.2 秒（NAgent Mock delayMs=8000）

### 4.6 验证点

- [x] UE 成功完成 5G-AKA 认证和注册
- [x] UE 发送 UL NAS Transport，PayloadContainerType=4
- [x] AMF 将 PayloadContainer 原样转发给 NAgent（payloadLength=76）
- [x] NAgent Mock 处理请求后返回响应（responseLength=76）
- [x] AMF 通过 DL NAS Transport 将响应回传 UE
- [x] UE 成功接收 DL NAS Transport，PayloadContainerType=4
- [x] 上行 payload 与下行 payload 内容完全一致（原样透传）

---

## 5. 测试二：AP Container Cooperation

### 5.1 测试配置

UE 配置中启用：
```yaml
cooperationTest:
  enabled: true
  messageIdentity: 1
  containerType: 0x0101
  pti: 0            # 0 = 自动生成
  payloadId: 0x1234
  flags: 0x02
  includeIe10: true
  ie10Value: 1
  intentId: 'ueransim-test'
  issuer: 'ueransim'
  intentPriority: 1
  intentType: 'location'
  intentDescription: 'locate target'
  object: 'ue'
  constraint: ''
  target: 'amf-test'
```

### 5.2 流程说明

```
UE                     gNB                    AMF                  NAgent Mock
 |                       |                     |                       |
 |--Registration------->|---NGAP------------->|                       |
 |<--Auth/Security----->|<---NGAP------------|                       |
 |<--Reg Accept-------->|<---NGAP------------|                       |
 |--Reg Complete------->|---NGAP------------->|                       |
 |                       |                     |                       |
 |--UL Cooperation----->|---UplinkNASTrans-->|                       |
 |  (type=0xE1,          |                     |                       |
 |   AP Container IE    |                     |  解析 TLV IE list      |
 |   IEI=0x71,           |                     |  重组 AP Container     |
 |   type=0x0101,        |                     |  提取 payload(177B)    |
 |   payload=JSON)       |                     |--HTTP POST---------->|
 |                       |                     |  body=payload(177B)   |
 |                       |                     |     [8s mock delay]   |
 |                       |                     |<--HTTP 200------------|
 |                       |                     |  body=response(177B) |
 |<--DL Cooperation----|<--DlinkNASTrans----|                       |
 |  (type=0xE2,          |                     |                       |
 |   AP Container IE    |                     |                       |
 |   payload=JSON)       |                     |                       |
```

### 5.3 UE 日志

```
[07:49:42.344] [nas] [debug] Authentication Request received
[07:49:42.347] [nas] [debug] Security Mode Command received
[07:49:42.353] [nas] [debug] Registration accept received
[07:49:42.353] [nas] [info] UE switches to state [MM-REGISTERED/NORMAL-SERVICE]
[07:49:42.353] [nas] [debug] Sending Registration Complete
[07:49:42.353] [nas] [info] Sending UL Cooperation test NAS plain=[7E00E10110010171BB010100B70112340200007B22696E74656E744964223A22756572616E73696D2D74657374222C22697373756572223A22756572616E73696D222C22696E74656E745072696F72697479223A312C22696E74656E7454797065223A226C6F636174696F6E222C22696E74656E744465736372697074696F6E223A226C6F6361746520746172676574222C226F626A656374223A227565222C22636F6E73747261696E74223A22222C22746172676574223A22616D662D74657374227D]
[07:49:42.353] [nas] [info] Initial Registration is successful
...（8 秒等待 NAgent Mock 响应）...
[07:49:50.560] [nas] [info] DL Cooperation received plain=[7E00E20110010171BB010100B70112340000007B22696E74656E744964223A22756572616E73696D2D74657374222C22697373756572223A22756572616E73696D222C22696E74656E745072696F72697479223A312C22696E74656E7454797065223A226C6F636174696F6E222C22696E74656E744465736372697074696F6E223A226C6F6361746520746172676574222C226F626A656374223A227565222C22636F6E73747261696E74223A22222C22746172676574223A22616D662D74657374227D]
[07:49:50.560] [nas] [info] DL Cooperation IE 0x10 value=[01]
[07:49:50.560] [nas] [info] DL AP Container type=0x0101 contentLength=183 pti=1 payloadId=0x1234 flags=0x00 fragmentOffset=0
[07:49:50.560] [nas] [info] DL AP Container payloadText={"intentId":"ueransim-test","issuer":"ueransim","intentPriority":1,"intentType":"location","intentDescription":"locate target","object":"ue","constraint":"","target":"amf-test"}
```

### 5.4 AMF 日志

```
[07:49:42.558] [INFO][AMF][Gmm] Handle UL Cooperation over 3GPP_ACCESS
[07:49:42.558] [INFO][AMF][Gmm] === UL Cooperation Message Details ===
[07:49:42.558] [INFO][AMF][Gmm] AP Container: type=0x0101 pti=0x01 payloadId=0x1234 DF=true MF=false offset=0 payloadLength=177
[07:49:42.558] [INFO][AMF][Gmm] [UL AP Container] Received: type=0x0101 pti=0x01 payloadId=0x1234 DF=true MF=false offset=0 payloadLength=177 payload={"intentId":"ueransim-test","issuer":"ueransim","intentPriority":1,"intentType":"location","intentDescription":"locate target","object":"ue","constraint":"","target":"amf-test"}
[07:49:42.558] [INFO][AMF][Gmm] [NAgent HTTP] Sending intent to NAgent: supi=imsi-001010000000001 payloadId=0x1234 pti=0x01 containerType=0x0101 accessType=3GPP_ACCESS payloadLength=177
[07:49:50.559] [INFO][AMF][Gmm] [NAgent HTTP] Received success response: payloadId=0x1234 responseLength=177
[07:49:50.559] [INFO][AMF][Gmm] [DL AP Container] Delivering NAgent response: payloadId=0x1234 responsePayloadLength=177
[07:49:50.559] [INFO][AMF][Gmm] [DL AP Container] Fragment: payloadId=0x1234 DF=false MF=false offset=0 payloadLength=177
[07:49:50.559] [INFO][AMF][Gmm] Sending DL Cooperation response
[07:49:50.559] [INFO][AMF][Gmm] Successfully built DLCooperation, sending via NGAP...
[07:49:50.559] [INFO][AMF][Ngap] Send Downlink Nas Transport
[07:49:50.559] [INFO][AMF][Gmm] DLCooperation sent successfully
```

### 5.5 时间线

| 时间 | 事件 |
|------|------|
| 07:49:42.335 | UE 开始注册 |
| 07:49:42.353 | UE 注册成功，发送 UL Cooperation (type=0xE1) |
| 07:49:42.558 | AMF 收到 UL Cooperation，解析 AP Container (payloadLength=177) |
| 07:49:42.558 | AMF 转发 intent 到 NAgent HTTP |
| 07:49:50.559 | NAgent Mock 响应返回 (responseLength=177) |
| 07:49:50.559 | AMF 构建 DL Cooperation，通过 NGAP 下发 |
| 07:49:50.560 | UE 收到 DL Cooperation (type=0xE2)，AP Container payload 完整匹配 |

**端到端延迟**：8.2 秒（NAgent Mock delayMs=8000）

### 5.6 验证点

- [x] UE 成功完成 5G-AKA 认证和注册
- [x] UE 发送 UL Cooperation (message type=0xE1)，包含 AP Container IE (IEI=0x71)
- [x] AMF 成功解码 UL Cooperation 消息（DecodeULCooperationV2）
- [x] AMF 解析 AP Container：type=0x0101, pti=0x01, payloadId=0x1234
- [x] AMF 提取 payload (177 bytes) 并通过 HTTP POST 转发到 NAgent
- [x] NAgent Mock 处理请求后返回响应 (responseLength=177)
- [x] AMF 构建 DL Cooperation (message type=0xE2)，封装 AP Container 响应
- [x] AMF 通过 NGAP DownlinkNasTransport 下发 DL Cooperation
- [x] UE 成功接收 DL Cooperation，解析 AP Container
- [x] DL AP Container payload 与 UL payload 内容完全一致

### 5.7 AP Container 字段对照

| 字段 | UL (UE 发送) | DL (UE 接收) |
|------|-------------|-------------|
| ContainerType | 0x0101 | 0x0101 |
| PTI | 0x01 (auto-generated) | 0x01 |
| PayloadId | 0x1234 | 0x1234 |
| DF (Don't Fragment) | true | false |
| MF (More Fragments) | false | false |
| FragmentOffset | 0 | 0 |
| PayloadLength | 177 | 177 |
| Payload | `{"intentId":"ueransim-test",...}` | `{"intentId":"ueransim-test",...}` |

---

## 6. gNB 日志

```
UERANSIM v3.3.0
[07:39:52.815] [sctp] [info] Trying to establish SCTP connection... (10.100.200.12:38412)
[07:39:52.817] [sctp] [info] SCTP connection established (10.100.200.12:38412)
[07:39:52.817] [ngap] [debug] Sending NG Setup Request
[07:39:52.819] [ngap] [debug] NG Setup Response received
[07:39:52.819] [ngap] [info] NG Setup procedure is successful
[07:49:42.335] [rrc] [debug] UE[4] new signal detected
[07:49:42.336] [rrc] [info] RRC Setup for UE[4]
[07:49:42.336] [ngap] [debug] Initial NAS message received from UE[4]
[07:49:42.352] [ngap] [debug] Initial Context Setup Request received
```

---

## 7. 两种模式对比

| 项目 | NAS Transport Passthrough | AP Container Cooperation |
|------|--------------------------|-------------------------|
| NAS message type (UL) | 0x67 (标准 UL NAS Transport) | 0xE1 (自定义 UL Cooperation) |
| NAS message type (DL) | 0x68 (标准 DL NAS Transport) | 0xE2 (自定义 DL Cooperation) |
| PayloadContainerType | 4 | 0x0101 (AP Container) |
| 内部格式 | 不透明 payload bytes | TLV IE list + AP Container header |
| AMF 是否解析 payload | 不解析 | 解析 TLV，提取 AP payload |
| HTTP request body | PayloadContainer 原始内容 | 重组后的 AP payload |
| DL payload 来源 | NAgent HTTP response body | NAgent HTTP response body 封入 AP Container |
| 是否标准 3GPP | 是（NAS Transport 是标准消息） | 否（Cooperation 0xE1/0xE2 为自定义扩展） |

---

## 8. 已知限制

1. **UPF 与 gNB 的 GTP 2152 端口冲突**：gNB 需要绑定 2152 端口，而 Docker 中的 UPF 容器映射了该端口。测试时需停止 UPF 容器（`docker stop upf`）释放端口。PDU Session 建立因此失败，但不影响 NAS Cooperation 测试。

2. **Docker SCTP NAT 问题**：通过 Docker 端口映射（`-p 38412:38412/sctp`）的 SCTP 连接会卡在 `COOKIE_WAIT` 状态。解决方案是让 gNB 直连 AMF 容器的 Docker 网络 IP（10.100.200.12）。

3. **Subscriber 数据**：需确保 MongoDB 中有与 UE 配置匹配的 subscriber 认证数据（K/OPC/AMF）。本测试使用的 subscriber 已通过 WebUI 注册。

4. **NAgent Mock 延迟**：当前 NAgent Mock 配置为 8 秒延迟（`delayMs: 8000`），实际部署时需替换为真实 NAgent 服务。

---

## 9. 测试结论

两种 NAS Cooperation 模式均端到端验证通过：

- **NAS Transport Passthrough**：标准 NAS Transport 消息（0x67/0x68）承载 PayloadContainerType=4，AMF 将 PayloadContainer 作为不透明字节透传给 NAgent，NAgent 响应通过 DL NAS Transport 回传 UE。payload 内容原样往返，完整一致。

- **AP Container Cooperation**：自定义 Cooperation 消息（0xE1/0xE2）承载 AP Container IE（IEI=0x71），AMF 解析 TLV IE list 并重组 AP Container，提取 payload 转发给 NAgent，NAgent 响应封装入 DL AP Container 后通过 DL Cooperation 回传 UE。AP Container 各字段（type/pti/payloadId/flags/offset）正确解析，payload 内容完整匹配。

测试环境信息：
- 测试日期：2026-07-30
- AMF commit：`b489a3d`
- AMF 镜像：`free5gc/amf:buildv16`
- UERANSIM 分支：`NewNAS-ApContaine`
