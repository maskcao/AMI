新威胁：ZHtrap僵尸网络分析报告

- [https://blog.netlab.360.com](https://blog.netlab.360.com)

[Botnet](https://blog.netlab.360.com/tag/botnet/)

- [DNSMon](https://blog.netlab.360.com/tag/dnsmon/)

- [DDoS](https://blog.netlab.360.com/tag/ddos/)

- [PassiveDNS](https://blog.netlab.360.com/tag/pdns/)

- [Mirai](https://blog.netlab.360.com/tag/mirai/)

- [DTA](https://blog.netlab.360.com/tag/dta/)

[New Threat](/tag/new-threat/)

# 新威胁：ZHtrap僵尸网络分析报告

- [/author/alex/](/author/alex/)

- [/author/liuyang/](/author/liuyang/)

- [/author/xuy1202/](/author/xuy1202/)

#### Alex.Turing ,  liuyang ,  YANG XU

12 Mar 2021

&bull;  15 min read

[Share](#/share)

## 版权

版权声明：本文为Netlab原创，依据 [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/?ref=blog.netlab.360.com) 许可证进行授权，转载请附上出处链接及本声明。

## 概述

从2021年2月28日起，360网络安全研究院的BotMon系统检测到IP( 107.189.30.190 )在持续传播一批未知ELF样本。经分析，我们确认这些样本隶属于一个新的botnet家族，结合其运行特点，我们将其命名为 ZHtrap ，本文对其做一分析，文章要点如下：

- ZHtrap的传播使用了4个Nday漏洞，主要功能依然是DDoS和扫描，同时集成了一些后门功能。

- Zhtrap能将被感染设备蜜罐化 ，目的是提高扫描效率。

- Zhtrap感染受害主机后会禁止运行新的命令，以此实现彻底控制和独占该设备。

- 在C2通信上，ZHtrap借鉴了[套娃](https://blog.netlab.360.com/matryosh-botnet-is-spreading/)，采用了Tor和云端配置。

## ZHtrap全情介绍

ZHtrap的代码由Mirai修改而来，支持x86, ARM, MIPS等主流CPU架构。但相对Mirai，ZHtrap变化较大，体现在如下方面：

- 在指令方面，加入了校验机制

- 在扫描传播方面， 增加了对真实设备和蜜罐的区分

- 在加密算法方面，重新设计了一套多重XOR加密算法

- 在主机行为方面， 能将被攻陷设备变成一个简易蜜罐，并实现了一套进程管控机制

- 在网络架构方面，吸收了我们之前曝光的套娃僵尸网络的实现

基本流程如下图所示：

![](/api/v1/intel/52/raw-assets/5ad2f3b58ed04484.png)

在功能方面，除了DDoS攻击和扫描，ZHtrap还实现了后门功能，这增加了该家族的危害性。ZHtrap的具体功能包括：

- 进程管控

- 反弹Shell

- DDoS攻击

- Telnet扫描

- Exploit传播

- 蜜罐化被感染设备

- 下载执行Payload

目前捕获的ZHtrap样本，依据其功能可以分成3个版本：v1，v2和v3，其中：v2在v1的基础上，加入了漏洞利用的功能; v3在v2的基础上对网络基础设施做了删减。它们的关系如下图所示，本文的分析是基于功能最全的的v2版本。

![](/api/v1/intel/52/raw-assets/c295a95626a25314.png)

跟我们之前分析过的botnet相比，ZHtrap最大的不同是能 将被感染设备蜜罐化 。蜜罐通常是安全研究人员用来捕获攻击的工具，比如收集扫描、exploits和样本，有趣的是，我们发现ZHtrap也使用了类似的技术，集成了一个扫描IP收集模块，收集到的IP会被用到自身的扫描模块中，基本流程如下所示：

![](/api/v1/intel/52/raw-assets/0e3bfbf660417a95.png)

实际工作时，ZHtrap会监听23个指定端口（如上图所示），如果发现有IP连接这些端口就将它当成scanner IP记录下来，这些记录的IP会被用到它自身的扫描模块中，这样，ZHtrap扫描过程中所用到的目标地址就有2个来源：

- 随机生成的IP；

- 上述模块捕获的扫描IP。

这方面更多的细节可以参考下面的样本分析部分。

## 样本逆向分析

本文选取ZHtrap v2的X86 CPU架构样本做为分析对象，基本信息如下：

MD5：5370e0b9484cb25fb3d5a4b648b5c203

ELF 32-bit LSB executable, Intel 80386, version 1 (SYSV), statically linked, stripped

Packer：None

ZHtrap成功侵入设备后，首先通过绑定本地端口实现单一实例，然后解密出加密的资源信息，接下来将进程重命名为 /bin/ZoneSec ，最后在Console输出

ZH infected your shit, message me on discord to compare assets ($reiko#0095), researchers hmu twitter[.]com/ZoneHax “调戏”安全研究人员。接着通过网络请求得到资源服务器的地址，以供扫描&amp;传播阶段使用。随后监听23个端口将设备变成一个&quot;蜜罐&quot;，通过telnet弱口令扫描收集可被入侵的设备信息，通过4个Nday漏洞扫描实现蠕虫式传播。最终和Tor C2进行通信，等待执行C2下发的指令。下面逐一剖析上述功能的技术细节。

### 0x0 加密算法

ZHtrap使用了一个比较少见的多重xor加密算法，来隐藏资源信息。

![](/api/v1/intel/52/raw-assets/3a92c845dd90bdd3.png)

等效的python实现如下所示：

xor_key_lst = [0x51, 0x6A, 0x66, 0x6A, 0x78, 0x53, 0x52, 0x44, 0x46, 0x47, 0x53, 0x46, 0x44, 0x64, 0x66] #QjfjxSRDFGSFDdf

xor_key = 0xD00DBAAF

def decode(content):

desc_lst = []

plaintext=&quot;&quot;

xor_key_lst_idx = 1

for idx in range(0, len(content)):

desc_lst.append(chr(content[idx] ^ (xor_key % 0xFF) ^ xor_key_lst[xor_key_lst_idx - 1] ^ (xor_key_lst_idx + idx)))

xor_key_lst_idx += 1

if xor_key_lst_idx &gt;= len(xor_key_lst):

xor_key_lst_idx = 0

else:

print(&quot;&quot;.join(desc_lst))

以下面密文为例

cipher =[

0x42, 0x69, 0x0B, 0x4C, 0x57, 0x76, 0x72, 0x60, 0x6B, 0x79,

0x6A, 0x39, 0x6C, 0x58, 0x55, 0x7B, 0x10, 0x49, 0x5C, 0x41,

0x75, 0x2A, 0x32, 0x43, 0x48, 0x4C, 0x5B, 0x45, 0x61, 0x56,

0x26, 0x6E, 0x68, 0x27, 0x78, 0x5C, 0x11, 0x45, 0x48, 0x4D,

0x4B, 0x54, 0x49, 0x71, 0x22, 0x43, 0x7D, 0x3C, 0x75, 0x69,

0x4E, 0x50, 0x51, 0x42, 0x2A, 0x79, 0x2B, 0x39, 0x17, 0x70,

0x50, 0x6E, 0x4F, 0x49, 0x51, 0x2E, 0x36, 0x2E, 0x28, 0x2F,

0x69, 0x6D, 0x69, 0x42, 0x51, 0x7C, 0x40, 0x5E, 0x02, 0x01,

0x3E, 0x27, 0x37, 0x20, 0x32, 0x13, 0x09, 0x1A, 0x39, 0x57,

0x2A, 0x12, 0x04, 0x63, 0x27, 0x09, 0x14, 0x11, 0x11, 0x07,

0x06, 0x51, 0x1C, 0x36, 0x2B, 0x5C, 0x14, 0x2F, 0x3C, 0x27,

0x27, 0x0D, 0x0C

]

解密后得到如下明文，正是输出在Console的提示信息：

ZH infected your shit, message me on discord to compare assets ($reiko#0095), researchers hmu twitter[.]com/ZoneHax

当前样本中一共有3条加密信息，解密后如下所示，其中 /proc/ 以及 /stat 在下一章节 进程管控 中使用。

index

plaintext

2

/stat

1

/proc/

0

ZH infected your shit.....

### 0x1 进程管控

ZHtrap通过白名单和快照机制实现进程管控，以实现对设备的独占。其中可执行文件包含以下路径的进程即被视为白名单进程。ZHtrap启动后会先获取当前的进程列表，然后通过kill -9结束非白名单进程。从这一步可以看出，ZHtrap并不想破坏系统的正常运行。

/bin

/sbin

/user/bin

/user/sbin

接下来ZHtrap会为系统建立进程快照，此后如果有新创建进程，如果不在快照中将被kill。如此一来，整个系统就保持在ZHtrap的控制下运行了，即使管理员发现设备有问题想通过系统工具进行管理，也无法正常执行，远程维护变成了一项不可能的任务。

第一步，清理非白名单进程

ZHtrap运行时，遍历系统当前进程，读取&quot;/proc/pid/stat&quot;中第22项， starttime 的值。这个值除以_SC_CLK_TCK可以得到进程是在内核启动后多久启动的。 _SC_CLK_TCK 的值一般等于100，10000/_SC_CLK_TCK的结果为100秒，即所有在内核启动100秒之后启动的进程都要进行检查，如果进程路径不在白名单中，则通过kill杀掉进程。

![](/api/v1/intel/52/raw-assets/287b189d448538d9.png)

第二步，建立进程快照，不在标的新进程都将被清理

清理完非白单名进程后，ZHtrap认为系统处在一个“纯净的状态”，首先为系统建立进程快照。

![](/api/v1/intel/52/raw-assets/b51048a865e97ebc.png)

在此之后新创建的进程都会和进程快照进行比较，不符合要求的进程，都将被直接kill。

![](/api/v1/intel/52/raw-assets/f74a5d6a28b32d22.png)

### 0x2 获取资源服务器

ZHtrap通过以下代码片段获得资源服务器(ResServ)的地址，这个过程可以分成俩步，第一步和Tor domain建立通信，第二步发送&quot; GET /sfkjdkfdj.txt &quot;请求。

![](/api/v1/intel/52/raw-assets/56a89fc2e593082c.png)

如何和Tor domain建立通信呢？

ZHtrap借鉴了套娃僵尸网络的思路，首先向远程主机 0xdeadbeef.tw 请求DNS TXT记录，得到下图中的Tor代理的列表，

![](/api/v1/intel/52/raw-assets/245fb58045efd2ad.png)

任选其中一个代理建立连接，然后向代理发送想要建立通信的 domain :port 信息，如果代理返回 05 00 00 01 00 00 00 00 00 00 时，说明通信已成功建立。然后发送后续的GET请求就能得到资源服务器地址( 107.189.30.190 )，下图的网络流量很清楚的反映这个过程。

![](/api/v1/intel/52/raw-assets/c07b92d0dbfb7ea2.png)

资源服务器有什么用呢？

![](/api/v1/intel/52/raw-assets/44a16d36c25e9bb9.png)

通过上图IDA交叉引用可知，ResServ的功能有俩个：

- 充当Reporter Server角色，为Telnet扫描过程提供信息上传服务

- 充当Downloader Server角色，为漏洞传播过程提供样本下载服务

### 0x3 蜜罐化被攻陷的设备

ZHtrap首先建立了2个管道用于信息传输，然后通过以下代码片段实现了一个监听23个端口的简易蜜罐

![](/api/v1/intel/52/raw-assets/b4fa5975c6eca565.png)

实际效果如下：

![](/api/v1/intel/52/raw-assets/0e7e3c44a9d24452.png)

当设备的这些端口被访问时，ZHtrap会将访问者的IP地址通过管道传递给telnet扫描&amp;漏洞扫描使用。

![](/api/v1/intel/52/raw-assets/14c5ad2555d98992.png)

为什么要这样做呢？

我们推测，ZHtrap的作者的想法是这样的：

许多botnet都会实现蠕虫式的扫描传播，当ZHtrap的蜜罐端口被访问时，其来源很有可能是已被别家botnet感染的设备。这设备能被感染，肯定存在缺陷，我使用我的扫描机制反扫一次有很大的机会能植入我的bot样本，再配合进程管控功能，就能独占设备了。Awesome！

### 0x4 Telnet扫描

ZHtrap使用Telnet扫描来收集使用弱口令的设备，扫描对象来源有以下2类：

- 主动随机生成的IP

- 被动接收由蜜罐投递的IP

不同来源的IP，有不同的处理办法:

对于第1类IP，使用SYN端口探测。

![](/api/v1/intel/52/raw-assets/2d4f9da791152e14.png)

对于第2类IP，直接调用connect函数。

![](/api/v1/intel/52/raw-assets/d66c3d2090d8b2e4.png)

当发现目标的23端口是开放时，使用硬编码的凭证列表尝试登录，将能成功登录的的 IP，端口，帐号，密码 等信息通过以下代码，回传给资源服务器。

![](/api/v1/intel/52/raw-assets/7efb08acd3209d4c.png)

在尝试登录这个过程中，ZHtrap会让被扫描设备执行以下命令:

enable

linuxshell

system

bash

ls /home

ps aux

/bin/busybox ZONESEC

然后根据返回的信息对设备类型进行判断，当包含以下字串时，说明设备是个蜜罐。

string

honeypot

Jun22

[cowrie](https://github.com/cowrie/cowrie?ref=blog.netlab.360.com)

Jun23

cowrie

phil

cowrie

sshd:

cowrie

richard

cowrie

@LocalHost:]

cowrie

Welcome to EmbyLinux 3.13.0-24-generic

[telnet-iot-honeypot](https://github.com/Phype/telnet-iot-honeypot?ref=blog.netlab.360.com)

如果是蜜罐，则将设备信息通过 2231端口 上报到资源服务器。

![](/api/v1/intel/52/raw-assets/bd50ecf1a64a8fb4.png)

如果是真实设备，则将设备信息通过 1282端口 上报给资源服务器。

![](/api/v1/intel/52/raw-assets/e306ba00691f9d22.png)

### 0x5 漏洞扫描&amp;传播

ZHtrap使用以下4个Nday漏洞传播样本：

- [JAWS_DVR_RCE](https://www.exploit-db.com/exploits/41471/?ref=blog.netlab.360.com)

- [NETGEAR](https://www.exploit-db.com/exploits/43055?ref=blog.netlab.360.com)

- [CCTV_DVR_RCE](https://www.exploit-db.com/exploits/39596/?ref=blog.netlab.360.com)

- [CVE-2014-8361](https://www.exploit-db.com/exploits/37169/?ref=blog.netlab.360.com)

首先构造SYN包探测设备的端口是否开放，支持的端口(exp_port)列表如下所示：

80

8080

8081

8083

5500

60001

52869

8089

8090

8000

81

82

83

84

85

8888

8181

8443

5555

然后发送“GET /”请求以获取端口所运行服务的信息：

![](/api/v1/intel/52/raw-assets/5cc7f72c7c29e0c1.png)

随后通过以下代码片段对服务返回的banner信息做判断，进而判断是否为目标设备。

![](/api/v1/intel/52/raw-assets/48d507a34f1034fc.png)

最终根据banner信息，选择相应的漏洞，和ResServ组装有效payload，尝试植入。

![](/api/v1/intel/52/raw-assets/97dd327e3ed7f89e.png)

和telnet扫描过程一样，扫描对象来源有以下2类：

- 主动随机生成的IP

- 被动接收由蜜罐投递的IP

对于第一类IP，随机选择一个exp_port进行探测，IP与exp_port的关系是1对1。

对于第二类IP，对全部的exp_port进行探测，IP与exp_port的关系是1对N，下图中的网络流量很直观的说明了这种情况。

![](/api/v1/intel/52/raw-assets/0e3058324ff2c2ad.png)

### 0x6 C2通信

ZHtrap使用Tor C2，因此必须要在代理的帮助下，才能和C2通信。这个过程在前文有过说明，此处就不再赘述 ，下图所示的网络流量说明和C2已成功的建立了连接。

![](/api/v1/intel/52/raw-assets/8b4241f2c3fc4b7a.png)

接着通过以下代码片段向C2发送上线信息：

![](/api/v1/intel/52/raw-assets/08e8d78d94e91b61.png)

实际中产生的网络流量如下：

![](/api/v1/intel/52/raw-assets/57c033b76e531e3f.png)

网络流量又是如何解析呢？ZHtrap的指令包由2个数据包组成，第一个包为header，第二个包为body，其中header的格式为：

//5 bytes ,big endian

struct header

{

uint16	body_len;

uint8	cmd_type;

uint16	checksum;

}

以上面的上线包流量为例，各字段含义如下所示：

header:

00 07        				-- length of body

e5			 				-- hardcode,cmd type,&quot;e5&quot;,

1a fa		 				-- tcp/ip checksum of (&quot;00 07 e5 00 00&quot;)

body:

5a 6f 6e 65 53 65 53		-- body string,&quot;ZoneSec&quot;

发送完上线包后Bot开始等待C2下发指令，当指令包的header成功通过校验之后，依据header中第3个字节指定的命令类型，选择相应的处理流程。

![](/api/v1/intel/52/raw-assets/f547870ba438188c.png)

可以看出目前一共支持5种指令，指令码与功能的对应关系如下表所示：

Value

Function

0x22

heartbeat

0x45

exit

0x89

download&amp;exec payload

0xF5

reverse shell

0xFE

DDoS

以心跳为例，下图中的心跳包网络流程验证了我们的分析。

![](/api/v1/intel/52/raw-assets/73eb1d5f23800df5.png)

## 总结

在后Mirai时代，我们见识过黑产从业者们或大或小的脑洞，其中ZHtrap将被攻陷设备蜜罐化以收集潜在可被攻击目标的做法，无疑是相当有新意的。扫描是攻击的前置过程，扫描的效率决定了攻击的质量。随机扫描是目前的主流，但它有一个致命的缺点：动静太大，极易被监测。ZHtrap的蜜罐反扫机制略显粗糙，但从单纯的随机扫描到有方向性的捕获这一转变是有非常重要的意义，如果其作者持续对此投入研究，很有可能让攻击行为从安全研究员的雷达消失。

## 联系我们

感兴趣的读者，可以在 [twitter](https://twitter.com/360Netlab?ref=blog.netlab.360.com) 或者通过邮件 netlab[at]360.cn 联系我们。

## IOC

Sample MD5

5370e0b9484cb25fb3d5a4b648b5c203

6c7cfbe0277e2ca0cbe7157cad7c663e

f1f70dc1274112ae5287ceb06f096d0e

9dded61f7de47409bc00e74c0a12210e

7b593fbbd6f81a3e9a2043a46949879d

ba17282481acca9636c5d01f5c2dd069

URL

0xdeadbeef.tw

h5vwy6o32sdcsa5xurde35dqw5sf3cdsoeewqqxmhoyzsvar4u6ooead.onion:8080

C2

oemojwe5loscudytzfo273nkdvalf7mumctwcm42zyutoo6tpfjsphyd.onion:3000

Reporter

107.189.30.190:1282

107.189.30.190:2231

Proxy Ip

51.178.54.234:9095

51.79.157.89:9095

167.114.185.33:9095

147.135.208.44:9095

198.245.53.58:9095

142.93.247.244:9050

66.70.188.235:9095

139.99.134.95:9095

144.217.243.21:9095

46.101.61.9:9050

Downloader

x86 arm5 arm6 arm7 mips

hxxp://107.189.30.190/bins/z.{CPU_ARCH}

oemojwe5loscudytzfo273nkdvalf7mumctwcm42zyutoo6tpfjsphyd.onion:8080/z.{CPU_ARCH}

[## 僵尸网络911 S5的数字遗产

概述

2024年5月29日，美国司法部发布通告，声称其执法活动摧毁了&quot;史上最大的僵尸网络&quot; 911 S5，查封了相关域名，并且逮捕了其管理员YunHe Wang。Wang及其同伙通过创建并分发包含恶意代码的免费VPN程序感染用户，并且在名为911 S5的住宅代理服务中出售对被感染设备构成的代理网络的访问权。

按照360威胁情报中心的分析，911S5从2014年开始运营，到2022年7月关停，在2023年10月又摇身一变，化名CloudRouter继续其肮脏生意，终于在2024年5月被多国联合执法摧毁。911S5的僵尸网络运行时间长、涉及多个国家的19M个IP地址、行为高调，虽然经过执法行动后大势已去，但是其数字遗产仍然对网络空间构成了现实且显著的威胁，下文是我们对威胁分析的结果。

“空手套白狼”的911 S5

911S5出售的代理服务背后是数千万被感染的设备。受害者主动或被动下载捆绑了恶意代码的软件、免费VPN程序等。在程序启动后，恶意代码将会创建持久化服务作为后门，为911S5客户提供代理服务。

在2023年以前，911S5使用的免费VPN包括:ProxyGate、Mas](/911s5/)

14 Jun 2024

7 min read

[## Heads up! Xdr33, A Variant Of CIA’s HIVE  Attack Kit Emerges

Overview

On Oct 21, 2022, 360Netlab&#39;s honeypot system captured a suspicious ELF file ee07a74d12c0bb3594965b51d0e45b6f, which propagated via F5 vulnerability with zero VT detection, our system observces that it communicates with IP 45.9.150.144 using SSL with forged Kaspersky certificates, this caught our attention. After further lookup,](/headsup_xdr33_variant_of_ciahive_emeerges/)

10 Jan 2023

13 min read

[## 警惕：魔改后的CIA攻击套件Hive进入黑灰产领域

概述

2022年10月21日，360Netlab的蜜罐系统捕获了一个通过F5漏洞传播，VT 0检测的可疑ELF文件ee07a74d12c0bb3594965b51d0e45b6f，流量监控系统提示它和IP45.9.150.144产生了SSL流量，而且双方都使用了伪造的Kaspersky证书，这引起了我们的关注。经过分析，我们确认它由CIA被泄露的Hive项目server源码改编而来。这是我们首次捕获到在野的CIA HIVE攻击套件变种，基于其内嵌Bot端证书的CN=xdr33， 我们内部将其命名为xdr33。关于CIA的Hive项目，互联网中有大量的源码分析的文章，读者可自行参阅，此处不再展开。

概括来说，xdr33是一个脱胎于CIA Hive项目的后门木马，主要目的是收集敏感信息，为后续的入侵提供立足点。从网络通信来看，xdr33使用XTEA或AES算法对原始流量进行加密，并采用开启了Client-Certificate Authentication模式的SSL对流量做进一步的保护；从功能来说，主要有beacon，trigger两大任务，其中beacon是周期性向硬编码的Be](/warning-hive-variant-xdr33-is-coming_cn/)

09 Jan 2023

17 min read

[360 Netlab Blog - Network Security Research Lab at 360](https://blog.netlab.360.com) &copy; 2026

- [Sign up](#/portal/)

[Powered by Ghost](https://ghost.org/)