package main

import "fmt"
import "log"
import "strings"
import "strconv"
import "net"
import "io"
import "io/ioutil"
import "time"
import "crypto/sha1"
import "os"
import "math/rand"
import "net/http"
import "encoding/json"

type Balance struct {
	Result struct {
		Balance struct {
			Balance float32 `json:"balance"`
			Created string
			Username string `json:"username"`
			Verifed string `json:"verifed"`
		} `json:"balance"`
		Miners []byte `json:"miners"`
		Transactions []byte `json:"transactions"`
	}
	Server string
	Success bool
}

type PoolInfo struct {
	Ip string
	Name string
	Port int32
	Server string
	Success bool
}

type Stats struct {
	Connections int `json:"Active connections"`
	TotalMined float64 `json:"All-timed mined DUCO"`
	CurrentDifficulty int `json:"Current difficulty"`
	DUCOHashrate string
	JustSwap float32
	NodeS float32
	PancakeSwap float32
	SushiSwap float32
	DUCO float32 `json:"DUCO price"`
	BCH float32 
	NANO float32
	TRX float32
	XMG float32
	Api string
	Kolka struct {
		Banned int
		Jailed int
	}
	LastHash string
	LastSync float64
	LastUpdated string
	MinedHashes string
}

var acc int
var username string = "ndraey";
var data Balance

var pool PoolInfo
var threads int = 7;
var priced Stats

func main(){
	args:=os.Args[1:]
	for i:=0; i<len(args); i++ {
		if args[i]=="-t" { //threads
			if len(args)>i+1 { //exists
				threads, _ = strconv.Atoi(args[i+1])
				args=RemoveIndex(args,i)
				args=RemoveIndex(args,i)
			}else{
				print("Threads (-t) parameter needs argument\n")
				os.Exit(1)
			}
		}
		if len(args)==0 {break}
		if args[i]=="-h" {
			print(os.Args[0]+` [-t THREADS] nickname`+"\n")
			os.Exit(1)
		}
	}
	if len(args)>=1 {
		username = args[len(args)-1]
	}
	bannerw:=42
	unlen:=14+len(username)
	threadslen:=12+len(strconv.Itoa(threads))
	print("------------------------------------------\n")
	print("| PulseMiner - DUCO Miner written in Go  |\n")
	print("|                                        |\n")
	print("| Mining to: "+username+strings.Repeat(" ",bannerw-unlen)+"|\n")
	print("| Threads: "+strconv.Itoa(threads)+strings.Repeat(" ",bannerw-threadslen)+"|\n")
	print("------------------------------------------\n")
	print("\nGetting best pool...\n")
	tothashrate := make([]int,threads)
	pooldata, err := getnet("https://server.duinocoin.com/getPool")
	if err!=nil {print("Error getting pool! Press Ctrl-C to exit...\n"); for{}}
	json.Unmarshal(pooldata,&pool)
	fmt.Printf("Selected pool %s:%d\n",pool.Ip,pool.Port)
	print("Connecting...\n")
	rand.Seed(time.Now().UnixNano())
	minerid:=rand.Int63()%1712
	for i:=0; i<threads; i++{
		go mine(pool.Ip,pool.Port,minerid,i,tothashrate)
	}
	for{
		info, err := http.Get("https://server.duinocoin.com/"+
							  "users/"+username)
		if err!=nil{
			continue
		}
		body, err := ioutil.ReadAll(info.Body)
		if err!=nil{
			continue
		}
		json.Unmarshal(body,&data)
		balance:=data.Result.Balance.Balance

		pricedata, err := getnet("https://server.duinocoin.com/statistics")
		if err!=nil {log.Fatal(err)}
		json.Unmarshal(pricedata, &priced)
		
		fmt.Printf("[feed] Balance: %.5f DUCO (≈%.5f USD)\n",
					balance,
					balance*priced.DUCO)
		time.Sleep(30*time.Second)
	}
}

func getnet(url string) ([]byte, error) {
	info, err := http.Get(url)
	body, err := ioutil.ReadAll(info.Body)
	return body, err
} 

func sum(array []int) int {  
	result := 0  
	for _, v := range array {  
		result += v  
	}  
	return result  
}

func mine(ip string, port int32, minerid int64, cpuid int, thr []int){
	print("Started...\n")
	conn, _ := net.Dial("tcp",ip+":"+strconv.Itoa(int(port)))
	ver := make([]byte,10)
	conn.Read(ver)
	//print("Version: "+string(ver)+"\n")
	for {
		job_ := string(writejobreq(conn))
		job := strings.Split(job_,",")
		if len(job)<3 {
			fmt.Printf("Skipping job: %v\n",job)
			//mine(ip,port,minerid,cpuid,thr)
			continue
		}
		jh := job[0]
		nh := job[1]
		diff := strings.Replace(string(job[2]), "\n", "", -1)
		diff = strings.Replace(string(diff), "\x00", "", -1)
		diffi, err := strconv.ParseInt(diff,10,32)
		if err!=nil { fmt.Println(err); os.Exit(1) }
		nonce, hashrate, time := DecodeHash(jh,nh,int(diffi))
		conn.Write([]byte(
			strconv.Itoa(int(nonce))+","+
			strconv.Itoa(int(hashrate))+",Official PC Miner,pulsemon-go,,"+
			strconv.Itoa(int(minerid))))
		result := make([]byte,32)
		conn.Read(result)
		thr[cpuid]=hashrate
		if strings.HasPrefix(string(result), "GOOD") {
			acc++
			fmt.Printf("[%d] [cpu%d] \033[32mGOOD\033[0m H: %sH/s T: %.2fs\n",
						acc,
						cpuid,
						convr(sum(thr)),
						time)
		}else if(strings.HasPrefix(string(result),"BAD")){
			fmt.Printf("%s\n",string(result))
			fmt.Printf("[%d] [cpu%d] \033[31;1mBAD\033[0m H: %sH/s T: %.2fs\n",
						acc,
						cpuid,
						convr(sum(thr)),
						time)
		}
	}
}

func gettime() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond)
}

func RemoveIndex(s []string, i int) []string {
    return s[:i+copy(s[i:], s[i+1:])]
}

func convr(num int) string {
	rts:=""
	if num>(1000*1000*1000){
		k := strconv.Itoa(int(num/(1000*1000*1000)))
		rts = k+"G"
	}else if num>(1000*1000) {
		k := strconv.Itoa(int(num/(1000*1000)))
		rts = k+"M"
	}else if(num>1000){
		k := strconv.Itoa(int(num/1000))
		rts = k+"k"
	}else{
		rts = strconv.Itoa(int(num))
	}
	return rts
}

func DecodeHash(prev string,result string,diff int) (int, int, float64) {
	i:=int(0)
	hr := int(0)
	tm := float64(gettime()/1000)
	tmm := float64(0)
	for i=0; i<(100*diff+1); i++ {
		hash := sha1.Sum([]byte(prev+strconv.Itoa(i)))
		hashd := fmt.Sprintf("%02x",hash)
		if hashd==result {
			tmm = float64(gettime()/1000)-tm
			if tmm<=1 { tmm = 1 }
			hr = i/int(tmm)
			break
		}
	}
	return i, hr, tmm
}

func writejobreq(con net.Conn) []byte {
	con.Write([]byte("JOB,"+username+",LOW,"))
	req := make([]byte,90)
	con.Read(req)
	return req
}

func copystr(dst io.Writer, src io.Reader) {
   if _, err := io.Copy(dst, src); err != nil {
      log.Fatal(err)
   }
}
