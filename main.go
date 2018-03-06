package  main

import (
	"flag"
	"github.com/tarm/goserial"
	"log"
	"io"
	"strings"
	"net"
	"fmt"
	"bufio"
	"reflect"
)
var (
	conFile = flag.String("configfile","/config.ini","config file")
)

var TOPIC = make(map[string]string)

func receiveCom(s io.ReadWriteCloser,conn net.Conn)( code int ,err error){
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
		var tcp_msg =  "1:"
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

func receiveAtPingCom(s io.ReadWriteCloser,conn net.Conn)( code int ,err error){
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
		var tcp_msg =  "1:"
		tcp_msg = tcp_msg + strResult
		log.Println(tcp_msg)
		send_tcp(conn,tcp_msg)
		if strings.Contains(strResult,"+CPING: 3,"){
			break
		}
	}
	return 10000,nil
}

func receiveAtNetOpenCom(s io.ReadWriteCloser,conn net.Conn)( code int ,err error){
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
		var tcp_msg =  "1:"
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

func receiveAtIpAddrCom(s io.ReadWriteCloser,mask string,conn net.Conn)( code int ,err error){
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
		var tcp_msg =  "1:"
		tcp_msg = tcp_msg + strResult
		log.Println(tcp_msg)
		send_tcp(conn,tcp_msg)
		if strings.Contains(strResult,"+IPADDR:"){
			Ip:=strings.Split(strResult, ":")[1]
			Ip = strings.Replace(Ip, "\r", "", -1)
			Ip = strings.Replace(Ip, "\n", "", -1)
			Ip = strings.Replace(Ip, " ", "", -1)
			tmp,_:=IpContains(Ip,mask)
			if !tmp {
				return 10004,nil
			}
			break
		}else if strings.Contains(strResult,"ERROR ") {
			return 10003,nil
		}
	}
	return 10000,nil
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
	}else{
		log.Println("recv msg : ",msg)
	}
}

func Ping_test(s io.ReadWriteCloser,conn net.Conn){

	n, err := s.Write([]byte("AT\r\n"))
	if err != nil {
		log.Fatal(err,n)
	}
	receiveCom(s,conn)

	n, err = s.Write([]byte("AT+CPIN?\r\n"))
	if err != nil {
		log.Fatal(err,n)
	}
	receiveCom(s,conn)

	//n, err = s.Write([]byte("AT+CNSMOD?\r\n"))
	//if err != nil {
	//	log.Fatal(err,n)
	//}
	//receiveCom(s,conn)

	n, err = s.Write([]byte("AT+CGDCONT=1,\"IP\",\"CMIOT\"\r\n"))
	if err != nil {
		log.Fatal(err,n)
	}
	receiveCom(s,conn)

	n, err = s.Write([]byte("AT+CGDCONT?\r\n"))
	if err != nil {
		log.Fatal(err,n)
	}
	receiveCom(s,conn)

	n, err = s.Write([]byte("AT+CSOCKSETPN=1\r\n"))
	if err != nil {
		log.Fatal(err,n)
	}
	receiveCom(s,conn)
	for ;  ;  {
		n, err = s.Write([]byte("AT+NETOPEN\r\n"))
		if err != nil {
			log.Fatal(err,n)
		}
		n, err =receiveAtNetOpenCom(s,conn)
		if n == 10002 {
			n, err = s.Write([]byte("AT+NETCLOSE\r\n"))
			if err != nil {
				log.Fatal(err,n)
			}
			receiveCom(s,conn)
		}else {
			break
		}
	}
	n, err = s.Write([]byte("AT+IPADDR\r\n"))
	if err != nil {
		log.Fatal(err,n)
	}
	mask := "112.11.2.1/22"
	receiveAtIpAddrCom(s,mask,conn)

	n, err = s.Write([]byte("AT+CPING=\"61.135.169.125\",1,4,64,1000,10000,255\r\n"))
	if err != nil {
		log.Fatal(err,n)
	}
	receiveAtPingCom(s,conn)

	n, err = s.Write([]byte("AT+NETCLOSE\r\n"))
	if err != nil {
		log.Fatal(err,n)
	}
	receiveCom(s,conn)
}

func IpContains(userIP string,mask string) (bool,error)  {
	ipUser := net.ParseIP(userIP)
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

func main() {
	hostInfo := "127.0.0.1:8010"
	conn,err1 := net.Dial("tcp",hostInfo)

	if err1 != nil {
		log.Println("connect (",hostInfo,") fail")
	}else{
		log.Println("connect (",hostInfo,") ok")
		defer conn.Close()
	}
	//设置串口编号
	c := &serial.Config{Name: "COM6", Baud: 115200}
	//打开串口
	s, err := serial.OpenPort(c)

	if err != nil {
		log.Fatal(err)
	}

	Ping_test(s,conn)
}
