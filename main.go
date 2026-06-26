package main

import (
	"fmt"
	"net/http"
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>端口连通测试工具</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:Arial;margin-top:80px;background:#f5f7fa;text-align:center}
.container{width:780px;margin:0 auto;background:#fff;padding:40px;border-radius:12px;box-shadow:0 2px 15px #ddd}
.info{margin:15px 0;color:#444}
.tip{color:#999;font-size:14px;margin:8px 0 25px}
.btn-group{margin-bottom:20px}
.btn-group button{margin:8px 7px;padding:10px 14px;background:#2d8cf0;color:#fff;border:none;border-radius:6px;font-size:13px;cursor:pointer}
.btn-group button:hover{background:#1b76d8}
.result-box{margin-top:30px;text-align:left;display:inline-block}
.success{color:#00b42a;line-height:1.8}
.fail{color:#f53f3f;line-height:1.8}
.running{color:#ff7d00;line-height:1.8}
</style>
</head>
<body>
<div class="container">
    <h1>✅ 端口连通检测工具</h1>
    <p class="info">检测目标服务器：<span id="hostText"></span></p>
    <p class="tip">检测逻辑：当前浏览器所在电脑 ↔ 服务器端口建立TCP连接验证</p>

    <div class="btn-group">
        <button onclick="singleTest(80)">单独测试 80</button>
        <button onclick="singleTest(443)">单独测试 443</button>
        <button onclick="singleTest(8080)">单独测试 8080</button>
        <button onclick="batchAllTest()">一键批量全部测试</button>
    </div>

    <div id="resultBox" class="result-box"></div>
</div>
<script>
const serverHost = location.hostname;
document.getElementById('hostText').innerText = serverHost;
const testPortList = [80, 443, 8080];

function singleTest(port) {
    const resBox = document.getElementById('resultBox');
    resBox.innerHTML = "<div class='running'>浏览器正在访问 " + serverHost + ":" + port + "，建立TCP连接中...</div>";
    const abortCtrl = new AbortController();
    const timer = setTimeout(function(){
        abortCtrl.abort();
        resBox.innerHTML = "<div class='fail'>❌ " + port + " 端口：连接超时，被安全组/防火墙拦截</div>";
    }, 2500);

    fetch("http://" + serverHost + ":" + port + "/", {
        signal: abortCtrl.signal,
        mode: 'no-cors'
    }).then(function(){
        clearTimeout(timer);
        resBox.innerHTML = "<div class='success'>✅ " + port + " 端口：访问正常，链路通畅</div>";
    }).catch(function(err){
        clearTimeout(timer);
        if(err.name !== 'AbortError'){
            resBox.innerHTML = "<div class='fail'>❌ " + port + " 端口：服务器拒绝连接，端口未放行</div>";
        }
    });
}

function batchAllTest() {
    const resBox = document.getElementById('resultBox');
    resBox.innerHTML = "<div class='running'>开始批量检测，浏览器依次访问所有端口...</div>";
    let resultHtml = '';
    let index = 0;

    function nextCheck(){
        if(index >= testPortList.length){
            return;
        }
        const port = testPortList[index];
        index++;
        resultHtml += "<div class='running'>正在访问 " + serverHost + ":" + port + "</div>";
        resBox.innerHTML = resultHtml;
        const abortCtrl = new AbortController();
        const timer = setTimeout(function(){
            abortCtrl.abort();
            resultHtml += "<div class='fail'>❌ " + port + " 端口：连接超时被拦截</div>";
            resBox.innerHTML = resultHtml;
            setTimeout(nextCheck, 300);
        }, 2500);

        fetch("http://" + serverHost + ":" + port + "/", {
            signal: abortCtrl.signal,
            mode: 'no-cors'
        }).then(function(){
            clearTimeout(timer);
            resultHtml += "<div class='success'>✅ " + port + " 端口：连通正常</div>";
            resBox.innerHTML = resultHtml;
            setTimeout(nextCheck, 300);
        }).catch(function(err){
            clearTimeout(timer);
            if(err.name !== 'AbortError'){
                resultHtml += "<div class='fail'>❌ " + port + " 端口：连接被拒绝</div>";
            }
            resBox.innerHTML = resultHtml;
            setTimeout(nextCheck, 300);
        });
    }
    nextCheck();
}
</script>
</body>
</html>
`

type handlerWithPort struct {
	listenPort string
}

func (h *handlerWithPort) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	fmt.Printf("【访问日志】客户端IP:%s 访问本机端口:%s\n", clientIP, h.listenPort)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(htmlTemplate))
}

func startListen(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", (&handlerWithPort{listenPort: port}).ServeHTTP)
	server := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: mux,
	}
	fmt.Printf("[监听成功] 0.0.0.0:%s\n", port)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("[监听失败] %s 端口: %v\n", port, err)
	}
}

func main() {
	ports := []string{"80", "443", "8080"}
	for _, p := range ports {
		go startListen(p)
	}
	select {}
}
