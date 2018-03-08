package  main

import (
	"flag"
	//"github.com/tarm/goserial"
	"log"
	"io"
	"strings"
	"net"
	"fmt"
	"bufio"
	"reflect"
	"strconv"
	"time"
	"github.com/tarm/goserial"
	"sync"
	"runtime"
)

//var unitNum int = 1

var (
	conFile = flag.String("configfile","/config.ini","config file")
)

var TOPIC = make(map[string]string)

func receiveCom(s io.ReadWriteCloser,unitNum int,conn net.Conn)( code int ,err error){
	buf := make([]byte,128)
	for{
		n, err := s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		strResult := string(buf[:n])
		if(isEmpty(strResult)){
			return 10001,nil
		}
		//strResult = strings.Replace(strResult, "\r", "", -1)
		strResult = strings.Replace(strResult, "\n", "", -1)
		var tcp_msg =  strconv.Itoa(unitNum) + "$"
		tcp_msg = tcp_msg + strResult
		log.Println(tcp_msg)
		send_tcp(conn,tcp_msg)
		if strings.Contains(strResult,"OK"){
			break
		}
		if strings.Contains(strResult,"+CME ERROR: SIM not inserted"){
			break
		}
		if strings.Contains(strResult,"ERROR"){
			break
		}
	}
	return 10000,nil
}

func receiveAtPingCom(s io.ReadWriteCloser,unitNum int,conn net.Conn)( code string ,err error){
	buf := make([]byte,128)
	//var PingText string
	var PingText string
	PingText = ""
	for{
		n, err := s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		strResult := string(buf[:n])
		if(isEmpty(strResult)){
			return "error",nil
		}
		PingText += strResult
		strResult = strings.Replace(strResult, "\r", "", -1)
		//strResult = strings.Replace(strResult, "\n", "", -1)
		var tcp_msg =  strconv.Itoa(unitNum) + "$"
		tcp_msg +=  strResult
		//log.Println(tcp_msg)

		send_tcp(conn,tcp_msg)
		if strings.Contains(strResult,"+CPING: 3,"){
			break
		}
	}
	//log.Println(PingText)
	return PingText,nil
}

func receiveAtNetOpenCom(s io.ReadWriteCloser,unitNum int,conn net.Conn)( code int ,err error){
	buf := make([]byte,128)
	for{
		n, err := s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		strResult := string(buf[:n])
		if(isEmpty(strResult)){
			return 10001,nil
		}
		//strResult = strings.Replace(strResult, "\r", "", -1)
		strResult = strings.Replace(strResult, "\n", "", -1)
		var tcp_msg =  strconv.Itoa(unitNum) + "$"
		tcp_msg = tcp_msg + strResult
		log.Println(tcp_msg)
		send_tcp(conn,tcp_msg)
		if strings.Contains(strResult,"+NETOPEN:"){
			break
		}else if strings.Contains(strResult,"+IP ERROR: Network is already opened") {
			return 10002,nil
		}
	}
	return 10000,nil
}

func receiveAtIpAddrCom(s io.ReadWriteCloser,unitNum int,mask string,conn net.Conn)( code int ,getIP string){
	buf := make([]byte,128)
	Ip := ""
	for{
		n, err := s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		strResult := string(buf[:n])
		if(isEmpty(strResult)){
			return 10001,""
		}
		//strResult = strings.Replace(strResult, "\r", "", -1)
		strResult = strings.Replace(strResult, "\n", "", -1)
		var tcp_msg =  strconv.Itoa(unitNum) + "$"
		tcp_msg = tcp_msg + strResult
		log.Println(tcp_msg)
		send_tcp(conn,tcp_msg)
		if strings.Contains(strResult,"+IPADDR:"){
			Ip =strings.Split(strResult, ":")[1]
			Ip = strings.Replace(Ip, "\r", "", -1)
			Ip = strings.Replace(Ip, "\n", "", -1)
			Ip = strings.Replace(Ip, " ", "", -1)
			Ip = strings.Replace(Ip, "OK", "", -1)
			log.Println("Ip = ",Ip)
			tmp,_:=IpContains(Ip,mask)
			if !tmp {
				return 10004,Ip
			}
			break
		}else if strings.Contains(strResult,"ERROR ") {
			return 10003,Ip
		}
	}
	return 10000,Ip
}

func isEmpty(a interface{}) bool {
	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		v=v.Elem()
	}
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

func send_tcp(conn net.Conn, msg string ) {

	if(!isEmpty(msg)) {
		fmt.Fprintf(conn, msg)
	}
	msg,err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Fatal("recv data error")
	}
	//}else{
	//	log.Println("recv msg : ",msg)
	//}
}

func Ping_test(s io.ReadWriteCloser,unitNum int,PingNum int,lock *sync.Mutex, conn net.Conn) ( error){
	lock.Lock()
	apnInfoes,apnNum :=get_cmd_info(unitNum,conn)
	log.Println("apnInfoes[0]=",apnInfoes[0].apnName,",num=",apnNum)

	timeStr1 := time.Now().Format("2006-01-02")
	fmt.Println(timeStr1)
	//使用Parse 默认获取为UTC时区 需要获取本地时区 所以使用ParseInLocation +" 23:59:59"
	t1, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr1 + " " + apnInfoes[0].startTime, time.Local)
	fmt.Println(t1.Unix() )
	fmt.Println(t1.Format("2006-01-02 15:04:05") )
	t2, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr1 + " " + apnInfoes[0].endIime, time.Local)
	fmt.Println(t2.Unix() )
	fmt.Println(t2.Format("2006-01-02 15:04:05") )
	now := time.Now()
	fmt.Println(now.Unix() )
	fmt.Println(now.Format("2006-01-02 15:04:05") )

	if t1.Unix() < now.Unix() && now.Unix() < t2.Unix() {

	}else {
		lock.Unlock()

		runtime.Gosched()
		return nil
	}

	var AT_cmd_str string
	for i:=0 ; i < apnNum ; i++  {
		apnStartTime :=time.Now()
		n, err := s.Write([]byte("AT\r\n"))
		if err != nil {
			log.Fatal(err,n)
		}
		receiveCom(s,unitNum,conn)

		n, err = s.Write([]byte("AT+CPIN?\r\n"))
		if err != nil {
			log.Fatal(err,n)
		}
		receiveCom(s,unitNum,conn)
		AT_cmd_str = "AT+CGDCONT=1,\"IP\",\"" + apnInfoes[i].apnName + "\"\r\n"
		n, err = s.Write([]byte(AT_cmd_str))
		if err != nil {
			log.Fatal(err,n)
		}
		receiveCom(s,unitNum,conn)

		n, err = s.Write([]byte("AT+CSOCKSETPN=1\r\n"))
		if err != nil {
			log.Fatal(err,n)
		}
		receiveCom(s,unitNum,conn)
		for ;  ;  {
			n, err = s.Write([]byte("AT+NETOPEN\r\n"))
			if err != nil {
				log.Fatal(err,n)
			}
			n, err =receiveAtNetOpenCom(s,unitNum,conn)
			if n == 10002 {
				n, err = s.Write([]byte("AT+NETCLOSE\r\n"))
				if err != nil {
					log.Fatal(err,n)
				}
				receiveCom(s,unitNum,conn)
			}else {
				break
			}
		}
		n, err = s.Write([]byte("AT+IPADDR\r\n"))
		if err != nil {
			log.Fatal(err,n)
		}
		//mask := "112.11.2.1/22"
		IpAddrCode,getIp := receiveAtIpAddrCom(s,unitNum,apnInfoes[i].ggsnIP,conn)
		log.Println("IpAddrCode = ",IpAddrCode)


		AT_cmd_str = "AT+CPING=\""+ apnInfoes[i].mobileIP +"\",1,4,64,1000,10000,255\r\n"

		n, err = s.Write([]byte(AT_cmd_str))
		if err != nil {
			log.Fatal(err,n)
		}
		PingMobleText,err:=receiveAtPingCom(s,unitNum,conn)
		if err != nil{
			log.Fatal(err,PingMobleText)
		}else {
			log.Println("PingMobleText = ",PingMobleText)
		}

		AT_cmd_str = "AT+CPING=\""+ apnInfoes[i].endIP +"\",1,4,64,1000,10000,255\r\n"

		n, err = s.Write([]byte(AT_cmd_str))
		if err != nil {
			log.Fatal(err,n)
		}
		PingEndText,err:=receiveAtPingCom(s,unitNum,conn)
		if err != nil{
			log.Fatal(err,PingEndText)
		}else {
			log.Println("PingEndText = ",PingEndText)
		}

		AT_cmd_str = "AT+CPING=\""+ apnInfoes[i].exchangeIP +"\",1,4,64,1000,10000,255\r\n"

		n, err = s.Write([]byte(AT_cmd_str))
		if err != nil {
			log.Fatal(err,n)
		}
		PingExchangedText,err:=receiveAtPingCom(s,unitNum,conn)
		if err != nil{
			log.Fatal(err,PingExchangedText)
		}else {
			log.Println("PingEndText = ",PingExchangedText)
		}

		n, err = s.Write([]byte("AT+NETCLOSE\r\n"))
		if err != nil {
			log.Fatal(err,n)
		}
		receiveCom(s,unitNum,conn)
		PingMobleStatus :=  -1
		PingEndStatus :=  -1
		PingExchangeStatus :=  -1

		if !strings.Contains(PingMobleText,"+CPING: 1") {
			PingMobleStatus = 0
		}else {
			PingMobleStatus = 1
		}

		if !strings.Contains(PingEndText,"+CPING: 1") {
			PingEndStatus = 0
		}else {
			PingEndStatus = 1
		}

		if !strings.Contains(PingExchangedText,"+CPING: 1") {
			PingExchangeStatus = 0
		}else {
			PingExchangeStatus = 1
		}

		StrBody :=""
		StrBody += "" + strconv.Itoa(unitNum)
		StrBody += "$data$"
		StrBody += "unitPhone=" + apnInfoes[i].phoneNum
		now := time.Now()
		year, month, day := now.Date()
		today_str := fmt.Sprintf("%04d%02d%02d", year, month, day)
		resultSn := fmt.Sprintf("%03d%d%s%08d",apnInfoes[i].deviceid,unitNum,today_str,PingNum )
		StrBody += "&resultSn=" + resultSn
		StrBody += "&cmdId=" + strconv.Itoa(apnInfoes[i].cmdid)
		StrBody += "&cmdType=" + strconv.Itoa(apnInfoes[i].cmdtype)
		StrBody += "&apnId=" + strconv.Itoa(apnInfoes[i].apnid)
		StrBody += "&netType=" + strconv.Itoa(apnInfoes[i].netType)

		apnActivate := -1
		errType := -1
		if PingMobleStatus == 1 {

			apnActivate = 1
		} else if PingEndStatus ==1 && PingExchangeStatus ==1 {
			apnActivate = 1
			errType = 0;
		}else {
			apnActivate = 0
			errType = 1;
		}
		StrBody += "&errType=" + strconv.Itoa(errType)
		StrBody += "&apnActivate=" + strconv.Itoa(apnActivate)

		if IpAddrCode == 10000 {
			StrBody += "&ipIsGet=" + strconv.Itoa(1)
		}else {
			StrBody += "&ipIsGet=" + strconv.Itoa(0)
		}
		StrBody += "&ipaddr=" + getIp

		StrBody += "&pingMobileIP=" + strconv.Itoa(PingMobleStatus)
		StrBody += "&pingMobileIPText=" + PingMobleText

		StrBody += "&pingEndIP=" + strconv.Itoa(PingEndStatus)
		StrBody += "&pingEndIPText=" + PingEndText

		StrBody += "&pingExchangeIP=" + strconv.Itoa(PingExchangeStatus)
		StrBody += "&pingExchangeIPText=" + PingExchangedText

		apnStartTimeStr := fmt.Sprintf("%d",apnStartTime.Unix())
		StrBody += "&startTime=" + apnStartTimeStr
		StrBody += "&endTime=" + fmt.Sprintf("%d",time.Now().Unix())

		log.Println("StrBody=",StrBody)
		send_tcp(conn,StrBody)
	}
	lock.Unlock()

	runtime.Gosched()
	return nil
}

func IpContains(userIP string,mask string) (bool,error)  {
	ipUser := net.ParseIP(userIP)
	log.Println("mask =",mask)
	if !strings.Contains(mask,"/") {
		return false,nil
	}
	_, n, err := net.ParseCIDR(mask)
	if err != nil {
		panic(err)
	}

	if n.Contains(ipUser){
		return true,err
	}else {
		return false,err
	}

}

func get_cmd_info(Uint int,conn net.Conn)([]APN_info,int){
	fmt.Fprintf(conn, "1$cmd")
	msg,err := bufio.NewReader(conn).ReadString('\n')
	msg = strings.Replace(msg, "\n", "", -1)
	if err != nil {
		log.Fatal("recv data error")
	}else{
		log.Println("recv msg : ",msg)
	}

	msg_info:=strings.Split(msg, "$")
	apnNum, _  := strconv.Atoi(msg_info[0])
	//msg_info[1] = strings.Replace(msg_info[1], "\n", "", -1)
	apn_list := strings.Split(msg_info[1], "&")
	apnInfoes := make([]APN_info,0)
	for i :=0; i < apnNum; i++  {
		apnInfoTmp:= APN_info{}
		apnInfoTmp.apnName = strings.Split(apn_list[i], ",")[1]
		apnInfoTmp.apnType,_ = strconv.Atoi(strings.Split(apn_list[i],",")[2])
		apnInfoTmp.addressType,_ = strconv.Atoi(strings.Split(apn_list[i],",")[3])
		apnInfoTmp.ggsnIP = strings.Split(apn_list[i], ",")[4]
		apnInfoTmp.mobileIP = strings.Split(apn_list[i], ",")[5]
		apnInfoTmp.endIP = strings.Split(apn_list[i], ",")[6]
		apnInfoTmp.exchangeIP = strings.Split(apn_list[i], ",")[7]
		apnInfoTmp.netType,_ = strconv.Atoi(strings.Split(apn_list[i],",")[8])
		apnInfoTmp.pingType,_ = strconv.Atoi(strings.Split(apn_list[i],",")[9])
		apnInfoTmp.phoneNum = strings.Split(apn_list[i], ",")[10]
		apnInfoTmp.deviceid,_ = strconv.Atoi(strings.Split(apn_list[i],",")[11])
		apnInfoTmp.apnid,_ = strconv.Atoi(strings.Split(apn_list[i],",")[12])
		apnInfoTmp.cmdid,_ = strconv.Atoi(strings.Split(apn_list[i],",")[13])
		apnInfoTmp.cmdtype,_ = strconv.Atoi(strings.Split(apn_list[i],",")[14])
		apnInfoTmp.packlen,_ = strconv.Atoi(strings.Split(apn_list[i],",")[15])
		apnInfoTmp.startTime = strings.Split(apn_list[i], ",")[16]
		apnInfoTmp.endIime = strings.Split(apn_list[i], ",")[17]
		apnInfoes = append(apnInfoes,apnInfoTmp)
	}
	return apnInfoes,len(apnInfoes)
}

func get_com_info(conn net.Conn)([]string){
	fmt.Fprintf(conn, "com")
	msg,err := bufio.NewReader(conn).ReadString('\n')
	msg = strings.Replace(msg, "\n", "", -1)
	if err != nil {
		log.Fatal("recv data error")
	}else{
		log.Println("recv msg : ",msg)
	}

	msg_info:=strings.Split(msg, ",")


	return msg_info
}

func main() {
	hostInfo := "127.0.0.1:8010"
	conn,err := net.Dial("tcp",hostInfo)
	lock := &sync.Mutex{}
	lock2 := &sync.Mutex{}
	lock3 := &sync.Mutex{}
	lock4 := &sync.Mutex{}
	if err != nil {
		log.Println("connect (",hostInfo,") fail")
	}else{
		log.Println("connect (",hostInfo,") ok")
		defer conn.Close()
	}
	comlist:=get_com_info(conn)
	var s[4] io.ReadWriteCloser
	for i := 0; i< (len(comlist)-1); i++  {
		log.Println("com= ",comlist[i])
		//设置串口编号
		c := &serial.Config{Name: comlist[i], Baud: 115200}
		//打开串口
		s[i], err = serial.OpenPort(c)
		if err != nil {
			log.Fatal(err)
		}
	}
	unitNumALL,_ := strconv.Atoi(comlist[len(comlist)-1])
	for num :=1 ;  ; num++ {
		timeStr := time.Now().Format("2006-01-02")
		fmt.Println(timeStr)
		//使用Parse 默认获取为UTC时区 需要获取本地时区 所以使用ParseInLocation +" 23:59:59"
		t, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
		fmt.Println(t.Unix() )

		for j := 0; j < unitNumALL;j++  {
			switch j {
			case 0:
				go Ping_test(s[j],j+1,num,lock,conn)
			case 1:
				go Ping_test(s[j],j+1,num,lock2,conn)
			case 2:
				go Ping_test(s[j],j+1,num,lock3,conn)
			case 4, 5, 6:
				go Ping_test(s[j],j+1,num,lock4,conn)
			}

		}
		lock.Lock()
		lock2.Lock()
		lock3.Lock()
		lock4.Lock()
		timeStr1 := time.Now().Format("2006-01-02")
		fmt.Println(timeStr1)
		//使用Parse 默认获取为UTC时区 需要获取本地时区 所以使用ParseInLocation +" 23:59:59"
		t1, _ := time.ParseInLocation("2006-01-02", timeStr1, time.Local)
		fmt.Println(t1.Unix() )

		if t.Unix() < t1.Unix() {
			num = 0
		}
		lock.Unlock()
		lock2.Unlock()
		lock3.Unlock()
		lock4.Unlock()

		runtime.Gosched()
		// 出让时间片
	}

}
