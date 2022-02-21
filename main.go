package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"github.com/NDRAEY/Pradz"
)

type Balance struct {
	Result struct {
		Balance struct {
			Balance  float32 `json:"balance"`
			Created  string
			Username string `json:"username"`
			Verifed  string `json:"verifed"`
		} `json:"balance"`
		Miners       []byte `json:"miners"`
		Transactions []byte `json:"transactions"`
	}
	Server  string
	Success bool
}

type PoolInfo struct {
	Ip      string
	Name    string
	Port    int32
	Server  string
	Success bool
}

type Stats struct {
	Connections       int     `json:"Active connections"`
	TotalMined        float64 `json:"All-timed mined DUCO"`
	CurrentDifficulty int     `json:"Current difficulty"`
	DUCOHashrate      string
	JustSwap          float32
	NodeS             float32
	PancakeSwap       float32
	SushiSwap         float32
	DUCO              float32 `json:"DUCO price"`
	BCH               float32
	NANO              float32
	TRX               float32
	XMG               float32
	Api               string
	Kolka             struct {
		Banned int
		Jailed int
	}
	LastHash    string
	LastSync    float64
	LastUpdated string
	MinedHashes string
}

type Currency struct {
	Date string
	Usd  struct {
		Ada, Aed, Afm                                                   float32
		Afn, All, Amd, Aoa, Ars, Aud, Awg, Azn                          float32
		Bam, Bbd, Bch, Bdt, Bgn, Bhd, Bif                               float32
		Bmd, Bnb, Bnd, Bob, Brl, Bsd, Btc, Btn, Bwp                     float32
		Byn, Byr, Bzd, Cad, Cdf, Chf, Clf, Clp, Cny                     float32
		Cop, Crc, Cuc, Cup, Cve, Czk, Djf, Dkk, Doge                    float32
		Dop, Dzd, Egp, Ern, Etb, Etc, Eth, Eur, Fjd                     float32
		Fkp, Gbp, Gel, Ggp, Ghs, Gip, Gmd, Gnf, Gtq                     float32
		Gyd, Hkd, Hnl, Hrk, Htg, Huf, Idr, Ils, Imp                     float32
		Inr, Iqd, Irr, Isk, Jep, bJmd, Jod, Jpy, Kes                    float32
		Kgs, Khr, Kmf, Kpw, Krw, Kwd, Kyd, Kzt, Lak                     float32
		Lbp, Link, Lkr, Lrd, Isl, Ltc, Itl, Lvl, Lyd                    float32
		Mad, Mdl, Mga, Mkd, Mmk, Mnt, Mop, Mro, Mur                     float32
		Mvr, Mwk, Mxn, Myr, Mzn, Nad, Ngn, Nio, Nok                     float32
		Npr, Nzd, Omr, Pab, Pen, Pgk, Php, Pkr, Pln, Pyg, Qar, Ron, Rsd float32
		Rub, Rwf, Sar, Sbd, Scr, Sdg, Sek                               float32
		Sgd, Shp, Sll, Sos, Srd, Std, Svc, Syp, Szl, Thb                float32
		Theta, Tjs, Tmt, Tnd, Top, Trx, Try, Ttd, Twd, Tzs              float32
		Uah, Ugx, Usd, Usdt                                             float32
		Uyu, Uzs, Vef, Vnd, Vuv, Wst, Xaf, Xag                          float32
		Xau, Xcd, Xdr, Xlm, Xmr, Xof, Xpf, Xrp                          float32
		Yer, Zar, Zmk, Zmw, Zwl                                         float32
	}
}

type Configuration struct {
	Username   string `json:"username"`
	Currency   string `json:"currency"`
	Difficulty string `json:"diff"`
	Threads    string `json:"threads"`
	FeedEvery  int32  `json:"feed-every"`
}

var acc int
var username string = "ndraey"
var data Balance

var pool PoolInfo
var threads int = 7
var priced Stats

var additconv string = "none"
var currency Currency
var config Configuration

var defconfigfile string = "config.json"
var feedevery time.Duration
var difficulty string = "low"

var balance float32
var oldbalance float32
var version string = "v1.2"

func main() {
	feedevery = time.Duration(45)
	threads = runtime.NumCPU()

	/*
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

	*/

	if _, err := os.Stat(defconfigfile); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Welcome to PulseMiner configuration menu!\n\n")
		fmt.Printf("\n")

		var usr, crn, dif, thr string
		var fev int

		for usr == "" {
			print("Your username: ")
			fmt.Scanf("%s", &usr)
		}
		print("Choose your currency [RUB,UAH,CZK... or none]: ")
		fmt.Scanf("%s", &crn)
		for dif != "low" && dif != "medium" && dif != "high" {
			print("Choose difficulty [low,medium,high]: ")
			fmt.Scanf("%s", &dif)
		}
		print("Threads [or auto]: ")
		fmt.Scanf("%s", &thr)
		print("Print balance every (seconds) [default: 45]: ")
		fmt.Scanf("%d", &fev)

		if fev == 0 {
			fev = 45
		}

		file, err2 := os.Create("config.json")
		defer file.Close()
		if err2 != nil {
			log.Fatal(err)
		}
		err3 := ioutil.WriteFile("config.json",
			[]byte("{\n\t"+
				mkjsonentry("username", usr, 1)+"\n\t"+
				mkjsonentry("currency", crn, 1)+"\n\t"+
				mkjsonentry("diff", dif, 1)+"\n\t"+
				mkjsonentry("threads", thr, 1)+"\n\t"+
				mkjsonentry("feed-every", fev, 0)+"\n"+
				"}\n"),
			0644)
		if err3 != nil {
			log.Fatal(err)
		}
	}

	rawconfig, err := ioutil.ReadFile(defconfigfile)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	json.Unmarshal(rawconfig, &config)

	if config.Username != "" {
		username = config.Username
	}
	if config.Currency != "" {
		if isinarray(getcurrs(), strings.Title(strings.ToLower(config.Currency))) {
			additconv = config.Currency
		} else {
			fmt.Printf("Warning: Currency %s not found in list. Setting to none...\n", config.Currency)
		}
	}
	if config.Threads != "" {
		if config.Threads != "auto" {
			th, err := strconv.Atoi(config.Threads)
			if err != nil {
				log.Fatal(err)
			}
			threads = th
		}
	}
	if config.FeedEvery != 0 {
		feedevery = time.Duration(config.FeedEvery)
	}
	if config.Difficulty != "" {
		difficulty = config.Difficulty
	}

	/*
		bannerw:=47
		unlen:=14+len(username)
		currlen:=21+len(additconv)
		threadslen:=12+len(strconv.Itoa(threads))
		fevlen:=31+len(strconv.Itoa(int(feedevery)))
		dlen:=15+len(difficulty)
		verlen:=26+len(version)
		print("-----------------------------------------------\n")
		print("| Welcome to PulseMiner "+version+strings.Repeat(" ",bannerw-verlen)+" |\n")
		print("|                                             |\n")
		print("| Mining to: "+username+strings.Repeat(" ",bannerw-unlen)+"|\n")
		print("| Currency to show: "+additconv+strings.Repeat(" ",bannerw-currlen)+"|\n")
		print("| Threads: "+strconv.Itoa(threads)+strings.Repeat(" ",bannerw-threadslen)+"|\n")
		print("| Difficulty: "+strings.ToUpper(difficulty)+strings.Repeat(" ",bannerw-dlen)+"|\n")
		print("| Show balance every: "+strconv.Itoa(int(feedevery))+" seconds"+strings.Repeat(" ",bannerw-fevlen)+"|\n")
		print("-----------------------------------------------\n")
	*/

	table := Pradz.PradzTable{}
	table.Init()
	table.AddElement("Welcome to PulseMiner " + version + "!")
	table.AddElement("")
	table.AddElement("Mining to: " + username)
	table.AddElement("Currency to show: " + additconv)
	table.AddElement("Threads: " + strconv.Itoa(threads))
	table.AddElement("Difficulty: " + strings.ToUpper(difficulty))
	table.AddElement("Show balance every: " + strconv.Itoa(int(feedevery)) + " seconds")
	fmt.Printf("%s\n", table.Render())

	print("\nGetting best pool...\n")
	tothashrate := make([]float64, threads)

	for {
		pooldata, err := getnet("https://server.duinocoin.com/getPool")
		if err != nil {
			print("Error getting pool! Press Ctrl-C to exit...\n")
			for {
			}
		}
		json.Unmarshal(pooldata, &pool)
		fmt.Printf("Selected pool %s:%d\n", pool.Ip, pool.Port)
		if pool.Ip == "" {
			fmt.Printf("%sWarning:%s Received empty ip, retrying...\n",
				colorize(93),
				colorize(0))
		} else {
			break
		}
	}

	print("Connecting...\n")
	rand.Seed(time.Now().UnixNano())
	minerid := rand.Int31() % 32768 * 2

	// Main function
	for i := 0; i < threads; i++ {
		go mine(pool.Ip, pool.Port, minerid, i, tothashrate)
		time.Sleep(10 * time.Millisecond)
	}

	print("[1] Preparing some data...\n")
	for {
		info, err := http.Get("https://server.duinocoin.com/users/" + username)
		if err != nil {
			log.Fatal(err)
			time.Sleep(2 * time.Second)
			continue
		}
		body, err := ioutil.ReadAll(info.Body)
		if err != nil {
			log.Fatal(err)
			time.Sleep(2 * time.Second)
			continue
		}
		json.Unmarshal(body, &data)
		oldbalance = data.Result.Balance.Balance
		balance = oldbalance // avoiding report at startup
		break
	}
	print("[2] Preparing some data...\n")
	for {
		info, err := http.Get("https://server.duinocoin.com/users/" + username)
		if err != nil {
			log.Fatal(err)
			time.Sleep(2 * time.Second)
			continue
		}
		body, err := ioutil.ReadAll(info.Body)
		if err != nil {
			log.Fatal(err)
			time.Sleep(2 * time.Second)
			continue
		}
		json.Unmarshal(body, &data)
		balance = data.Result.Balance.Balance

		pricedata, err := getnet("https://server.duinocoin.com/statistics")
		if err != nil {
			log.Fatal(err)
			time.Sleep(2 * time.Second)
		}
		json.Unmarshal(pricedata, &priced)

		if additconv != "none" {
			fmt.Printf("[feed] Balance: %s%.5f%s DUCO (≈%s%.5f%s USD)"+
				currconv(balance*priced.DUCO)+"\n",
				colorize(93),
				balance,
				colorize(0),
				colorize(93),
				balance*priced.DUCO,
				colorize(0))
		} else {
			fmt.Printf("[feed] Balance: %s%.5f%s DUCO (≈%s%.5f%s USD)\n",
				colorize(93),
				balance,
				colorize(0),
				colorize(93),
				balance*priced.DUCO,
				colorize(0))
		}
		if balance-oldbalance > 0 {
			baldiffs := ((balance - oldbalance) / float32(feedevery))
			print(strings.Repeat("=", 65) + "\n")
			fmt.Printf("[report] %s+%.8f%s DUCO\n",
				colorize(92),
				balance-oldbalance,
				colorize(0))
			if additconv != "none" {
				fmt.Printf("[report] Hourly:  %s%.5f%s/hr  (≈%s%.5f%s USD) %s\n",
					colorize(92),
					baldiffs*3600,
					colorize(0),
					colorize(92),
					baldiffs*3600*priced.DUCO,
					colorize(0),
					currconv(baldiffs*3600*priced.DUCO))

				fmt.Printf("[report] Daily:   %s%.5f%s/day (≈%s%.5f%s USD) %s\n",
					colorize(92),
					baldiffs*3600*24,
					colorize(0),
					colorize(92),
					baldiffs*3600*24*priced.DUCO,
					colorize(0),
					currconv(baldiffs*3600*24*priced.DUCO))

				fmt.Printf("[report] Weekly:   %s%.5f%s/wk (≈%s%.5f%s USD) %s\n",
					colorize(92),
					baldiffs*3600*24*7,
					colorize(0),
					colorize(92),
					baldiffs*3600*24*7*priced.DUCO,
					colorize(0),
					currconv(baldiffs*3600*24*7*priced.DUCO))

				// Just average 30.
				fmt.Printf("[report] Monthly: %s%.5f%s/mon (≈%s%.5f%s USD) %s\n",
					colorize(92),
					baldiffs*3600*24*30,
					colorize(0),
					colorize(92),
					baldiffs*3600*24*30*priced.DUCO,
					colorize(0),
					currconv(baldiffs*3600*24*30*priced.DUCO))

			} else {
				fmt.Printf("[report] Hourly:  %s%.5f%s/hr  (≈%s%.5f%s USD)\n",
					colorize(92),
					baldiffs*3600,
					colorize(0),
					colorize(92),
					baldiffs*3600*priced.DUCO,
					colorize(0))

				fmt.Printf("[report] Daily:   %s%.5f%s/day (≈%s%.5f%s USD)\n",
					colorize(92),
					baldiffs*3600*24,
					colorize(0),
					colorize(92),
					baldiffs*3600*24*priced.DUCO,
					colorize(0))

				fmt.Printf("[report] Weekly:   %s%.5f%s/wk (≈%s%.5f%s USD)\n",
					colorize(92),
					baldiffs*3600*24*7,
					colorize(0),
					colorize(92),
					baldiffs*3600*24*7*priced.DUCO,
					colorize(0))

				fmt.Printf("[report] Monthly: %s%.5f%s/mon (≈%s%.5f%s USD)\n",
					colorize(92),
					baldiffs*3600*24*30,
					colorize(0),
					colorize(92),
					baldiffs*3600*24*30*priced.DUCO,
					colorize(0))
			}
			print(strings.Repeat("=", 65) + "\n")
		}
		time.Sleep(feedevery * time.Second)
		oldbalance = balance
	}
}

func getnet(url string) ([]byte, error) {
	info, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return []byte(""), err
	} else {
		body, err := ioutil.ReadAll(info.Body)
		return body, err
	}
}

func sum(array []float64) float64 {
	result := float64(0)
	for _, v := range array {
		result += v
	}
	return result
}

func currconv(value float32) string {
	r := fmt.Sprintf(" (%.5f %s)",
		currconvf(value), strings.ToUpper(additconv))
	return r
}

func currconvf(value float32) float32 {
	st1, err := getnet("https://cdn.jsdelivr.net/gh/fawazahmed0/currency-api@1/latest/currencies/usd.json")
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(st1, &currency)
	r := reflect.ValueOf(currency.Usd)
	res := reflect.Indirect(r).FieldByName(strings.Title(strings.ToLower(additconv)))

	return float32(res.Float()) * value
}

func colorize(color int) string {
	str := ""
	if runtime.GOOS == "windows" {
		str = "\x1b[" + strconv.Itoa(color) + "m"
	} else {
		str = "\033[" + strconv.Itoa(color) + "m"
	}
	return str
}

func mine(ip string, port, minerid int32, cpuid int, thr []float64) {
	fmt.Printf("[cpu%d] Started...\n", cpuid)
	conn, _ := net.Dial("tcp", ip+":"+strconv.Itoa(int(port)))
	ver := make([]byte, 10)
	conn.Read(ver)

	for {
		job_ := string(writejobreq(conn))
		job := strings.Split(job_, ",")
		if len(job) < 3 {
			fmt.Printf("[cpu%d] %sError:%s Thread #%d got empty job! PulseMiner stopped this thread.\n",
				cpuid,
				colorize(91),
				colorize(0),
				cpuid)
			thr[cpuid] = 0
			fmt.Printf("[cpu%d] Trying to reload in 5s...\n",
				cpuid)
			time.Sleep(5 * time.Second)
			mine(ip, port, minerid, cpuid, thr)
			break
		}
		jh := job[0]
		nh := job[1]
		diff := strings.Replace(string(job[2]), "\n", "", -1)
		diff = strings.Replace(string(diff), "\x00", "", -1)
		diffi, err := strconv.ParseInt(diff, 10, 32)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		nonce, hashrate, time := DecodeHash(jh, nh, int(diffi))
		conn.Write([]byte(
			strconv.Itoa(int(nonce)) + "," +
			strconv.Itoa(int(hashrate)) + ",Official PC Miner,pulsemon-go,," +
			strconv.Itoa(int(minerid))))
		
		result := make([]byte, 32)
		conn.Read(result)
		thr[cpuid] = hashrate
		if strings.HasPrefix(string(result), "GOOD") {
			acc++
			fmt.Printf("[%d] [cpu%d] \033[32mGOOD\033[0m H: %sH/s T: %.2fs\n",
				acc,
				cpuid,
				convr(sum(thr)),
				time)
		} else if strings.HasPrefix(string(result), "BAD") {
			fmt.Printf("%s\n", string(result))
			fmt.Printf("[%d] [cpu%d] \033[31;1mBAD\033[0m H: %sH/s T: %.2fs\n",
				acc,
				cpuid,
				convr(sum(thr)),
				time)
		}
	}
}

func gettime() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func RemoveIndex(s []string, i int) []string {
	return s[:i+copy(s[i:], s[i+1:])]
}

func convr(num float64) string {
	rts := ""
	if num > (1000 * 1000 * 1000) {
		k := strconv.FormatFloat(num/(1000*1000*1000), 'f', 2, 64)
		rts = k + "G"
	} else if num > (1000 * 1000) {
		k := strconv.FormatFloat(num/(1000*1000), 'f', 2, 64)
		rts = k + "M"
	} else if num > 1000 {
		k := strconv.FormatFloat(num/1000, 'f', 2, 64)
		rts = k + "k"
	} else {
		rts = strconv.Itoa(int(num))
	}
	return rts
}

func DecodeHash(prev string, result string, diff int) (int, float64, float64) {
	i := int(0)
	hr := float64(0)
	tm := float64(gettime() / 1000)
	tmm := float64(0)
	for i = 0; i < (100*diff + 1); i++ {
		hash := sha1.Sum([]byte(prev + strconv.Itoa(i)))
		hashd := fmt.Sprintf("%02x", hash)
		if hashd == result {
			tmm = float64(gettime()/1000) - tm
			if tmm <= 1 {
				tmm = 1
			}
			hr = float64(i) / tmm
			break
		}
	}
	return i, hr, tmm
}

func writejobreq(con net.Conn) []byte {
	con.Write([]byte("JOB," + username + "," + strings.ToUpper(difficulty)))
	//con.Write([]byte("JOB," + username + ",AVR"))
	req := make([]byte, 90)
	con.Read(req)
	time.Sleep(20 * time.Millisecond) // Not affect on mining speed
	return req
}

func copystr(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}

func mkjsonentry(key string, value interface{}, cl int) string {
	tstr := "\"" + key + "\":"
	switch value.(type) {
	case int32:
		tstr += strconv.Itoa(value.(int))
	case int64:
		tstr += strconv.Itoa(value.(int))
	case int:
		tstr += strconv.Itoa(value.(int))
	case string:
		tstr += "\"" + value.(string) + "\""
	}
	if cl == 1 {
		tstr += ","
	}
	return tstr
}

func getcurrs() []string {
	// Idk how to get all struct keys and made this stuff...
	// It's just a simple parser turns a map with one object to string
	// and finds keys

	exa := map[Currency]string{Currency{}: "Pulsemon-Digimon"}
	raw := fmt.Sprintf("%+v", exa)
	raw = raw[strings.Index(raw, "{")+1:]
	raw = raw[strings.Index(raw, "{")+1:]
	raw = raw[:strings.Index(raw, "}")]
	spl := strings.Split(raw, " ")
	for idx, elm := range spl {
		spl[idx] = strings.Split(elm, ":")[0]
	}
	return spl
}

func isinarray(array []string, value string) bool {
	for _, b := range array {
		if b == value {
			return true
		}
	}
	return false
}
