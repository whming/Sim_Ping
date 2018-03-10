package  main

import (
	"flag"
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
	"os"
	"net/http"
	"io/ioutil"
)

//var unitNum int = 1
var lock_cmd = &sync.Mutex{}
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
		}else if strings.Contains(strResult,"+CME ERROR: SIM not inserted"){
			return 10001,nil
		}else if strings.Contains(strResult,"ERROR") {
			return 10002, nil
		}
		//else {
		//	break
		//}
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
		}else  if strings.Contains(strResult,"ERROR"){
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
		} else if strings.Contains(strResult,"+NETOPEN: 1") {
			return 10003, nil
		}
	}
	return 10000,nil
}

func receiveCLOSECom(s io.ReadWriteCloser,unitNum int,conn net.Conn)( code int ,err error){
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
		if strings.Contains(strResult,"+NETCLOSE:"){
			break
		}else if strings.Contains(strResult,"ERROR"){
			break
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
			return 10001,Ip
		}
		//strResult = strings.Replace(strResult, "\r", "", -1)
		strResult = strings.Replace(strResult, "\n", "", -1)
		var tcp_msg =  strconv.Itoa(unitNum) + "$"
		tcp_msg = tcp_msg + strResult
		log.Println(tcp_msg)
		send_tcp(conn,tcp_msg)
		if strings.Contains(strResult,"ERROR") {
			return 10003,Ip
		}else if strings.Contains(strResult,"+IPADDR:"){
			Ip =strings.Split(strResult, ":")[1]
			Ip = strings.Replace(Ip, "\r", "", -1)
			Ip = strings.Replace(Ip, "\n", "", -1)
			Ip = strings.Replace(Ip, " ", "", -1)
			Ip = strings.Replace(Ip, "OK", "", -1)
			log.Println("Ip = ",Ip)
			maskList :=strings.Split(mask, ",")
			//var tmp[len(maskList)] bool
			BooL_temp  := false
			for i:=0; i< len(maskList); i++  {
				//log.Println("mask =",maskList[i])
				tmp,_ :=IpContains(Ip,maskList[i])
				BooL_temp = BooL_temp || tmp
			}
			if !BooL_temp {
				return 10004,Ip
			}

			break
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
		_,err := fmt.Fprintf(conn, msg)
		//msg,err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal("send data error")
		}
		//send_tcp(conn net.Conn, msg string )
		//}else{
		//	log.Println("send msg : ",msg,"NUm=",n)
		//}
	}


}
func get_url(conn net.Conn, msg string )( string ) {
	var remsg string
	if(!isEmpty(msg)) {
		_,err := fmt.Fprintf(conn, msg)
		remsg,err = bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal("send data error")
			return ""
		}
		remsg=strings.Replace(remsg, "\n", "", -1)
		remsg=strings.Replace(remsg, " ", "", -1)
		return remsg
	} else {
		return ""
	}
}



func httpPost( strurl string,strbody string) {
	//"http://47.94.150.126:9081/api/comm/result"
	resp, err := http.Post(strurl,
		"application/x-www-form-urlencoded",
		strings.NewReader(strbody))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	log.Println(string(body))
}

func Ping_test(s io.ReadWriteCloser,unitNum int,PingNum int, conn net.Conn) ( error){

	apnInfoes,apnNum :=get_cmd_info(unitNum,conn)
	//log.Println("apnInfoes[0]=",apnInfoes[0].apnName,",num=",apnNum,"unitNUM=",unitNum)

	timeStr1 := time.Now().Format("2006-01-02")
	//fmt.Println(timeStr1)
	//使用Parse 默认获取为UTC时区 需要获取本地时区 所以使用ParseInLocation +" 23:59:59"
	t1, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr1 + " " + apnInfoes[0].startTime, time.Local)
	//fmt.Println(t1.Unix() )
	//fmt.Println(t1.Format("2006-01-02 15:04:05") )
	t2, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr1 + " " + apnInfoes[0].endIime, time.Local)
	//fmt.Println(t2.Unix() )
	//fmt.Println(t2.Format("2006-01-02 15:04:05") )
	now := time.Now()
	//fmt.Println(now.Unix() )
	//fmt.Println(now.Format("2006-01-02 15:04:05") )

	if t1.Unix() < now.Unix() && now.Unix() < t2.Unix() {

	}else {
		return nil
	}

	var AT_cmd_str string
	for i:=0 ; i < apnNum ; i++  {
		apnStartTime :=time.Now()
		//n, err := s.Write([]byte("AT\r\n"))
		//if err != nil {
		//	log.Fatal(err,n)
		//}
		//receiveCom(s,unitNum,conn)

		n, err := s.Write([]byte("AT+CPIN?\r\n"))
		if err != nil {
			log.Fatal(err,n)
		}
		n,_ = receiveCom(s,unitNum,conn)
		if n == 10001 {
			time.Sleep(1*time.Minute)
			break
		}
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

			n, err = s.Write([]byte("AT+NETOPEN\r\n"))
			if err != nil {
				log.Fatal(err,n)
			}
			opencode, err :=receiveAtNetOpenCom(s,unitNum,conn)
			if opencode == 10002 {
				n, err = s.Write([]byte("AT+NETCLOSE\r\n"))
				if err != nil {
					log.Fatal(err,n)
				}
				receiveCom(s,unitNum,conn)
				i -= 1
				continue
			}else if opencode == 10003 {
				StrBody := "unitPhone=" + apnInfoes[i].phoneNum
				now := time.Now()
				year, month, day := now.Date()
				today_str := fmt.Sprintf("%04d%02d%02d", year, month, day)
				resultSn := fmt.Sprintf("%03d%d%s%08d",apnInfoes[i].deviceid,unitNum,today_str,PingNum )
				StrBody += "&resultSn=" + resultSn
				StrBody += "&cmdId=" + strconv.Itoa(apnInfoes[i].cmdid)
				StrBody += "&cmdType=" + strconv.Itoa(apnInfoes[i].cmdtype)
				StrBody += "&apnId=" + strconv.Itoa(apnInfoes[i].apnid)
				StrBody += "&netType=" + strconv.Itoa(apnInfoes[i].netType)
				StrBody += "&errType=" + strconv.Itoa(1)
				StrBody += "&apnActivate=" + strconv.Itoa(0)
				StrBody += "&ipIsGet=" + strconv.Itoa(0)

				StrBody += "&ipaddr=" + ""

				StrBody += "&pingMobileIP=" + "0"
				StrBody += "&pingMobileIPText=" + ""

				StrBody += "&pingEndIP=" + "0"
				StrBody += "&pingEndIPText=" + ""

				StrBody += "&pingExchangeIP=" + strconv.Itoa(0)
				StrBody += "&pingExchangeIPText=" + ""

				apnStartTimeStr := fmt.Sprintf("%d",apnStartTime.Unix())
				StrBody += "&startTime=" + apnStartTimeStr
				StrBody += "&endTime=" + fmt.Sprintf("%d",time.Now().Unix())

				log.Println("i=",i)
				//send_tcp(conn,StrBody)
				msgstrul := "" + strconv.Itoa(unitNum)
				msgstrul += "$strurl"
				msgstrul = get_url(conn,msgstrul)
				log.Println("strurl=",msgstrul)
				if !isEmpty(msgstrul) {
					httpPost(msgstrul,StrBody)
				}
				data_str :=strconv.Itoa(unitNum)
				data_str += "$data"
				send_tcp(conn,data_str)
				time.Sleep(5*time.Second)
				continue
			}

		n, err = s.Write([]byte("AT+IPADDR\r\n"))
		if err != nil {
			log.Fatal(err,n)
		}
		//mask := "112.11.2.1/22"
		IpAddrCode,getIp := receiveAtIpAddrCom(s,unitNum,apnInfoes[i].ggsnIP,conn)
		log.Println("IpAddrCode = ",IpAddrCode)
		//if 10000 != IpAddrCode{
		//正常使用
		// 测试使用
		if 10003 == IpAddrCode{
			n, err = s.Write([]byte("AT+NETCLOSE\r\n"))
			if err != nil {
				log.Fatal(err,n)
			}
			receiveCLOSECom(s,unitNum,conn)
			StrBody := "unitPhone=" + apnInfoes[i].phoneNum
				now := time.Now()
				year, month, day := now.Date()
				today_str := fmt.Sprintf("%04d%02d%02d", year, month, day)
				resultSn := fmt.Sprintf("%03d%d%s%08d",apnInfoes[i].deviceid,unitNum,today_str,PingNum )
				StrBody += "&resultSn=" + resultSn
				StrBody += "&cmdId=" + strconv.Itoa(apnInfoes[i].cmdid)
				StrBody += "&cmdType=" + strconv.Itoa(apnInfoes[i].cmdtype)
				StrBody += "&apnId=" + strconv.Itoa(apnInfoes[i].apnid)
				StrBody += "&netType=" + strconv.Itoa(apnInfoes[i].netType)
				StrBody += "&errType=" + strconv.Itoa(1)
				StrBody += "&apnActivate=" + strconv.Itoa(0)
				StrBody += "&ipIsGet=" + strconv.Itoa(0)

				StrBody += "&ipaddr=" + ""

				StrBody += "&pingMobileIP=" + "0"
				StrBody += "&pingMobileIPText=" + ""

				StrBody += "&pingEndIP=" + "0"
				StrBody += "&pingEndIPText=" + ""

				StrBody += "&pingExchangeIP=" + strconv.Itoa(0)
				StrBody += "&pingExchangeIPText=" + ""

				apnStartTimeStr := fmt.Sprintf("%d",apnStartTime.Unix())
				StrBody += "&startTime=" + apnStartTimeStr
				StrBody += "&endTime=" + fmt.Sprintf("%d",time.Now().Unix())

				log.Println("i=",i)
				msgstrul := strconv.Itoa(unitNum)
				msgstrul += "$strurl"
				msgstrul = get_url(conn,msgstrul)
			log.Println("strurl=",msgstrul)
				if !isEmpty(msgstrul) {
					httpPost(msgstrul,StrBody)
				}
			data_str :=strconv.Itoa(unitNum)
			data_str += "$data"
			send_tcp(conn,data_str)
				time.Sleep(5*time.Second)
			continue
		}

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

		AT_cmd_str = "AT+CPINGSTOP\r\n"

		n, err = s.Write([]byte(AT_cmd_str))
		if err != nil {
			log.Fatal(err,n)
		}
		n,_ = receiveCom(s,unitNum,conn)
		//if err != nil{
		//	log.Fatal(err)
		//}else {
		//	log.Println("n = ",n)
		//}

		AT_cmd_str = "AT+CPING=\""+ apnInfoes[i].endIP +"\",1,4,64,1000,10000,255\r\n"

		n, err = s.Write([]byte(AT_cmd_str))
		if err != nil {
			log.Fatal(err,n)
		}
		PingEndText,err:=receiveAtPingCom(s,unitNum,conn)
		if err != nil{
			log.Fatal(err)
		}else {
			log.Println("n = ",PingEndText)
		}

		AT_cmd_str = "AT+CPINGSTOP\r\n"

		n, err = s.Write([]byte(AT_cmd_str))
		if err != nil {
			log.Fatal(err,n)
		}
		_,_ = receiveCom(s,unitNum,conn)
		//if err != nil{
		//	log.Fatal(err)
		//}
		//else {
		//	log.Println("n = ",n)
		//}

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

		AT_cmd_str = "AT+CPINGSTOP\r\n"

		n, err = s.Write([]byte(AT_cmd_str))
		if err != nil {
			log.Fatal(err,n)
		}
		_,_ = receiveCom(s,unitNum,conn)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//}else {
		//	log.Println("n = ",n)
		//}


		n, err = s.Write([]byte("AT+NETCLOSE\r\n"))
		if err != nil {
			log.Fatal(err,n)
		}
		receiveCLOSECom(s,unitNum,conn)

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

		StrBody := "unitPhone=" + apnInfoes[i].phoneNum
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

		log.Println("i=",i)
		//send_tcp(conn,StrBody)
		msgstrul := strconv.Itoa(unitNum)
		msgstrul += "$strurl"
		msgstrul = get_url(conn,msgstrul)
		log.Println("strurl=",msgstrul)
		if !isEmpty(msgstrul) {
			httpPost(msgstrul,StrBody)
		}
		data_str :=strconv.Itoa(unitNum)
		data_str += "$data"
		send_tcp(conn,data_str)
		time.Sleep(3*time.Second)
	}
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
	cmdstr := strconv.Itoa(Uint) + "$cmd"
	fmt.Fprintf(conn, cmdstr)
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
		apnInfoTmp.apnName = strings.Split(apn_list[i], "-")[1]
		apnInfoTmp.apnType,_ = strconv.Atoi(strings.Split(apn_list[i],"-")[2])
		apnInfoTmp.addressType,_ = strconv.Atoi(strings.Split(apn_list[i],"-")[3])
		apnInfoTmp.ggsnIP = strings.Split(apn_list[i], "-")[4]
		apnInfoTmp.mobileIP = strings.Split(apn_list[i], "-")[5]
		apnInfoTmp.endIP = strings.Split(apn_list[i], "-")[6]
		apnInfoTmp.exchangeIP = strings.Split(apn_list[i], "-")[7]
		apnInfoTmp.netType,_ = strconv.Atoi(strings.Split(apn_list[i],"-")[8])
		apnInfoTmp.pingType,_ = strconv.Atoi(strings.Split(apn_list[i],"-")[9])
		apnInfoTmp.phoneNum = strings.Split(apn_list[i], "-")[10]
		apnInfoTmp.deviceid,_ = strconv.Atoi(strings.Split(apn_list[i],"-")[11])
		apnInfoTmp.apnid,_ = strconv.Atoi(strings.Split(apn_list[i],"-")[12])
		apnInfoTmp.cmdid,_ = strconv.Atoi(strings.Split(apn_list[i],"-")[13])
		apnInfoTmp.cmdtype,_ = strconv.Atoi(strings.Split(apn_list[i],"-")[14])
		apnInfoTmp.packlen,_ = strconv.Atoi(strings.Split(apn_list[i],"-")[15])
		apnInfoTmp.startTime = strings.Split(apn_list[i], "-")[16]
		apnInfoTmp.endIime = strings.Split(apn_list[i], "-")[17]
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
	}//}else{
	//	log.Println("recv msg : ",msg)
	//}

	msg_info:=strings.Split(msg, ",")


	return msg_info
}

func test(Uint int,conn net.Conn){
	for ; ;  {
		fmt.Fprintf(conn, "com")
		msg,err := bufio.NewReader(conn).ReadString('\n')
		msg = strings.Replace(msg, "\n", "", -1)
		if err != nil {
			log.Fatal("recv data error")
		}else{
			log.Println("recv msg : ",msg,"===",Uint)
		}
		time.Sleep(10 * time.Second)
	}
}

func main() {
	//var Unit_Num_T uint
	//Unit_Num := flag.Int("Unit_Num", 0, "Unit_Num")
	//Unit_Num_T = Unit_Num
	hostInfo := "127.0.0.1:8010"
	conn,err := net.Dial("tcp",hostInfo)
	if err != nil {
		log.Println("connect (",hostInfo,") fail")
	}else{
		log.Println("connect (",hostInfo,") ok")
		defer conn.Close()
	}
	comlist:=get_com_info(conn)
	Unit_Num,_ := strconv.Atoi(os.Args[1])
	//UNITNUM := *Unit_Num
	log.Println("UNITNUM= ",Unit_Num)
	//var s[4] io.ReadWriteCloser

	log.Println("com= ",comlist[Unit_Num-1])
	//设置串口编号
	c := &serial.Config{Name: comlist[Unit_Num-1], Baud: 115200}
	//打开串口
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	for num :=1 ;  ; num++ {
		timeStr := time.Now().Format("2006-01-02")
		//fmt.Println(timeStr)
		//使用Parse 默认获取为UTC时区 需要获取本地时区 所以使用ParseInLocation +" 23:59:59"
		t, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
		//Ping_test(s[0],1,num,lock,conn)
		//fmt.Println(t.Unix() )
		Ping_test(s,Unit_Num,num,conn)

		timeStr1 := time.Now().Format("2006-01-02")
		//fmt.Println(timeStr1)
		//使用Parse 默认获取为UTC时区 需要获取本地时区 所以使用ParseInLocation +" 23:59:59"
		t1, _ := time.ParseInLocation("2006-01-02", timeStr1, time.Local)
		//fmt.Println(t1.Unix() )

		if t.Unix() < t1.Unix() {
			num = 1
		}
		// 出让时间片
	}

}