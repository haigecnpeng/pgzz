/*
开发者：海格通信---廖晓鹏
仪表查询与读写示例：
实际仪表操作中，查询与读写已可满足99%的业务需求。
实现方式很多，该示例采用调用VISA32.DLL动态函数库方式实现，应该是目前比较简单易用的实现方案。虽简单，但实际功效应该是一样的。
之前使用C.CString来传递参数（详见visa.go.bak），总感觉不太舒服，后更改为目前的方式。
API说明详见IO Libraries Suite自带文档Visa.chm
*/
package pgzzVISA

import (
	"fmt"
	"syscall"
	"unsafe"
)

// 新建类型，代替C.CString
// C.CString使用会产生拷贝，而且内存不会自动释放，需要进行free。
type MyString struct {
	// Str *C.char	//完全不需要使用CGO。
	Str unsafe.Pointer
	Len int
}

//类型转换
// \x00是必须的，go的string是由首字符指针+长度组成。而C的string只有首字符指针，长度由字节0来确定，即顺序读，直到读到0。\x00即代表字符0
func getMyString(s string) *MyString {
	s = s + "\x00"
	return (*MyString)(unsafe.Pointer(&s))
}

//设备管理器，用于管理设备。
var resourceManager int = 0

//ret如果不为0，则连接失败，ret为缺陷代码。
//打开设备管理器。类似初始化。
func OpenRM() uintptr {
	VISA32 := syscall.NewLazyDLL("visa32.dll")
	viOpenDefaultRM := VISA32.NewProc("viOpenDefaultRM")
	// viOpenDefaultRM(ViPSession sesn);
	ret, _, _ := viOpenDefaultRM.Call(uintptr(unsafe.Pointer(&resourceManager)))
	fmt.Println("硬件信息为：", resourceManager)
	return ret
}

//关闭设备管理器
func CloseRM() uintptr {
	VISA32 := syscall.NewLazyDLL("visa32.dll")
	viClose := VISA32.NewProc("viClose")
	// viClose(ViSession/ViEvent/ViFindList vi);
	ret, _, _ := viClose.Call(uintptr(resourceManager))
	return ret
}

//清除数据？
func ClearFindList() uintptr {
	VISA32 := syscall.NewLazyDLL("visa32.dll")
	viClear := VISA32.NewProc("viClear")
	// viClose(ViSession/ViEvent/ViFindList vi);
	ret, _, _ := viClear.Call(uintptr(resourceManager))
	return ret
}

// 查找仪表清单，并打印输出。
// 有时候刷新到旧的数据，看IO Libraries Suite也是一样的，断开后仍然有上一次的数据，只不过是错的。
// 关闭了仍然有信息。有可能需要调用viClear(ViSession vi); ---->   Clear a device. This operation performs an IEEE 488.1-style clear of the device.
// 错误代码详见Visa.chm
func FindRsrc() bool {
	VISA32 := syscall.NewLazyDLL("visa32.dll")
	// viFindRsrc(ViSession sesn, ViString expr, ViPFindList findList, ViPUInt32 retcnt, ViPRsrc instrDesc);
	viFindRsrc := VISA32.NewProc("viFindRsrc")
	var list int = 0
	ViString := getMyString("?*")
	var data [256]byte
	retcnt := 0
	ret, _, _ := viFindRsrc.Call(uintptr(resourceManager), uintptr(unsafe.Pointer(ViString.Str)), uintptr(unsafe.Pointer(&list)), uintptr(unsafe.Pointer(&retcnt)), uintptr(unsafe.Pointer(&data)))
	if ret != 0 {
		fmt.Println("查询错误代码：", ret)
		return false
	}
	fmt.Println("查询成功->")
	viFindNext := VISA32.NewProc("viFindNext")
	// viFindNext(ViFindList findList, ViPRsrc instrDesc);
	viFindNext.Call(uintptr(list), uintptr(unsafe.Pointer(&data)))
	s := string(Bytes2string(data))
	viClose := VISA32.NewProc("viClose")
	// viClose(ViSession/ViEvent/ViFindList vi);
	ret2, _, _ := viClose.Call(uintptr(list))
	if ret2 != 0 {
		fmt.Println("查询错误代码：", ret)
		return false
	}
	fmt.Println(s)
	return true
}

// 发送信息给仪表
// addr为仪器地址，m为需要发送的信息，如GPIB0::15::INSTR，send:DISP TX
func SendMsg(addr, m string) string {
	rsrcName := getMyString(addr)
	VISA32 := syscall.NewLazyDLL("visa32.dll")
	viOpen := VISA32.NewProc("viOpen")
	var vi int = 0
	// viOpen(ViSession sesn, ViRsrc rsrcName, ViAccessMode accessMode, ViUInt32 timeout, ViPSession vi);
	ret, _, _ := viOpen.Call(uintptr(resourceManager), uintptr(unsafe.Pointer(rsrcName.Str)), uintptr(0), uintptr(0), uintptr(unsafe.Pointer(&vi)))
	if vi == 0 {
		return "打开仪器失败，错误代码为：" + fmt.Sprint(ret)
	}
	m = m + "\n"
	msg := getMyString(m)
	viPrintf := VISA32.NewProc("viPrintf")
	// viPrintf(ViSession vi, ViString writeFmt, arg1, arg2,...);
	ret2, _, _ := viPrintf.Call(uintptr(vi), uintptr(unsafe.Pointer(msg.Str)))
	viClose := VISA32.NewProc("viClose")
	viClose.Call(uintptr(vi))
	if ret2 == 0 {
		return "信息发送成功！---" + m
	}
	return "发送失败，代码为：" + fmt.Sprint(ret2)
}

// 发送信息给仪表，并等待仪表返回数据。
func ReadData(addr, m string) string {
	instrDor := getMyString(addr)
	VISA32 := syscall.NewLazyDLL("visa32.dll")
	viOpen := VISA32.NewProc("viOpen")
	var vi int = 0
	// viOpen(ViSession sesn, ViRsrc rsrcName, ViAccessMode accessMode, ViUInt32 timeout, ViPSession vi);
	ret, _, _ := viOpen.Call(uintptr(resourceManager), uintptr(unsafe.Pointer(instrDor.Str)), uintptr(0), uintptr(0), uintptr(unsafe.Pointer(&vi)))
	if vi == 0 {
		return "打开仪器失败，错误代码为：" + fmt.Sprint(ret)
	}
	m = m + "\n"
	msg := getMyString(m)
	viPrintf := VISA32.NewProc("viPrintf")
	// viPrintf(ViSession vi, ViString writeFmt, arg1, arg2,...);
	viPrintf.Call(uintptr(vi), uintptr(unsafe.Pointer(msg.Str)))
	readFmt := getMyString("%t")
	viScanf := VISA32.NewProc("viScanf")
	var data [256]byte
	// viScanf(ViSession vi, ViString readFmt, arg1, arg2,...);
	viScanf.Call(uintptr(vi), uintptr(unsafe.Pointer(readFmt.Str)), uintptr(unsafe.Pointer(&data)))
	s := string(Bytes2string(data))
	viClose := VISA32.NewProc("viClose")
	viClose.Call(uintptr(vi))
	return s
}

// 返回的数据后面有很多个0，截断处理会更好些。
func Bytes2string(data [256]byte) string {
	for i := 0; i < len(data); i++ {
		if data[i] == 10 || data[i] == 0 {
			return string(data[:i])
		}
	}
	return string(data[:])
}
