package main

import (
	"bufio"
	"fmt"
	"os"
	"pgzzVISA"
	"strings"
)

// GPIB0::15::INSTR->send:DISP TX
func getMsg(s string) (addr, msg string) {
	if strings.Contains(s, "->") {
		temp := strings.Split(s, "->")
		addr = temp[0]
		msg = temp[1]
		return addr, msg
	}
	return "", ""
}

func CtlVISA(strMsg string) string {
	addr, msg := getMsg(strMsg)
	if addr == "" {
		return "接收到错误信息：" + strMsg
	}
	if strings.Contains(msg, "read:") {
		msg := strings.TrimLeft(msg, "read:")
		return pgzzVISA.ReadData(addr, msg)
	}
	if strings.Contains(strMsg, "send:") {
		msg := strings.TrimLeft(msg, "send:")
		return pgzzVISA.SendMsg(addr, msg)
	}
	return "格式错误：" + strMsg
}

func main() {
	//初始化硬件，记得退出时关闭。
	ret := pgzzVISA.OpenRM()
	//找不到硬件就退出，显示错误代码，代码对应含义请查看文档，注意，要先转换为十六进制。
	if int(ret) < 0 {
		fmt.Println("未找到硬件，错误代码为：", ret)
		return
	}
	if !pgzzVISA.FindRsrc() {
		return
	}
	fmt.Println("已打开VISA通道")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		strMsg := scanner.Text()
		if strMsg == "quit" {
			break //退出程序
		}
		if strMsg == "find" {
			pgzzVISA.FindRsrc()
			continue //退出当前循环序列，继续下一循环
		}
		fmt.Println(CtlVISA(strMsg))
	}
	fmt.Println("正在退出，请稍候……")
	pgzzVISA.CloseRM()
}

// func main() {
// 	var addr string
// 	if len(os.Args) > 1 {
// 		addr = os.Args[1]
// 		fmt.Println("接收到仪器地址：", addr)
// 	} else {
// 		fmt.Println("仪器地址通过参数传递！")
// 		return
// 	}
// 	ret := pgzzVISA.OpenRM()
// 	//找不到硬件就退出
// 	if int(ret) < 0 {
// 		fmt.Println("未找到硬件，错误代码为：", ret)
// 		return
// 	}
// 	// fmt.Println("硬件信息为：", pgzzVISA.ResourceManager)
// 	fmt.Println("发送指令信息：w>...\n发送接收信息：r>...")
// 	fmt.Println("退出发送quit")
// 	scanner := bufio.NewScanner(os.Stdin)
// 	for scanner.Scan() {
// 		// fmt.Scanln(&strMsg)
// 		// fmt.Println(strMsg)
// 		strMsg := scanner.Text()
// 		if strings.Contains(strMsg, "r>") {
// 			msg := strings.TrimLeft(strMsg, "r>")
// 			pgzzVISA.ReadData(addr, msg)
// 			continue
// 		}
// 		if strings.Contains(strMsg, "w>") {
// 			msg := strings.TrimLeft(strMsg, "w>")
// 			pgzzVISA.SendMsg(addr, msg)
// 			continue
// 		}
// 		if strMsg == "quit" {
// 			break
// 		}
// 		fmt.Println("格式错误：", strMsg)
// 		fmt.Println("发送指令信息：w->...\n发送接收信息：r->...")
// 	}
// 	fmt.Println("退出")
// 	pgzzVISA.CloseRM()
// }
